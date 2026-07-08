// Package assets embeds the app icon (Obsidian icon theme) into the binary.
// Release = 1 executable (icon embedded) + 1 config file. No external assets.
package assets

import _ "embed"

//go:embed icon.png
var IconPNG []byte

//go:embed icon.svg
var IconSVG string
