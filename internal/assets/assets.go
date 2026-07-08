// Package assets embeds app icons (Obsidian icon theme, CC0) and the default
// fallback icon into the binary. Release = 1 executable + 1 config file.
package assets

import (
	"embed"
	"sort"
)

//go:embed icon.png
var IconPNG []byte

//go:embed icons/*.png
var iconFS embed.FS

// IconNames returns sorted list of available embedded icon names (without extension).
func IconNames() []string {
	entries, err := iconFS.ReadDir("icons")
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		// Strip .png suffix.
		name := e.Name()
		if len(name) > 4 && name[len(name)-4:] == ".png" {
			name = name[:len(name)-4]
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// IconData returns the raw PNG bytes for an embedded icon by name, or nil if not found.
func IconData(name string) []byte {
	b, _ := iconFS.ReadFile("icons/" + name + ".png")
	return b
}
