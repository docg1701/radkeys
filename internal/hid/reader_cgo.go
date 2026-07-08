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

// emitRise forwards a press only on the 0->1 transition (debounced).
func (b *baseReader) emitRise(i int, pressed bool, prev byte) {
	if pressed && prev == 0 {
		select {
		case b.ch <- Event{Index: i, Pressed: true}:
		default:
		}
	}
}

// elgatoReader implements the Elgato Stream Deck input protocol.
// Input report: Report ID 0x01, Command 0x00, payload = 1 byte per button
// (0x00 released, 0x01 pressed). Host polls with a timed HID READ.
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
			e.emitRise(i, st == 0x01, e.prev[i])
			e.prev[i] = st
		}
	}
}

// diyButtonCount is the number of buttons on the DIY RadKeys device.
const diyButtonCount = 24

// diyReader implements the RadKeys DIY protocol: input report = N bytes,
// one per button (0x00 released, 0x01 pressed), N = diyButtonCount.
type diyReader struct {
	baseReader
	prev [diyButtonCount]byte
}

func (d *diyReader) Open() error {
	go d.loop()
	return nil
}

func (d *diyReader) loop() {
	defer close(d.done)
	buf := make([]byte, diyButtonCount)
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
		for i := 0; i < n && i < diyButtonCount; i++ {
			d.emitRise(i, buf[i] == 0x01, d.prev[i])
			d.prev[i] = buf[i]
		}
	}
}
