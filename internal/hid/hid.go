// Package hid abstracts the USB HID custom device that feeds button presses.
// The real device (RadKeys DIY keypad with RP2040-Zero) is read via hidapi
// when CGO is enabled; a MockReader is used for development.
package hid

import "sync"

// Event reports a change in a button's pressed state at (row, col).
type Event struct {
	Row     int
	Col     int
	Pressed bool
}

// Reader polls a USB HID custom device for button events.
type Reader interface {
	Open() error
	Events() <-chan Event
	Close() error
}

// MockReader is an in-process Reader for development without hardware.
type MockReader struct {
	ch     chan Event
	done   chan struct{}
	closed bool
	mu     sync.Mutex
	once   sync.Once
}

// NewMock returns a MockReader.
func NewMock() *MockReader {
	return &MockReader{ch: make(chan Event, 64), done: make(chan struct{})}
}

func (m *MockReader) Open() error          { return nil }
func (m *MockReader) Events() <-chan Event { return m.ch }

func (m *MockReader) Close() error {
	m.mu.Lock()
	m.closed = true
	m.mu.Unlock()
	m.once.Do(func() { close(m.done); close(m.ch) })
	return nil
}

// Put injects an event. Non-blocking: drops if the buffer is full.
// Safe to call before, during, or after Close — the mutex prevents sending
// on a channel that Close is about to close.
func (m *MockReader) Put(e Event) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return
	}
	select {
	case m.ch <- e:
	default:
	}
}
