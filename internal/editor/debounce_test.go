package editor

import (
	"testing"
	"time"
)

func TestDebouncerCoalescesCalls(t *testing.T) {
	d := newDebouncer(50 * time.Millisecond)
	var count int
	for i := 0; i < 5; i++ {
		d.Add(func() { count++ })
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(100 * time.Millisecond)
	if count != 1 {
		t.Fatalf("debouncer fired %d times, want 1", count)
	}
}

func TestDebouncerUsesLatestCallback(t *testing.T) {
	d := newDebouncer(50 * time.Millisecond)
	var last int
	d.Add(func() { last = 1 })
	d.Add(func() { last = 2 })
	d.Add(func() { last = 3 })
	time.Sleep(100 * time.Millisecond)
	if last != 3 {
		t.Fatalf("last callback = %d, want 3", last)
	}
}
