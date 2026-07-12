// Package theme defines 13 preset colour themes for the RadKeys UI.
// Display names are translated via i18n keys (theme.<id>).
// Colours are explicit per variant — NOT derived with lighten/darken/blend.
package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Preset holds a named theme with explicit colours for light and dark.
type Preset struct {
	ID    string // machine-readable key (i18n: theme.<id>)
	Name  string // fallback display name (English)
	Light PresetColours
	Dark  PresetColours
}

// SystemDefault is the pseudo-preset that delegates to theme.DefaultTheme().
var SystemDefault = Preset{ID: "system", Name: "System default"}

// Presets is the list of all selectable themes (13). Index 0 is the system default.
var Presets = []Preset{
	SystemDefault,
	dracula,
	solarizedDark,
	monokai,
	gruvboxDark,
	nord,
	oneDark,
	tokyoNight,
	catppuccinMocha,
	solarizedLight,
	gruvboxLight,
	lightGray,
	darkGray,
}

// FindPreset returns the Preset with the given ID, or Presets[0] if not found.
func FindPreset(id string) (Preset, bool) {
	for _, p := range Presets {
		if p.ID == id || p.Name == id {
			return p, true
		}
	}
	return Presets[0], false
}

// NewCustomTheme constructs a fyne.Theme from a Preset.
// The system preset returns DefaultTheme.
func NewCustomTheme(p Preset) fyne.Theme {
	if p.ID == "system" {
		return theme.DefaultTheme()
	}
	return newTheme(p.Light, p.Dark)
}

// newTheme is the internal constructor used by presets.
func newTheme(light, dark PresetColours) *RadKeysTheme {
	return &RadKeysTheme{light: light, dark: dark}
}

// Helper to build PresetColours inline.
func pc(pairs ...interface{}) PresetColours {
	m := make(PresetColours, len(pairs)/2)
	for i := 0; i < len(pairs)-1; i += 2 {
		name := pairs[i].(fyne.ThemeColorName)
		c := pairs[i+1].(color.NRGBA)
		m[name] = c
	}
	return m
}

// nrgb is a shorthand for color.NRGBA{} with full alpha.
func nrgb(r, g, b uint8) color.NRGBA { return color.NRGBA{R: r, G: g, B: b, A: 0xff} }

// nrgbA is a shorthand for color.NRGBA{} with explicit alpha.
func nrgbA(r, g, b, a uint8) color.NRGBA { return color.NRGBA{R: r, G: g, B: b, A: a} }

// ─────────────────────────────────────────────────────────────────────────────
// Theme definitions — each preset defines explicit colours for light + dark.
// Colours not listed fall back to theme.DefaultTheme().Color(name, variant).
// ─────────────────────────────────────────────────────────────────────────────

var dracula = Preset{
	ID: "dracula", Name: "Dracula",
	Dark: pc(
		theme.ColorNameBackground, nrgb(0x28, 0x2a, 0x36),
		theme.ColorNameButton, nrgb(0x44, 0x47, 0x5a),
		theme.ColorNameHeaderBackground, nrgb(0x35, 0x37, 0x47),
		theme.ColorNameInputBackground, nrgb(0x34, 0x36, 0x46),
		theme.ColorNameForeground, nrgb(0xf8, 0xf8, 0xf2),
		theme.ColorNamePrimary, nrgb(0xbd, 0x93, 0xf9),
		theme.ColorNameSelection, nrgbA(0x62, 0x72, 0xa4, 0x40),
		theme.ColorNameHover, nrgb(0x50, 0x52, 0x68),
	),
}

