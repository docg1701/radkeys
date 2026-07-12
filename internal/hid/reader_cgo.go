//go:build cgo

package hid

import (
	"fmt"
	"time"

	"github.com/docg1701/radkeys/internal/config"
	"github.com/sstallion/go-hid"
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
	case config.ProtocolElgato:
		return &elgatoReader{baseReader: newBase(d), prev: make([]byte, 64)}, nil
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

// elgatoReader implements the Elgato Stream Deck input protocol.
// Input report: Report ID 0x01, Command 0x00, payload = 1 byte per button
// (0x00 released, 0x01 pressed). Each button index maps to Col; Row is always 0.
type elgatoReader struct {
	baseReader
	prev []byte
}

func (e *elgatoReader) Open() error {
	go e.loop()
	return nil
}

func (e *elgatoReader) loop() {
	defer close(e.done)
	buf := make([]byte, 512)
	for {
		select {
		case <-e.stop:
			return
		default:
		}
		n, err := e.dev.ReadWithTimeout(buf, pollTimeout)
		if err != nil {
			if err == hid.ErrTimeout {
				continue
			}
			return
		}
		if n < 4 || buf[0] != 0x01 || buf[1] != 0x00 {
			continue
		}
		payloadLen := int(buf[2]) | int(buf[3])<<8
		if n < 4+payloadLen {
			continue
		}
		states := buf[4 : 4+payloadLen]
		for i, st := range states {
			if i >= len(e.prev) {
				break
			}
			if st == 0x01 && e.prev[i] == 0 {
				e.emit(Event{Row: 0, Col: i, Pressed: true})
			}
			e.prev[i] = st
		}
	}
}

// ---------------------------------------------------------------------------
// DIY reader: protocolo (row, col) — 2 bytes
// ---------------------------------------------------------------------------

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
