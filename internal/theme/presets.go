// Package theme defines 13 preset colour themes for the RadKeys UI.
// Display names are translated via i18n keys (theme.<id>).
// Each preset defines a few base colours; the theme derives the remaining
// 28 ThemeColorName values from these base colours.
package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Preset holds a named theme with base colours for light and dark variants.
type Preset struct {
	ID    string // machine-readable key (i18n: theme.<id>)
	Name  string // fallback display name (English)
	Light *BaseColours
	Dark  *BaseColours
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
func NewCustomTheme(p Preset) fyne.Theme {
	if p.ID == "system" {
		return theme.DefaultTheme()
	}
	return newCustomTheme(p.Light, p.Dark)
}

// bc is a shorthand for &BaseColours{}.
func bc(bg, fg, primary, button, header, input, hover [3]uint8) *BaseColours {
	return &BaseColours{
		Bg:      nrgb(bg[0], bg[1], bg[2]),
		Fg:      nrgb(fg[0], fg[1], fg[2]),
		Primary: nrgb(primary[0], primary[1], primary[2]),
		Button:  nrgb(button[0], button[1], button[2]),
		Header:  nrgb(header[0], header[1], header[2]),
		Input:   nrgb(input[0], input[1], input[2]),
		Hover:   nrgb(hover[0], hover[1], hover[2]),
	}
}

func nrgb(r, g, b uint8) color.NRGBA { return color.NRGBA{R: r, G: g, B: b, A: 0xff} }

// ─── Theme definitions ─────────────────────────────────────────────────────

var dracula = Preset{ID: "dracula", Name: "Dracula",
	Dark: bc(
		[3]uint8{0x28, 0x2a, 0x36}, // bg
		[3]uint8{0xf8, 0xf8, 0xf2}, // fg
		[3]uint8{0xbd, 0x93, 0xf9}, // primary
		[3]uint8{0x44, 0x47, 0x5a}, // button
		[3]uint8{0x35, 0x37, 0x47}, // header
		[3]uint8{0x34, 0x36, 0x46}, // input
		[3]uint8{0x50, 0x52, 0x68}, // hover
	),
}

var solarizedDark = Preset{ID: "solarized_dark", Name: "Solarized Dark",
	Dark: bc(
		[3]uint8{0x00, 0x2b, 0x36},
		[3]uint8{0x83, 0x94, 0x96},
		[3]uint8{0x2a, 0xa1, 0x98},
		[3]uint8{0x07, 0x36, 0x42},
		[3]uint8{0x05, 0x30, 0x3c},
		[3]uint8{0x06, 0x32, 0x3e},
		[3]uint8{0x0d, 0x3e, 0x4a},
	),
}

var monokai = Preset{ID: "monokai", Name: "Monokai",
	Dark: bc(
		[3]uint8{0x27, 0x28, 0x22},
		[3]uint8{0xf8, 0xf8, 0xf2},
		[3]uint8{0xa6, 0xe2, 0x2e},
		[3]uint8{0x3e, 0x3d, 0x32},
		[3]uint8{0x32, 0x32, 0x2a},
		[3]uint8{0x33, 0x32, 0x2a},
		[3]uint8{0x49, 0x48, 0x3e},
	),
}

var gruvboxDark = Preset{ID: "gruvbox_dark", Name: "Gruvbox Dark",
	Dark: bc(
		[3]uint8{0x28, 0x28, 0x28},
		[3]uint8{0xeb, 0xdb, 0xb2},
		[3]uint8{0xd7, 0x99, 0x21},
		[3]uint8{0x3c, 0x38, 0x36},
		[3]uint8{0x32, 0x31, 0x2f},
		[3]uint8{0x31, 0x2f, 0x2e},
		[3]uint8{0x46, 0x42, 0x40},
	),
}

var nord = Preset{ID: "nord", Name: "Nord",
	Dark: bc(
		[3]uint8{0x2e, 0x34, 0x40},
		[3]uint8{0xe5, 0xe9, 0xf0},
		[3]uint8{0x88, 0xc0, 0xd0},
		[3]uint8{0x3b, 0x42, 0x52},
		[3]uint8{0x37, 0x3e, 0x4c},
		[3]uint8{0x35, 0x3c, 0x4a},
		[3]uint8{0x44, 0x4c, 0x5e},
	),
}

var oneDark = Preset{ID: "one_dark", Name: "One Dark",
	Dark: bc(
		[3]uint8{0x28, 0x2c, 0x34},
		[3]uint8{0xab, 0xb2, 0xbf},
		[3]uint8{0x61, 0xaf, 0xef},
		[3]uint8{0x35, 0x3b, 0x41},
		[3]uint8{0x30, 0x34, 0x3c},
		[3]uint8{0x30, 0x34, 0x3c},
		[3]uint8{0x3e, 0x44, 0x4c},
	),
}

var tokyoNight = Preset{ID: "tokyo_night", Name: "Tokyo Night",
	Dark: bc(
		[3]uint8{0x1a, 0x1b, 0x26},
		[3]uint8{0xc0, 0xca, 0xf5},
		[3]uint8{0x7a, 0xa2, 0xf7},
		[3]uint8{0x24, 0x28, 0x3b},
		[3]uint8{0x1f, 0x21, 0x33},
		[3]uint8{0x1f, 0x20, 0x30},
		[3]uint8{0x2c, 0x31, 0x46},
	),
}

var catppuccinMocha = Preset{ID: "catppuccin_mocha", Name: "Catppuccin Mocha",
	Dark: bc(
		[3]uint8{0x1e, 0x1e, 0x2e},
		[3]uint8{0xcd, 0xd6, 0xf4},
		[3]uint8{0xcb, 0xa6, 0xf7},
		[3]uint8{0x31, 0x32, 0x44},
		[3]uint8{0x26, 0x26, 0x3a},
		[3]uint8{0x25, 0x25, 0x38},
		[3]uint8{0x3b, 0x3c, 0x50},
	),
}

var solarizedLight = Preset{ID: "solarized_light", Name: "Solarized Light",
	Light: bc(
		[3]uint8{0xfd, 0xf6, 0xe3},
		[3]uint8{0x58, 0x6e, 0x75},
		[3]uint8{0x2a, 0xa1, 0x98},
		[3]uint8{0xee, 0xe8, 0xd5},
		[3]uint8{0xf0, 0xe9, 0xd0},
		[3]uint8{0xf4, 0xee, 0xdc},
		[3]uint8{0xdf, 0xd9, 0xc8},
	),
}

var gruvboxLight = Preset{ID: "gruvbox_light", Name: "Gruvbox Light",
	Light: bc(
		[3]uint8{0xfb, 0xf1, 0xc7},
		[3]uint8{0x3c, 0x38, 0x36},
		[3]uint8{0xd7, 0x99, 0x21},
		[3]uint8{0xeb, 0xdb, 0xb2},
		[3]uint8{0xee, 0xde, 0xb5},
		[3]uint8{0xf2, 0xe5, 0xbc},
		[3]uint8{0xdd, 0xcd, 0xa5},
	),
}

var lightGray = Preset{ID: "light_gray", Name: "Light Gray",
	Light: bc(
		[3]uint8{0xe0, 0xe0, 0xe0},
		[3]uint8{0x20, 0x20, 0x20},
		[3]uint8{0x40, 0x40, 0xff},
		[3]uint8{0xc0, 0xc0, 0xc0},
		[3]uint8{0xcc, 0xcc, 0xcc},
		[3]uint8{0xd5, 0xd5, 0xd5},
		[3]uint8{0xb0, 0xb0, 0xb0},
	),
}

var darkGray = Preset{ID: "dark_gray", Name: "Dark Gray",
	Dark: bc(
		[3]uint8{0x20, 0x20, 0x20},
		[3]uint8{0xd0, 0xd0, 0xd0},
		[3]uint8{0x60, 0x60, 0xff},
		[3]uint8{0x30, 0x30, 0x30},
		[3]uint8{0x28, 0x28, 0x28},
		[3]uint8{0x28, 0x28, 0x28},
		[3]uint8{0x38, 0x38, 0x38},
	),
}
