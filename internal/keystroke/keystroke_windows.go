//go:build windows

package keystroke

import "syscall"

var (
	user32     = syscall.NewLazyDLL("user32.dll")
	keybdEvent = user32.NewProc("keybd_event")
)

const (
	vkControl = 0x11
	vkKeyV    = 0x56
	keyDown   = 0
	keyUp     = 0x02
)

func sendCtrlV() error {
	keybdEvent.Call(uintptr(vkControl), 0, uintptr(keyDown), 0)
	keybdEvent.Call(uintptr(vkKeyV), 0, uintptr(keyDown), 0)
	keybdEvent.Call(uintptr(vkKeyV), 0, uintptr(keyUp), 0)
	keybdEvent.Call(uintptr(vkControl), 0, uintptr(keyUp), 0)
	return nil
}
