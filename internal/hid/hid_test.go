package hid

import (
	"sync"
	"testing"
	"time"
)

func TestMockDeviceOpenClose(t *testing.T) {
	m := NewMock()
	if err := m.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := m.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestMockDevicePutAndReceive(t *testing.T) {
	m := NewMock()
	_ = m.Open()
	defer func() { _ = m.Close() }()

	m.Put(Event{Row: 0, Col: 3, Pressed: true})

	select {
	case ev := <-m.Events():
		if ev.Row != 0 || ev.Col != 3 || !ev.Pressed {
			t.Fatalf("got %+v, want row=0 col=3 pressed=true", ev)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestMockDevicePutAfterCloseIsSafe(t *testing.T) {
	m := NewMock()
	_ = m.Open()
	_ = m.Close()
	// Put after close must not panic.
	m.Put(Event{Row: 0, Col: 0, Pressed: true})
}

func TestMockDeviceEventsClosedAfterClose(t *testing.T) {
	m := NewMock()
	_ = m.Open()
	_ = m.Close()
	// Events channel should be closed.
	_, ok := <-m.Events()
	if ok {
		t.Fatal("Events channel should be closed after Close")
	}
}

func TestMockDeviceFirePasteRecordsCalls(t *testing.T) {
	m := NewMock()
	_ = m.Open()
	defer func() { _ = m.Close() }()

	if err := m.FirePaste(ModifierCtrl); err != nil {
		t.Fatalf("FirePaste Ctrl: %v", err)
	}
	if err := m.FirePaste(ModifierGUI); err != nil {
		t.Fatalf("FirePaste GUI: %v", err)
	}

	calls := m.PasteCalls()
	if len(calls) != 2 {
		t.Fatalf("PasteCalls len = %d, want 2", len(calls))
	}
	if calls[0] != ModifierCtrl || calls[1] != ModifierGUI {
		t.Fatalf("PasteCalls = %v, want [Ctrl, GUI]", calls)
	}
}

func TestMockDevicePasteCallsReturnsCopy(t *testing.T) {
	m := NewMock()
	_ = m.FirePaste(ModifierCtrl)
	calls := m.PasteCalls()
	calls[0] = ModifierGUI
	if m.PasteCalls()[0] != ModifierCtrl {
		t.Fatal("PasteCalls should return a copy, not the internal slice")
	}
}

func TestMockDeviceCloseIdempotent(t *testing.T) {
	m := NewMock()
	_ = m.Open()
	if err := m.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := m.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

// TestMockDevicePutConcurrentCloseNoPanic is a regression test for the
// send-on-closed-channel race: Put used to unlock before selecting on m.ch,
// so a concurrent Close could close m.ch and panic the send. Run with
// `go test -race` to also catch the data race.
func TestMockDevicePutConcurrentCloseNoPanic(t *testing.T) {
	for i := 0; i < 5000; i++ {
		m := NewMock()
		_ = m.Open()
		start := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); <-start; m.Put(Event{Row: 0, Col: 0, Pressed: true}) }()
		go func() { defer wg.Done(); <-start; _ = m.Close() }()
		close(start)
		wg.Wait()
	}
}

// TestMockDeviceFirePasteConcurrentCloseNoPanic ensures FirePaste and Close
// can run concurrently without panicking or racing on the pastes slice.
func TestMockDeviceFirePasteConcurrentCloseNoPanic(t *testing.T) {
	for i := 0; i < 5000; i++ {
		m := NewMock()
		_ = m.Open()
		start := make(chan struct{})
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); <-start; _ = m.FirePaste(ModifierCtrl) }()
		go func() { defer wg.Done(); <-start; _ = m.Close() }()
		close(start)
		wg.Wait()
	}
}

func TestFirmwareOutdated(t *testing.T) {
	tests := []struct {
		name  string
		major byte
		minor byte
		known bool
		want  bool
	}{
		{"unknown", 0, 0, false, true},
		{"v0.9 known", 0, 9, true, true},
		{"v1.0 known", 1, 0, true, false},
		{"v1.1 known", 1, 1, true, false},
		{"v2.0 known", 2, 0, true, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := FirmwareOutdated(tc.major, tc.minor, tc.known); got != tc.want {
				t.Fatalf("FirmwareOutdated(%d, %d, %v) = %v, want %v",
					tc.major, tc.minor, tc.known, got, tc.want)
			}
		})
	}
}

func TestMockDeviceVersion(t *testing.T) {
	m := NewMock()
	maj, min, err := m.Version()
	if err != nil || maj != 1 || min != 0 {
		t.Fatalf("Version() = (%d, %d, %v), want (1, 0, nil)", maj, min, err)
	}
}

func TestMockDeviceSetFirmwareVersion(t *testing.T) {
	m := NewMock()
	m.SetFirmwareVersion(2, 3)
	maj, min, err := m.Version()
	if err != nil || maj != 2 || min != 3 {
		t.Fatalf("Version() = (%d, %d, %v), want (2, 3, nil)", maj, min, err)
	}

	m.SetFirmwareVersion(0, 0)
	_, _, err = m.Version()
	if err == nil {
		t.Fatal("Version() should return error when major=0 (unknown)")
	}
}