var solarizedDark = Preset{
	ID: "solarized_dark", Name: "Solarized Dark",
	Dark: pc(
		theme.ColorNameBackground, nrgb(0x00, 0x2b, 0x36),
		theme.ColorNameButton, nrgb(0x07, 0x36, 0x42),
		theme.ColorNameHeaderBackground, nrgb(0x05, 0x30, 0x3c),
		theme.ColorNameInputBackground, nrgb(0x06, 0x32, 0x3e),
		theme.ColorNameForeground, nrgb(0x83, 0x94, 0x96),
		theme.ColorNamePrimary, nrgb(0x2a, 0xa1, 0x98),
		theme.ColorNameHover, nrgb(0x0d, 0x3e, 0x4a),
	),
}

var monokai = Preset{
	ID: "monokai", Name: "Monokai",
	Dark: pc(
		theme.ColorNameBackground, nrgb(0x27, 0x28, 0x22),
		theme.ColorNameButton, nrgb(0x3e, 0x3d, 0x32),
		theme.ColorNameHeaderBackground, nrgb(0x32, 0x32, 0x2a),
		theme.ColorNameInputBackground, nrgb(0x33, 0x32, 0x2a),
		theme.ColorNameForeground, nrgb(0xf8, 0xf8, 0xf2),
		theme.ColorNamePrimary, nrgb(0xa6, 0xe2, 0x2e),
		theme.ColorNameHover, nrgb(0x49, 0x48, 0x3e),
	),
}

var gruvboxDark = Preset{
	ID: "gruvbox_dark", Name: "Gruvbox Dark",
	Dark: pc(
		theme.ColorNameBackground, nrgb(0x28, 0x28, 0x28),
		theme.ColorNameButton, nrgb(0x3c, 0x38, 0x36),
		theme.ColorNameHeaderBackground, nrgb(0x32, 0x31, 0x2f),
		theme.ColorNameInputBackground, nrgb(0x31, 0x2f, 0x2e),
		theme.ColorNameForeground, nrgb(0xeb, 0xdb, 0xb2),
		theme.ColorNamePrimary, nrgb(0xd7, 0x99, 0x21),
		theme.ColorNameHover, nrgb(0x46, 0x42, 0x40),
	),
}

var nord = Preset{
	ID: "nord", Name: "Nord",
	Dark: pc(
		theme.ColorNameBackground, nrgb(0x2e, 0x34, 0x40),
		theme.ColorNameButton, nrgb(0x3b, 0x42, 0x52),
		theme.ColorNameHeaderBackground, nrgb(0x37, 0x3e, 0x4c),
		theme.ColorNameInputBackground, nrgb(0x35, 0x3c, 0x4a),
		theme.ColorNameForeground, nrgb(0xe5, 0xe9, 0xf0),
		theme.ColorNamePrimary, nrgb(0x88, 0xc0, 0xd0),
		theme.ColorNameHover, nrgb(0x44, 0x4c, 0x5e),
	),
}

var oneDark = Preset{
	ID: "one_dark", Name: "One Dark",
	Dark: pc(
		theme.ColorNameBackground, nrgb(0x28, 0x2c, 0x34),
		theme.ColorNameButton, nrgb(0x35, 0x3b, 0x41),
		theme.ColorNameHeaderBackground, nrgb(0x30, 0x34, 0x3c),
		theme.ColorNameInputBackground, nrgb(0x30, 0x34, 0x3c),
		theme.ColorNameForeground, nrgb(0xab, 0xb2, 0xbf),
		theme.ColorNamePrimary, nrgb(0x61, 0xaf, 0xef),
		theme.ColorNameHover, nrgb(0x3e, 0x44, 0x4c),
	),
}

var tokyoNight = Preset{
	ID: "tokyo_night", Name: "Tokyo Night",
	Dark: pc(
		theme.ColorNameBackground, nrgb(0x1a, 0x1b, 0x26),
		theme.ColorNameButton, nrgb(0x24, 0x28, 0x3b),
		theme.ColorNameHeaderBackground, nrgb(0x1f, 0x21, 0x33),
		theme.ColorNameInputBackground, nrgb(0x1f, 0x20, 0x30),
		theme.ColorNameForeground, nrgb(0xc0, 0xca, 0xf5),
		theme.ColorNamePrimary, nrgb(0x7a, 0xa2, 0xf7),
		theme.ColorNameHover, nrgb(0x2c, 0x31, 0x46),
	),
}

