//go:build windows

package keystroke

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	vkControl = 0x11
	vkKeyV    = 0x56

	inputKeyboard  = 1
	keyeventfKeyUp = 0x0002
)

var (
	user32    = syscall.NewLazyDLL("user32.dll")
	sendInput = user32.NewProc("SendInput")
)

// keyboardInput mirrors the Win32 KEYBDINPUT struct (amd64 layout):
// wVk(2) + wScan(2) + dwFlags(4) + time(4) + pad(4) + dwExtraInfo(8) = 24 bytes.
type keyboardInput struct {
	wVk         uint16
	wScan       uint16
	dwFlags     uint32
	time        uint32
	dwExtraInfo uintptr
}

// input mirrors the Win32 INPUT struct (amd64 layout):
// type(4) + pad(4) + union(32) = 40 bytes. The union is MOUSEINPUT-sized (32),
// so the keyboardInput (24) is followed by 8 padding bytes.
type input struct {
	typ uint32
	ki  keyboardInput
	_   [8]byte
}

// sendCtrlV sends Ctrl+V to the focused window via SendInput, the modern
// replacement for the deprecated keybd_event. SendInput returns the number of
// events actually injected, so a partial/failed injection is surfaced as an
// error instead of being silently lost.
func sendCtrlV() error {
	inputs := []input{
		{typ: inputKeyboard, ki: keyboardInput{wVk: vkControl}},
		{typ: inputKeyboard, ki: keyboardInput{wVk: vkKeyV}},
		{typ: inputKeyboard, ki: keyboardInput{wVk: vkKeyV, dwFlags: keyeventfKeyUp}},
		{typ: inputKeyboard, ki: keyboardInput{wVk: vkControl, dwFlags: keyeventfKeyUp}},
	}
	cbSize := unsafe.Sizeof(input{})
	n, _, err := sendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(cbSize),
	)
	if n != uintptr(len(inputs)) {
		return fmt.Errorf("sendinput: injected %d of %d events: %w", n, len(inputs), err)
	}
	return nil
}
