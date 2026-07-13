//go:build cgo

package hid

import (
	"bytes"
	"errors"
	"log"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/sstallion/go-hid"
)

// fakeHIDDevice is a programmable stand-in for *hid.Device used by diyDevice.loop.
// When the programmed reads are exhausted it returns hid.ErrTimeout (matching
// real device behavior when no report is pending), keeping the loop alive
// until Close signals stop. Write captures every written byte for assertions.
type fakeHIDDevice struct {
	mu          sync.Mutex
	calls       []readCall
	next        int
	closedCalls int
	writes      [][]byte
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

func (f *fakeHIDDevice) Write(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := make([]byte, len(p))
	copy(cp, p)
	f.writes = append(f.writes, cp)
	return len(p), nil
}

func (f *fakeHIDDevice) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closedCalls++
	return nil
}

// written returns a copy of all byte slices passed to Write.
func (f *fakeHIDDevice) written() [][]byte {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([][]byte, len(f.writes))
	copy(out, f.writes)
	return out
}

func TestDIYDeviceReportDelivered(t *testing.T) {
	dev := newFakeHIDDevice([]readCall{
		{report: []byte{2, 5}, n: 2, err: nil},
	})
	r := &diyDevice{deviceBase: newBase(dev)}
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

func TestDIYDeviceReadErrorClosesEvents(t *testing.T) {
	readErr := errors.New("usb disconnect")
	dev := newFakeHIDDevice([]readCall{
		{err: readErr},
	})
	r := &diyDevice{deviceBase: newBase(dev)}
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

func TestDIYDeviceStopClosesEvents(t *testing.T) {
	dev := newFakeHIDDevice(nil)
	r := &diyDevice{deviceBase: newBase(dev)}
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

func TestDIYDeviceCloseIdempotent(t *testing.T) {
	dev := newFakeHIDDevice(nil)
	r := &diyDevice{deviceBase: newBase(dev)}
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
	br := deviceBase{
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

func TestDIYDeviceFirePasteWritesCtrlBytes(t *testing.T) {
	dev := newFakeHIDDevice(nil)
	d := &diyDevice{deviceBase: newBase(dev)}
	if err := d.FirePaste(ModifierCtrl); err != nil {
		t.Fatalf("FirePaste: %v", err)
	}
	writes := dev.written()
	if len(writes) != 1 {
		t.Fatalf("writes len = %d, want 1", len(writes))
	}
	want := []byte{0x00, 0x01, byte(ModifierCtrl)}
	if !bytes.Equal(writes[0], want) {
		t.Fatalf("write = %v, want %v", writes[0], want)
	}
}

func TestDIYDeviceFirePasteWritesGUIBytes(t *testing.T) {
	dev := newFakeHIDDevice(nil)
	d := &diyDevice{deviceBase: newBase(dev)}
	if err := d.FirePaste(ModifierGUI); err != nil {
		t.Fatalf("FirePaste: %v", err)
	}
	writes := dev.written()
	if len(writes) != 1 {
		t.Fatalf("writes len = %d, want 1", len(writes))
	}
	want := []byte{0x00, 0x01, byte(ModifierGUI)}
	if !bytes.Equal(writes[0], want) {
		t.Fatalf("write = %v, want %v", writes[0], want)
	}
}

func TestModifierForOS(t *testing.T) {
	got := ModifierForOS()
	want := ModifierCtrl
	if runtime.GOOS == "darwin" {
		want = ModifierGUI
	}
	if got != want {
		t.Fatalf("ModifierForOS() = %d, want %d", got, want)
	}
}

func TestSelectVendorPathPicksVendorInterface(t *testing.T) {
	infos := []*hid.DeviceInfo{
		{Path: "/dev/hidraw0", UsagePage: 0x0001},
		{Path: "/dev/hidraw1", UsagePage: 0xFF00},
	}
	path, ok := selectVendorPath(infos)
	if !ok {
		t.Fatal("expected ok=true for vendor interface present")
	}
	if path != "/dev/hidraw1" {
		t.Fatalf("path = %q, want /dev/hidraw1", path)
	}
}

func TestSelectVendorPathFallsBackWhenNone(t *testing.T) {
	infos := []*hid.DeviceInfo{
		{Path: "/dev/hidraw0", UsagePage: 0x0001},
	}
	_, ok := selectVendorPath(infos)
	if ok {
		t.Fatal("expected ok=false when no vendor interface")
	}
}

func TestSelectVendorPathEmpty(t *testing.T) {
	_, ok := selectVendorPath(nil)
	if ok {
		t.Fatal("expected ok=false for empty infos")
	}
}
