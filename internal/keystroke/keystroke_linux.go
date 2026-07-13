//go:build linux

package keystroke

import "os/exec"

// sendCtrlV sends Ctrl+V to the focused window via xdotool.
func sendCtrlV() error {
	return exec.Command("xdotool", "key", "ctrl+v").Run()
}
