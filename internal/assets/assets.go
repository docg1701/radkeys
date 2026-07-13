// Package assets embeds the app icon into the binary.
// Release = 1 executable + 1 config file.
package assets

import _ "embed"

//go:embed icon.png
var IconPNG []byte
