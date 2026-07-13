// Package hid abstracts the USB HID custom device that feeds button presses.
// The real device (RadKeys DIY keypad with RP2040-Zero) is read via hidapi
// when CGO is enabled; a MockDevice is used for development.
package hid

import (
	"runtime"
	"sync"
)

// Event reports a change in a button's pressed state at (row, col).
type Event struct {
	Row     int
	Col     int
	Pressed bool
}

// Modifier selects the paste keystroke modifier sent to the firmware.
// Values match the firmware protocol MOD_CTRL / MOD_GUI argument byte.
type Modifier uint8

const (
	ModifierCtrl Modifier = 0x01
	ModifierGUI  Modifier = 0x02
)

// ModifierForOS returns the paste modifier for the host OS: GUI (Cmd) on
// macOS, Ctrl elsewhere.
func ModifierForOS() Modifier {
	if runtime.GOOS == "darwin" {
		return ModifierGUI
	}
	return ModifierCtrl
}

// Device reads button events and writes host-to-device commands on one HID
// handle (the vendor interface, usage page 0xFF00).
type Device interface {
	Open() error
	Events() <-chan Event
	FirePaste(mod Modifier) error
	Close() error
}

// MockDevice is an in-process Device for development without hardware.
type MockDevice struct {
	ch     chan Event
	done   chan struct{}
	closed bool
	mu     sync.Mutex
	once   sync.Once
	pastes []Modifier
}

// NewMock returns a MockDevice.
func NewMock() *MockDevice {
	return &MockDevice{ch: make(chan Event, 64), done: make(chan struct{})}
}

func (m *MockDevice) Open() error          { return nil }
func (m *MockDevice) Events() <-chan Event { return m.ch }

// FirePaste records the modifier so tests can assert the call was made.
func (m *MockDevice) FirePaste(mod Modifier) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pastes = append(m.pastes, mod)
	return nil
}

// PasteCalls returns a copy of the modifiers passed to FirePaste.
func (m *MockDevice) PasteCalls() []Modifier {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Modifier, len(m.pastes))
	copy(out, m.pastes)
	return out
}

func (m *MockDevice) Close() error {
	m.mu.Lock()
	m.closed = true
	m.mu.Unlock()
	m.once.Do(func() { close(m.done); close(m.ch) })
	return nil
}

// Put injects an event. Non-blocking: drops if the buffer is full.
// Safe to call before, during, or after Close — the mutex prevents sending
// on a channel that Close is about to close.
func (m *MockDevice) Put(e Event) {
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
