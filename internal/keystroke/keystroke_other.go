//go:build !linux && !windows

package keystroke

import "errors"

func sendCtrlV() error {
	return errors.New("keystroke: unsupported platform")
}
