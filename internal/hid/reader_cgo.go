//go:build cgo

package hid

import (
	"fmt"
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

// baseReader carries the hidapi handle and the event/stop channels.
type baseReader struct {
	dev  *hid.Device
	ch   chan Event
	stop chan struct{}
	done chan struct{}
}

func newBase(dev *hid.Device) baseReader {
	return baseReader{dev: dev, ch: make(chan Event, 64), stop: make(chan struct{}), done: make(chan struct{})}
}

func (b *baseReader) Events() <-chan Event { return b.ch }

func (b *baseReader) Close() error {
	close(b.stop)
	<-b.done
	return b.dev.Close()
}

// emit sends an event. Non-blocking — drops if channel is full (shouldn't happen).
func (b *baseReader) emit(e Event) {
	select {
	case b.ch <- e:
	default:
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
			return
		}
		if n >= diyReportLen {
			d.emit(Event{Row: int(buf[0]), Col: int(buf[1]), Pressed: true})
		}
	}
}