var catppuccinMocha = Preset{
	ID: "catppuccin_mocha", Name: "Catppuccin Mocha",
	Dark: pc(
		theme.ColorNameBackground, nrgb(0x1e, 0x1e, 0x2e),
		theme.ColorNameButton, nrgb(0x31, 0x32, 0x44),
		theme.ColorNameHeaderBackground, nrgb(0x26, 0x26, 0x3a),
		theme.ColorNameInputBackground, nrgb(0x25, 0x25, 0x38),
		theme.ColorNameForeground, nrgb(0xcd, 0xd6, 0xf4),
		theme.ColorNamePrimary, nrgb(0xcb, 0xa6, 0xf7),
		theme.ColorNameHover, nrgb(0x3b, 0x3c, 0x50),
	),
}

var solarizedLight = Preset{
	ID: "solarized_light", Name: "Solarized Light",
	Light: pc(
		theme.ColorNameBackground, nrgb(0xfd, 0xf6, 0xe3),
		theme.ColorNameButton, nrgb(0xee, 0xe8, 0xd5),
		theme.ColorNameHeaderBackground, nrgb(0xf0, 0xe9, 0xd0),
		theme.ColorNameInputBackground, nrgb(0xf4, 0xee, 0xdc),
		theme.ColorNameForeground, nrgb(0x58, 0x6e, 0x75),
		theme.ColorNamePrimary, nrgb(0x2a, 0xa1, 0x98),
		theme.ColorNameHover, nrgb(0xdf, 0xd9, 0xc8),
	),
}

var gruvboxLight = Preset{
	ID: "gruvbox_light", Name: "Gruvbox Light",
	Light: pc(
		theme.ColorNameBackground, nrgb(0xfb, 0xf1, 0xc7),
		theme.ColorNameButton, nrgb(0xeb, 0xdb, 0xb2),
		theme.ColorNameHeaderBackground, nrgb(0xee, 0xde, 0xb5),
		theme.ColorNameInputBackground, nrgb(0xf2, 0xe5, 0xbc),
		theme.ColorNameForeground, nrgb(0x3c, 0x38, 0x36),
		theme.ColorNamePrimary, nrgb(0xd7, 0x99, 0x21),
		theme.ColorNameHover, nrgb(0xdd, 0xcd, 0xa5),
	),
}

var lightGray = Preset{
	ID: "light_gray", Name: "Light Gray",
	Light: pc(
		theme.ColorNameBackground, nrgb(0xe0, 0xe0, 0xe0),
		theme.ColorNameButton, nrgb(0xc0, 0xc0, 0xc0),
		theme.ColorNameHeaderBackground, nrgb(0xcc, 0xcc, 0xcc),
		theme.ColorNameInputBackground, nrgb(0xd5, 0xd5, 0xd5),
		theme.ColorNameForeground, nrgb(0x20, 0x20, 0x20),
		theme.ColorNamePrimary, nrgb(0x40, 0x40, 0xff),
		theme.ColorNameHover, nrgb(0xb0, 0xb0, 0xb0),
	),
}

var darkGray = Preset{
	ID: "dark_gray", Name: "Dark Gray",
	Dark: pc(
		theme.ColorNameBackground, nrgb(0x20, 0x20, 0x20),
		theme.ColorNameButton, nrgb(0x30, 0x30, 0x30),
		theme.ColorNameHeaderBackground, nrgb(0x28, 0x28, 0x28),
		theme.ColorNameInputBackground, nrgb(0x28, 0x28, 0x28),
		theme.ColorNameForeground, nrgb(0xd0, 0xd0, 0xd0),
		theme.ColorNamePrimary, nrgb(0x60, 0x60, 0xff),
		theme.ColorNameHover, nrgb(0x38, 0x38, 0x38),
	),
}
