// Package hid abstracts the USB HID custom device that feeds button presses.
// The real device (Elgato Stream Deck protocol, or a DIY RadKeys device) is
// read via hidapi when CGO is enabled; a MockReader is used for development.
package hid

import "sync"

// Event reports a change in a button's pressed state.
type Event struct {
	Index   int
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

// Put injects an event (non-blocking). Safe to call before or after Close.
func (m *MockReader) Put(e Event) {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return
	}
	m.mu.Unlock()
	select {
	case <-m.done:
	case m.ch <- e:
	}
}
