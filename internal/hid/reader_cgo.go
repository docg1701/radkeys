//go:build cgo

package hid

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/sstallion/go-hid"

	"github.com/docg1701/radkeys/internal/config"
)

const vendorUsagePage = 0xFF00

// Open connects to the configured USB HID device and returns a Device.
// Requires CGO + hidapi prerequisites (libudev on Linux).
func Open(dev config.Device) (Device, error) {
	if err := hid.Init(); err != nil {
		return nil, fmt.Errorf("hid: init: %w", err)
	}
	d, err := openVendorInterface(dev.VendorID, dev.ProductID)
	if err != nil {
		return nil, fmt.Errorf("hid: open %04x:%04x: %w", dev.VendorID, dev.ProductID, err)
	}
	switch dev.Protocol {
	case config.ProtocolDIY:
		return &diyDevice{deviceBase: newBase(d)}, nil
	default:
		_ = d.Close()
		return nil, fmt.Errorf("hid: unsupported protocol %q", dev.Protocol)
	}
}

// openVendorInterface opens the vendor HID interface (usage page 0xFF00)
// for the device. When no vendor interface is enumerated (e.g. the current
// vendor-only firmware, or empty enumeration) it falls back to OpenFirst so
// today's behavior is preserved.
func openVendorInterface(vendorID, productID uint16) (*hid.Device, error) {
	infos, err := enumerateDevices(vendorID, productID)
	if err != nil {
		return nil, err
	}
	if path, ok := selectVendorPath(infos); ok {
		return hid.OpenPath(path)
	}
	return hid.OpenFirst(vendorID, productID)
}

// enumerateDevices collects all DeviceInfo entries matching the vendor and
// product ID via hid.Enumerate.
func enumerateDevices(vendorID, productID uint16) ([]*hid.DeviceInfo, error) {
	var infos []*hid.DeviceInfo
	err := hid.Enumerate(vendorID, productID, func(info *hid.DeviceInfo) error {
		infos = append(infos, info)
		return nil
	})
	return infos, err
}

// selectVendorPath returns the path of the first device info whose usage page
// is the vendor page 0xFF00, or ok=false when none matches.
func selectVendorPath(infos []*hid.DeviceInfo) (path string, ok bool) {
	for _, info := range infos {
		if info.UsagePage == vendorUsagePage {
			return info.Path, true
		}
	}
	return "", false
}

const pollTimeout = 50 * time.Millisecond

// hidDevice is the minimal surface of *hid.Device that diyDevice uses.
// It exists only to allow loop() and FirePaste() to be exercised with a
// fake device in tests.
type hidDevice interface {
	ReadWithTimeout(p []byte, timeout time.Duration) (int, error)
	Write(p []byte) (int, error)
	Close() error
}

// deviceBase carries the hidapi handle and the event/stop channels.
type deviceBase struct {
	dev         hidDevice
	ch          chan Event
	stop        chan struct{}
	done        chan struct{}
	once        sync.Once
	dropCount   int
	lastDropLog time.Time
}

func newBase(dev hidDevice) deviceBase {
	return deviceBase{dev: dev, ch: make(chan Event, 64), stop: make(chan struct{}), done: make(chan struct{})}
}

func (b *deviceBase) Events() <-chan Event { return b.ch }

// Close is idempotent: a second call returns nil without re-closing the
// device or panicking on a closed channel. The event channel is closed by
// loop() when it exits (stop or read error), so pollHID stops cleanly.
func (b *deviceBase) Close() error {
	var err error
	b.once.Do(func() {
		close(b.stop)
		<-b.done
		err = b.dev.Close()
	})
	return err
}

// emit sends an event. Non-blocking: if the consumer is behind and the
// channel is full, the event is dropped and counted for logging. Blocking
// the HID read loop is not safe because a stuck UI would freeze the device.
func (b *deviceBase) emit(e Event) {
	select {
	case b.ch <- e:
	default:
		b.dropCount++
		if b.lastDropLog.IsZero() || time.Since(b.lastDropLog) > 5*time.Second {
			log.Printf("radkeys: hid event dropped: channel full; %d event(s) lost since last log", b.dropCount)
			b.dropCount = 0
			b.lastDropLog = time.Now()
		}
	}
}

const (
	diyReportLen = 2
	cmdFirePaste = 0x01
	reportIDNone = 0x00
)

// diyDevice implements the RadKeys DIY protocol: input report = 2 bytes
// [row, col] via HID vendor-defined (TinyUSB on RP2040-Zero), and writes the
// fire-paste vendor OUT command.
type diyDevice struct {
	deviceBase
}

func (d *diyDevice) Open() error {
	go d.loop()
	return nil
}

// FirePaste writes the vendor OUT command [reportID, CMD_FIRE_PASTE, modifier]
// to the device. The report ID byte (0x00) is consumed by HIDAPI/TinyUSB and
// not passed to the device callback; the device receives [cmd, arg].
func (d *diyDevice) FirePaste(mod Modifier) error {
	out := []byte{reportIDNone, cmdFirePaste, byte(mod)}
	if _, err := d.dev.Write(out); err != nil {
		return fmt.Errorf("hid: fire paste: %w", err)
	}
	return nil
}

func (d *diyDevice) loop() {
	defer close(d.done)
	defer close(d.ch) // signals pollHID to stop on stop OR read error
	buf := make([]byte, diyReportLen)
	for {
		select {
		case <-d.stop:
			return
		default:
		}
		n, err := d.dev.ReadWithTimeout(buf, pollTimeout)
		if err != nil {
			if err == hid.ErrTimeout {
				continue
			}
			log.Printf("radkeys: hid read failed: %v", err)
			return
		}
		if n >= diyReportLen {
			d.emit(Event{Row: int(buf[0]), Col: int(buf[1]), Pressed: true})
		}
	}
}
