package editor

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestDebouncerCoalescesCalls(t *testing.T) {
	d := newDebouncer(50 * time.Millisecond)
	var count atomic.Int32
	for i := 0; i < 5; i++ {
		d.Add(func() { count.Add(1) })
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(100 * time.Millisecond)
	if got := count.Load(); got != 1 {
		t.Fatalf("debouncer fired %d times, want 1", got)
	}
}

func TestDebouncerUsesLatestCallback(t *testing.T) {
	d := newDebouncer(50 * time.Millisecond)
	var last atomic.Int32
	d.Add(func() { last.Store(1) })
	d.Add(func() { last.Store(2) })
	d.Add(func() { last.Store(3) })
	time.Sleep(100 * time.Millisecond)
	if got := last.Load(); got != 3 {
		t.Fatalf("last callback = %d, want 3", got)
	}
}
