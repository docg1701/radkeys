// Package hid abstracts the USB HID custom device that feeds button presses.
// The real device (RadKeys DIY keypad with RP2040-Zero) is read via hidapi
// when CGO is enabled; a MockDevice is used for development.
package hid

import (
	"errors"
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

// MinFirmwareMajor and MinFirmwareMinor are the minimum firmware version
// the host requires. The host warns the user once at connect if the device
// reports a lower version or fails to respond.
const (
	MinFirmwareMajor = 1
	MinFirmwareMinor = 0
)

// errFirmwareVersionUnknown is returned by Version when the device did not
// respond to GET_VERSION (pre-v1.0 firmware or no reply within timeout).
var errFirmwareVersionUnknown = errors.New("hid: firmware version unknown")

// Device reads button events and writes host-to-device commands on one HID
// handle (the vendor interface, usage page 0xFF00).
type Device interface {
	Open() error
	Events() <-chan Event
	FirePaste(mod Modifier) error
	Version() (major, minor byte, err error)
	Close() error
}

// FirmwareOutdated reports whether the firmware is too old or its version
// is unknown. known=false always returns true (unknown = warn). When known,
// returns true if major < MinFirmwareMajor, or major == MinFirmwareMajor and
// minor < MinFirmwareMinor. Example:
//
//	FirmwareOutdated(0, 9, true)  // true — v0.9 is older than v1.0
//	FirmwareOutdated(1, 0, true)  // false — v1.0 meets the minimum
//	FirmwareOutdated(0, 0, false) // true — version unknown
func FirmwareOutdated(major, minor byte, known bool) bool {
	if !known {
		return true
	}
	if major < MinFirmwareMajor {
		return true
	}
	if major == MinFirmwareMajor && minor < MinFirmwareMinor {
		return true
	}
	return false
}

// MockDevice is an in-process Device for development without hardware.
type MockDevice struct {
	ch           chan Event
	done         chan struct{}
	closed       bool
	mu           sync.Mutex
	once         sync.Once
	pastes       []Modifier
	versionMajor byte
	versionMinor byte
	versionKnown bool
}

// NewMock returns a MockDevice with firmware v1.0 known (mock mode does
// not trigger the outdated warning).
func NewMock() *MockDevice {
	return &MockDevice{
		ch:           make(chan Event, 64),
		done:         make(chan struct{}),
		versionMajor: 1,
		versionMinor: 0,
		versionKnown: true,
	}
}

func (m *MockDevice) Open() error          { return nil }
func (m *MockDevice) Events() <-chan Event { return m.ch }

// Version returns the mock firmware version. Returns an error when the
// version is unknown (set via SetFirmwareVersion with major=0).
func (m *MockDevice) Version() (byte, byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.versionKnown {
		return 0, 0, errFirmwareVersionUnknown
	}
	return m.versionMajor, m.versionMinor, nil
}

// SetFirmwareVersion overrides the mock firmware version for tests. Passing
// major=0 marks the version as unknown (simulates a pre-v1.0 device).
func (m *MockDevice) SetFirmwareVersion(major, minor byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.versionMajor = major
	m.versionMinor = minor
	m.versionKnown = major != 0
}

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
