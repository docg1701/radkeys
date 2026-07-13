//go:build cgo

package hid

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sstallion/go-hid"
)

// fakeHIDDevice is a programmable stand-in for *hid.Device used by diyReader.loop.
// When the programmed calls are exhausted it returns hid.ErrTimeout (matching
// real device behavior when no report is pending), keeping the loop alive
// until Close signals stop.
type fakeHIDDevice struct {
	mu          sync.Mutex
	calls       []readCall
	next        int
	closedCalls int
}

type readCall struct {
	report []byte
	n      int
	err    error
}

func newFakeHIDDevice(calls []readCall) *fakeHIDDevice {
	return &fakeHIDDevice{calls: calls}
}

func (f *fakeHIDDevice) ReadWithTimeout(p []byte, _ time.Duration) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.next >= len(f.calls) {
		return 0, hid.ErrTimeout
	}
	call := f.calls[f.next]
	f.next++
	copy(p, call.report)
	return call.n, call.err
}

func (f *fakeHIDDevice) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closedCalls++
	return nil
}

func TestDIYReaderReportDelivered(t *testing.T) {
	dev := newFakeHIDDevice([]readCall{
		{report: []byte{2, 5}, n: 2, err: nil},
	})
	r := &diyReader{baseReader: newBase(dev)}
	if err := r.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}

	select {
	case ev := <-r.Events():
		if ev.Row != 2 || ev.Col != 5 || !ev.Pressed {
			t.Fatalf("got %+v, want row=2 col=5 pressed=true", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}

	if err := r.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if dev.closedCalls != 1 {
		t.Fatalf("fake device Close called %d times, want 1", dev.closedCalls)
	}
}

func TestDIYReaderReadErrorClosesEvents(t *testing.T) {
	readErr := errors.New("usb disconnect")
	dev := newFakeHIDDevice([]readCall{
		{err: readErr},
	})
	r := &diyReader{baseReader: newBase(dev)}
	_ = r.Open()

	select {
	case _, ok := <-r.Events():
		if ok {
			t.Fatal("expected channel to be closed, got an event")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for Events() to close")
	}

	// Close must still be idempotent after loop has exited.
	if err := r.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestDIYReaderStopClosesEvents(t *testing.T) {
	dev := newFakeHIDDevice(nil)
	r := &diyReader{baseReader: newBase(dev)}
	_ = r.Open()

	if err := r.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	select {
	case _, ok := <-r.Events():
		if ok {
			t.Fatal("expected channel to be closed, got an event")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for Events() to close after Close")
	}

	if dev.closedCalls != 1 {
		t.Fatalf("fake device Close called %d times, want 1", dev.closedCalls)
	}
}

func TestDIYReaderCloseIdempotent(t *testing.T) {
	dev := newFakeHIDDevice(nil)
	r := &diyReader{baseReader: newBase(dev)}
	_ = r.Open()
	_ = r.Close()
	if err := r.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
	if dev.closedCalls != 1 {
		t.Fatalf("Close called %d times, want 1", dev.closedCalls)
	}
}

func TestEmitLogsWhenChannelFull(t *testing.T) {
	br := baseReader{
		ch:   make(chan Event, 1),
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}

	var buf bytes.Buffer
	old := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(old)

	br.emit(Event{Row: 0, Col: 0, Pressed: true})
	if got := len(br.ch); got != 1 {
		t.Fatalf("first emit should fill buffer, got %d buffered events", got)
	}

	br.emit(Event{Row: 1, Col: 2, Pressed: true}) // dropped
	if got := len(br.ch); got != 1 {
		t.Fatalf("second emit should be dropped, buffer still %d", got)
	}

	msg := buf.String()
	if !strings.Contains(msg, "hid event dropped") {
		t.Fatalf("expected log about dropped event, got %q", msg)
	}
}
