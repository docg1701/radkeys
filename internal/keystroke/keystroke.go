// Package keystroke sends keyboard events to the focused window.
// Used by the Paste action to send Ctrl+V to the RIS/PACS without
// the RadKeys app stealing focus.
package keystroke

// SendCtrlV sends Ctrl+V to the currently focused window.
func SendCtrlV() error {
	return sendCtrlV()
}
