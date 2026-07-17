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
		dd := newDIYDevice(d)
		dd.versionMajor, dd.versionMinor, dd.versionKnown = readFirmwareVersion(d)
		return dd, nil
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
// It exists only to allow loop() and FireCommand() to be exercised with a
// fake device in tests.
type hidDevice interface {
	ReadWithTimeout(p []byte, timeout time.Duration) (int, error)
	Write(p []byte) (int, error)
	Close() error
}

func (d *diyDevice) Events() <-chan Event { return d.ch }

// Close is idempotent: a second call returns nil without re-closing the
// device or panicking on a closed channel. The event channel is closed by
// loop() when it exits (stop or read error), so pollHID stops cleanly.
func (d *diyDevice) Close() error {
	var err error
	d.once.Do(func() {
		close(d.stop)
		<-d.done
		err = d.dev.Close()
	})
	return err
}

// emit sends an event. Non-blocking: if the consumer is behind and the
// channel is full, the event is dropped and counted for logging. Blocking
// the HID read loop is not safe because a stuck UI would freeze the device.
func (d *diyDevice) emit(e Event) {
	select {
	case d.ch <- e:
	default:
		d.dropCount++
		if d.lastDropLog.IsZero() || time.Since(d.lastDropLog) > 5*time.Second {
			log.Printf("radkeys: hid event dropped: channel full; %d event(s) lost since last log", d.dropCount)
			d.dropCount = 0
			d.lastDropLog = time.Now()
		}
	}
}

const (
	diyReportLen          = 2
	cmdGetVersion         = 0x02 // GET_VERSION: host->device, expects a 2-byte version IN reply (not a keyboard command)
	reportIDNone          = 0x00
	versionReadTimeout    = 500 * time.Millisecond
	versionReadAttempts   = 3
	versionPlausibleMajor = 10 // reject replies with major >= 10 as likely button events (rows/cols are 0..5)
	versionPlausibleMinor = 100
)

// diyDevice implements the RadKeys DIY protocol: input report = 2 bytes
// [row, col] via HID vendor-defined (TinyUSB on RP2040-Zero), and writes the
// fire-paste vendor OUT command.
type diyDevice struct {
	dev         hidDevice
	ch          chan Event
	stop        chan struct{}
	done        chan struct{}
	once        sync.Once
	dropCount   int
	lastDropLog time.Time

	versionMajor byte
	versionMinor byte
	versionKnown bool
}

func newDIYDevice(dev hidDevice) *diyDevice {
	return &diyDevice{
		dev:  dev,
		ch:   make(chan Event, 64),
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
}

func (d *diyDevice) Open() error {
	go d.loop()
	return nil
}

// Version returns the firmware version read once at connect. Returns
// errFirmwareVersionUnknown when the device did not respond to GET_VERSION.
func (d *diyDevice) Version() (byte, byte, error) {
	if !d.versionKnown {
		return 0, 0, errFirmwareVersionUnknown
	}
	return d.versionMajor, d.versionMinor, nil
}

// FireCommand writes the vendor OUT command [reportID, cmd, arg] to the
// device. The report ID byte (0x00) is consumed by HIDAPI/TinyUSB and not
// passed to the device callback; the device receives [cmd, arg].
func (d *diyDevice) FireCommand(cmd Command, arg byte) error {
	out := []byte{reportIDNone, byte(cmd), arg}
	if _, err := d.dev.Write(out); err != nil {
		return fmt.Errorf("hid: fire command 0x%02x: %w", byte(cmd), err)
	}
	return nil
}

// readFirmwareVersion sends CMD_GET_VERSION to the device and reads the
// 2-byte [major, minor] response. It retries a few times if the first reply
// looks like a stray button event (row, col). Returns known=false on any
// error or timeout. Must be called before the event loop starts (no concurrent
// reader).
func readFirmwareVersion(dev hidDevice) (major, minor byte, known bool) {
	out := []byte{reportIDNone, cmdGetVersion, 0x00}
	var last [2]byte
	for attempt := 0; attempt < versionReadAttempts; attempt++ {
		if _, err := dev.Write(out); err != nil {
			return 0, 0, false
		}
		buf := make([]byte, 2)
		n, err := dev.ReadWithTimeout(buf, versionReadTimeout)
		if err != nil || n < 2 {
			return 0, 0, false
		}
		last[0], last[1] = buf[0], buf[1]
		if isPlausibleVersion(buf[0], buf[1]) {
			return buf[0], buf[1], true
		}
	}
	return last[0], last[1], true
}

// isPlausibleVersion rejects replies that are clearly not a version reply.
// Button events carry row and col in [0, 5], which overlaps with small
// version numbers, so the heuristic uses a loose upper bound: major < 10 and
// minor < 100. This filters high-value bytes (e.g. 0xFF from a noisy line)
// while accepting current and near-future firmware versions.
func isPlausibleVersion(major, minor byte) bool {
	return major < versionPlausibleMajor && minor < versionPlausibleMinor
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
