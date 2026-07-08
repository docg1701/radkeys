//go:build !cgo

package hid

import (
	"errors"

	"github.com/docg1701/radkeys/internal/config"
)

// Open returns an error when built without CGO; real HID needs hidapi (CGO).
// Use NewMock for development without a device.
func Open(dev config.Device) (Reader, error) {
	return nil, errors.New("hid: real device requires CGO + hidapi (build with CGO_ENABLED=1); use NewMock for development")
}
