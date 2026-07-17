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
	onWrite     func(p []byte) []byte // optional: returns a read reply to queue
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
	if f.onWrite != nil {
		if reply := f.onWrite(cp); reply != nil {
			f.calls = append(f.calls, readCall{report: reply, n: len(reply)})
		}
	}
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
	r := newDIYDevice(dev)
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
	r := newDIYDevice(dev)
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
	r := newDIYDevice(dev)
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
	r := newDIYDevice(dev)
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
	br := diyDevice{
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

func TestDIYDeviceFireCommandWritesCtrlBytes(t *testing.T) {
	dev := newFakeHIDDevice(nil)
	d := newDIYDevice(dev)
	if err := d.FireCommand(CmdFirePaste, byte(ModifierCtrl)); err != nil {
		t.Fatalf("FireCommand: %v", err)
	}
	writes := dev.written()
	if len(writes) != 1 {
		t.Fatalf("writes len = %d, want 1", len(writes))
	}
	want := []byte{0x00, byte(CmdFirePaste), byte(ModifierCtrl)}
	if !bytes.Equal(writes[0], want) {
		t.Fatalf("write = %v, want %v", writes[0], want)
	}
}

func TestDIYDeviceFireCommandWritesGUIBytes(t *testing.T) {
	dev := newFakeHIDDevice(nil)
	d := newDIYDevice(dev)
	if err := d.FireCommand(CmdFirePaste, byte(ModifierGUI)); err != nil {
		t.Fatalf("FireCommand: %v", err)
	}
	writes := dev.written()
	if len(writes) != 1 {
		t.Fatalf("writes len = %d, want 1", len(writes))
	}
	want := []byte{0x00, byte(CmdFirePaste), byte(ModifierGUI)}
	if !bytes.Equal(writes[0], want) {
		t.Fatalf("write = %v, want %v", writes[0], want)
	}
}

func TestDIYDeviceFireCommandSelectAllGUI(t *testing.T) {
	dev := newFakeHIDDevice(nil)
	d := newDIYDevice(dev)
	if err := d.FireCommand(CmdSelectAll, byte(ModifierGUI)); err != nil {
		t.Fatalf("FireCommand: %v", err)
	}
	writes := dev.written()
	if len(writes) != 1 {
		t.Fatalf("writes len = %d, want 1", len(writes))
	}
	want := []byte{0x00, byte(CmdSelectAll), byte(ModifierGUI)}
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

func TestReadFirmwareVersionReply(t *testing.T) {
	dev := newFakeHIDDevice(nil)
	dev.onWrite = func(p []byte) []byte {
		if len(p) >= 3 && p[0] == reportIDNone && p[1] == cmdGetVersion && p[2] == 0x00 {
			return []byte{1, 0} // simulate firmware v1.0
		}
		return nil
	}
	major, minor, known := readFirmwareVersion(dev)
	if !known || major != 1 || minor != 0 {
		t.Fatalf("readFirmwareVersion = (%d, %d, %v), want (1, 0, true)", major, minor, known)
	}
	writes := dev.written()
	if len(writes) != 1 || !bytes.Equal(writes[0], []byte{reportIDNone, cmdGetVersion, 0x00}) {
		t.Fatalf("writes = %v, want [[0x00 0x02 0x00]]", writes)
	}
}

func TestReadFirmwareVersionTimeout(t *testing.T) {
	dev := newFakeHIDDevice(nil)
	// No onWrite: the device never replies; ReadWithTimeout returns ErrTimeout.
	major, minor, known := readFirmwareVersion(dev)
	if known {
		t.Fatalf("readFirmwareVersion = (%d, %d, %v), want unknown", major, minor, known)
	}
	writes := dev.written()
	if len(writes) != 1 || !bytes.Equal(writes[0], []byte{reportIDNone, cmdGetVersion, 0x00}) {
		t.Fatalf("writes = %v, want [[0x00 0x02 0x00]]", writes)
	}
}

func TestReadFirmwareVersionRetriesOnImplausibleReply(t *testing.T) {
	// First two replies look like plausible button events (row=0, col=1 and
	// row=5, col=5); the third is the real version [1, 0]. The retry logic
	// should reject the first two because major/minor fall inside button
	// ranges but are below our loose bound too — wait, row/col 5 is below
	// versionPlausibleMajor=10. We need implausible bytes, so use [255, 255]
	// and [42, 0] as noise, then [1, 0].
	dev := newFakeHIDDevice(nil)
	call := 0
	dev.onWrite = func(p []byte) []byte {
		if len(p) < 3 || p[0] != reportIDNone || p[1] != cmdGetVersion || p[2] != 0x00 {
			return nil
		}
		call++
		switch call {
		case 1:
			return []byte{0xff, 0xff} // implausible
		case 2:
			return []byte{42, 0} // implausible major
		default:
			return []byte{1, 0} // real version
		}
	}
	major, minor, known := readFirmwareVersion(dev)
	if !known || major != 1 || minor != 0 {
		t.Fatalf("readFirmwareVersion = (%d, %d, %v), want (1, 0, true)", major, minor, known)
	}
	if call < 3 {
		t.Fatalf("expected at least 3 GET_VERSION writes, got %d", call)
	}
}

func TestReadFirmwareVersionAcceptsFirstPlausibleReply(t *testing.T) {
	dev := newFakeHIDDevice(nil)
	dev.onWrite = func(p []byte) []byte {
		if len(p) >= 3 && p[0] == reportIDNone && p[1] == cmdGetVersion && p[2] == 0x00 {
			return []byte{1, 0} // simulate firmware v1.0
		}
		return nil
	}
	major, minor, known := readFirmwareVersion(dev)
	if !known || major != 1 || minor != 0 {
		t.Fatalf("readFirmwareVersion = (%d, %d, %v), want (1, 0, true)", major, minor, known)
	}
	writes := dev.written()
	if len(writes) != 1 || !bytes.Equal(writes[0], []byte{reportIDNone, cmdGetVersion, 0x00}) {
		t.Fatalf("writes = %v, want [[0x00 0x02 0x00]]", writes)
	}
}

func TestDIYDeviceVersionKnown(t *testing.T) {
	dev := newFakeHIDDevice(nil)
	d := &diyDevice{
		dev:          dev,
		ch:           make(chan Event, 64),
		stop:         make(chan struct{}),
		done:         make(chan struct{}),
		versionMajor: 1,
		versionMinor: 0,
		versionKnown: true,
	}
	maj, min, err := d.Version()
	if err != nil || maj != 1 || min != 0 {
		t.Fatalf("Version() = (%d, %d, %v), want (1, 0, nil)", maj, min, err)
	}
}

func TestDIYDeviceVersionUnknown(t *testing.T) {
	dev := newFakeHIDDevice(nil)
	d := newDIYDevice(dev)
	_, _, err := d.Version()
	if err == nil {
		t.Fatal("Version() should return error when version is unknown")
	}
}

func TestIsPlausibleVersion(t *testing.T) {
	tests := []struct {
		major byte
		minor byte
		want  bool
	}{
		{1, 0, true},
		{5, 99, true},
		{9, 99, true},
		{10, 0, false},
		{0, 100, false},
		{255, 255, false},
	}
	for _, tc := range tests {
		got := isPlausibleVersion(tc.major, tc.minor)
		if got != tc.want {
			t.Fatalf("isPlausibleVersion(%d, %d) = %v, want %v", tc.major, tc.minor, got, tc.want)
		}
	}
}
