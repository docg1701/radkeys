//go:build darwin

package keystroke

import "os/exec"

// sendCtrlV sends Cmd+V to the focused window via AppleScript.
// Requires accessibility permissions (System Settings > Privacy & Security > Accessibility).
func sendCtrlV() error {
	return exec.Command("osascript", "-e",
		`tell application "System Events" to keystroke "v" using command down`,
	).Run()
}