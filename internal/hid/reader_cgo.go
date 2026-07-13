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

// Open connects to the configured USB HID custom device and returns a Reader.
// Requires CGO + hidapi prerequisites (libudev on Linux).
func Open(dev config.Device) (Reader, error) {
	if err := hid.Init(); err != nil {
		return nil, fmt.Errorf("hid: init: %w", err)
	}
	d, err := hid.OpenFirst(dev.VendorID, dev.ProductID)
	if err != nil {
		return nil, fmt.Errorf("hid: open %04x:%04x: %w", dev.VendorID, dev.ProductID, err)
	}
	switch dev.Protocol {
	case config.ProtocolDIY:
		return &diyReader{baseReader: newBase(d)}, nil
	default:
		_ = d.Close()
		return nil, fmt.Errorf("hid: unsupported protocol %q", dev.Protocol)
	}
}

const pollTimeout = 50 * time.Millisecond

// hidDevice is the minimal surface of *hid.Device that the reader uses.
// It exists only to allow loop() to be exercised with a fake device in tests.
type hidDevice interface {
	ReadWithTimeout(p []byte, timeout time.Duration) (int, error)
	Close() error
}

// baseReader carries the hidapi handle and the event/stop channels.
type baseReader struct {
	dev         hidDevice
	ch          chan Event
	stop        chan struct{}
	done        chan struct{}
	once        sync.Once
	dropCount   int
	lastDropLog time.Time
}

func newBase(dev hidDevice) baseReader {
	return baseReader{dev: dev, ch: make(chan Event, 64), stop: make(chan struct{}), done: make(chan struct{})}
}

func (b *baseReader) Events() <-chan Event { return b.ch }

// Close is idempotent: a second call returns nil without re-closing the
// device or panicking on a closed channel. The event channel is closed by
// loop() when it exits (stop or read error), so pollHID stops cleanly.
func (b *baseReader) Close() error {
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
func (b *baseReader) emit(e Event) {
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

const diyReportLen = 2

// diyReader implements the RadKeys DIY protocol: input report = 2 bytes
// [row, col] via HID vendor-defined (TinyUSB on RP2040-Zero).
type diyReader struct {
	baseReader
}

func (d *diyReader) Open() error {
	go d.loop()
	return nil
}

func (d *diyReader) loop() {
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
