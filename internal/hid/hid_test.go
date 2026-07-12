package hid

import (
	"testing"
	"time"
)

func TestMockReaderOpenClose(t *testing.T) {
	m := NewMock()
	if err := m.Open(); err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := m.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestMockReaderPutAndReceive(t *testing.T) {
	m := NewMock()
	_ = m.Open()
	defer m.Close()

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

func TestMockReaderPutAfterCloseIsSafe(t *testing.T) {
	m := NewMock()
	_ = m.Open()
	_ = m.Close()
	// Put after close must not panic.
	m.Put(Event{Row: 0, Col: 0, Pressed: true})
}

func TestMockReaderEventsClosedAfterClose(t *testing.T) {
	m := NewMock()
	_ = m.Open()
	_ = m.Close()
	// Events channel should be closed.
	_, ok := <-m.Events()
	if ok {
		t.Fatal("Events channel should be closed after Close")
	}
}
