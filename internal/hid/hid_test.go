package hid

import (
	"sync"
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

// TestMockReaderPutConcurrentCloseNoPanic is a regression test for the
// send-on-closed-channel race: Put used to unlock before selecting on m.ch,
// so a concurrent Close could close m.ch and panic the send. Run with
// `go test -race` to also catch the data race.
func TestMockReaderPutConcurrentCloseNoPanic(t *testing.T) {
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
