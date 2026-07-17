// Package theme — 13 preset color themes for RadKeys.
// Each preset defines 7 base colors; the theme engine derives the remaining
// 28 Fyne ThemeColorName values from these bases. No DefaultTheme fallback.
package theme

import (
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

var _ fyne.Theme = (*radKeysTheme)(nil)

// ─── Theme engine ──────────────────────────────────────────────────────────

// baseColors are the minimum colors a preset must define.
type baseColors struct {
	bg, fg, primary, button, header, input, hover color.NRGBA
}

// radKeysTheme resolves every ThemeColorName from baseColors.
type radKeysTheme struct {
	light, dark *baseColors
}

// CustomThemeMarker is implemented only by *radKeysTheme so that internal/ui
// can recognize our themes without comparing interface values to DefaultTheme().
type CustomThemeMarker interface {
	fyne.Theme
	isRadKeysTheme()
}

func (t *radKeysTheme) isRadKeysTheme() {}

func newTheme(p preset) fyne.Theme {
	if p.id == "system" {
		return theme.DefaultTheme()
	}
	return &radKeysTheme{light: p.light, dark: p.dark}
}

func (t *radKeysTheme) Color(name fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	bc := t.resolve(v)
	if bc == nil {
		return theme.DefaultTheme().Color(name, v)
	}
	sign := -1.0
	if v == theme.VariantLight {
		sign = 1.0
	}

	switch name {
	case theme.ColorNameBackground:
		return bc.bg
	case theme.ColorNameButton:
		return bc.button
	case theme.ColorNameHeaderBackground:
		return bc.header
	case theme.ColorNameInputBackground:
		return bc.input
	case theme.ColorNameMenuBackground:
		return bc.button
	case theme.ColorNameOverlayBackground:
		return setAlpha(bc.fg, 0xCC)
	case theme.ColorNameForeground:
		return bc.fg
	case theme.ColorNameDisabled:
		return blend(bc.bg, bc.fg, 0.50)
	case theme.ColorNamePlaceHolder:
		return blend(bc.bg, bc.fg, 0.60)
	case theme.ColorNameHyperlink:
		return bc.primary
	case theme.ColorNameHover:
		return bc.hover
	case theme.ColorNamePressed:
		return shift(bc.button, sign*-0.08)
	case theme.ColorNameFocus:
		return setAlpha(bc.primary, 0x5c)
	case theme.ColorNameSelection:
		return setAlpha(bc.primary, 0x40)
	case theme.ColorNameDisabledButton:
		return blend(bc.button, bc.bg, 0.50)
	case theme.ColorNamePrimary:
		return bc.primary
	case theme.ColorNameScrollBar:
		return blend(bc.bg, bc.fg, 0.14)
	case theme.ColorNameScrollBarBackground:
		return blend(bc.bg, bc.fg, 0.05)
	case theme.ColorNameSeparator:
		return blend(bc.bg, bc.fg, 0.12)
	case theme.ColorNameShadow:
		return setAlpha(bc.fg, 0x10)
	case theme.ColorNameInputBorder:
		return blend(bc.bg, bc.fg, 0.20)
	case theme.ColorNameError:
		return color.NRGBA{0xd3, 0x2f, 0x2f, 0xff}
	case theme.ColorNameSuccess:
		return color.NRGBA{0x38, 0x8e, 0x3c, 0xff}
	case theme.ColorNameWarning:
		return color.NRGBA{0xf5, 0x7c, 0x00, 0xff}
	case theme.ColorNameForegroundOnPrimary:
		return contrastOf(bc.primary)
	case theme.ColorNameForegroundOnError:
		return color.NRGBA{0xff, 0xff, 0xff, 0xff}
	case theme.ColorNameForegroundOnSuccess:
		return color.NRGBA{0xff, 0xff, 0xff, 0xff}
	case theme.ColorNameForegroundOnWarning:
		return color.NRGBA{0x1a, 0x1a, 0x1a, 0xff}
	}
	return bc.bg
}

// resolve returns the base colors for the requested variant.
// Falls back to the other variant if the requested one is nil.
func (t *radKeysTheme) resolve(v fyne.ThemeVariant) *baseColors {
	if v == theme.VariantLight {
		if t.light != nil {
			return t.light
		}
		return t.dark
	}
	if t.dark != nil {
		return t.dark
	}
	return t.light
}

func (t *radKeysTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}
func (t *radKeysTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}
func (t *radKeysTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

// ─── Color operations ─────────────────────────────────────────────────────

func blend(a, b color.NRGBA, t float64) color.NRGBA {
	return color.NRGBA{
		R: lerp(a.R, b.R, t), G: lerp(a.G, b.G, t), B: lerp(a.B, b.B, t), A: a.A,
	}
}

func setAlpha(c color.NRGBA, a uint8) color.NRGBA { c.A = a; return c }

func contrastOf(c color.NRGBA) color.NRGBA {
	if 0.2126*float64(c.R)+0.7152*float64(c.G)+0.0722*float64(c.B) > 0.45*255 {
		return color.NRGBA{0x00, 0x00, 0x00, 0xff}
	}
	return color.NRGBA{0xff, 0xff, 0xff, 0xff}
}

func shift(c color.NRGBA, factor float64) color.NRGBA {
	// Clamp magnitude to [0,1] so |factor| > 1 cannot wrap uint8 and
	// produce a wrong color.
	mag := math.Abs(factor)
	if mag > 1 {
		mag = 1
	}
	d := uint8(255 * mag)
	if factor >= 0 {
		return color.NRGBA{satAdd(c.R, d), satAdd(c.G, d), satAdd(c.B, d), c.A}
	}
	return color.NRGBA{satSub(c.R, d), satSub(c.G, d), satSub(c.B, d), c.A}
}

func lerp(a, b uint8, t float64) uint8 { return uint8(float64(a)*(1-t) + float64(b)*t) }
func satAdd(a, b uint8) uint8 {
	if uint16(a)+uint16(b) <= 255 {
		return a + b
	}
	return 255
}
func satSub(a, b uint8) uint8 {
	if b > a {
		return 0
	}
	return a - b
}

// ─── Presets ───────────────────────────────────────────────────────────────

type preset struct {
	id    string
	light *baseColors
	dark  *baseColors
}

func newBaseColors(bg, fg, primary, button, header, input, hover [3]uint8) *baseColors {
	return &baseColors{
		bg:      color.NRGBA{bg[0], bg[1], bg[2], 0xff},
		fg:      color.NRGBA{fg[0], fg[1], fg[2], 0xff},
		primary: color.NRGBA{primary[0], primary[1], primary[2], 0xff},
		button:  color.NRGBA{button[0], button[1], button[2], 0xff},
		header:  color.NRGBA{header[0], header[1], header[2], 0xff},
		input:   color.NRGBA{input[0], input[1], input[2], 0xff},
		hover:   color.NRGBA{hover[0], hover[1], hover[2], 0xff},
	}
}

// Presets is the ordered list of all selectable themes. Index 0 = system.
// Adding a new theme requires exactly one new literal here.
var Presets = []preset{
	{id: "system"},
	{id: "dracula", dark: newBaseColors(
		[3]uint8{0x28, 0x2a, 0x36}, [3]uint8{0xf8, 0xf8, 0xf2},
		[3]uint8{0xbd, 0x93, 0xf9}, [3]uint8{0x44, 0x47, 0x5a},
		[3]uint8{0x35, 0x37, 0x47}, [3]uint8{0x34, 0x36, 0x46},
		[3]uint8{0x50, 0x52, 0x68},
	)},
	{id: "solarized_dark", dark: newBaseColors(
		[3]uint8{0x00, 0x2b, 0x36}, [3]uint8{0x83, 0x94, 0x96},
		[3]uint8{0x2a, 0xa1, 0x98}, [3]uint8{0x07, 0x36, 0x42},
		[3]uint8{0x05, 0x30, 0x3c}, [3]uint8{0x06, 0x32, 0x3e},
		[3]uint8{0x0d, 0x3e, 0x4a},
	)},
	{id: "monokai", dark: newBaseColors(
		[3]uint8{0x27, 0x28, 0x22}, [3]uint8{0xf8, 0xf8, 0xf2},
		[3]uint8{0xa6, 0xe2, 0x2e}, [3]uint8{0x3e, 0x3d, 0x32},
		[3]uint8{0x32, 0x32, 0x2a}, [3]uint8{0x33, 0x32, 0x2a},
		[3]uint8{0x49, 0x48, 0x3e},
	)},
	{id: "gruvbox_dark", dark: newBaseColors(
		[3]uint8{0x28, 0x28, 0x28}, [3]uint8{0xeb, 0xdb, 0xb2},
		[3]uint8{0xd7, 0x99, 0x21}, [3]uint8{0x3c, 0x38, 0x36},
		[3]uint8{0x32, 0x31, 0x2f}, [3]uint8{0x31, 0x2f, 0x2e},
		[3]uint8{0x46, 0x42, 0x40},
	)},
	{id: "nord", dark: newBaseColors(
		[3]uint8{0x2e, 0x34, 0x40}, [3]uint8{0xe5, 0xe9, 0xf0},
		[3]uint8{0x88, 0xc0, 0xd0}, [3]uint8{0x3b, 0x42, 0x52},
		[3]uint8{0x37, 0x3e, 0x4c}, [3]uint8{0x35, 0x3c, 0x4a},
		[3]uint8{0x44, 0x4c, 0x5e},
	)},
	{id: "one_dark", dark: newBaseColors(
		[3]uint8{0x28, 0x2c, 0x34}, [3]uint8{0xab, 0xb2, 0xbf},
		[3]uint8{0x61, 0xaf, 0xef}, [3]uint8{0x35, 0x3b, 0x41},
		[3]uint8{0x30, 0x34, 0x3c}, [3]uint8{0x30, 0x34, 0x3c},
		[3]uint8{0x3e, 0x44, 0x4c},
	)},
	{id: "tokyo_night", dark: newBaseColors(
		[3]uint8{0x1a, 0x1b, 0x26}, [3]uint8{0xc0, 0xca, 0xf5},
		[3]uint8{0x7a, 0xa2, 0xf7}, [3]uint8{0x24, 0x28, 0x3b},
		[3]uint8{0x1f, 0x21, 0x33}, [3]uint8{0x1f, 0x20, 0x30},
		[3]uint8{0x2c, 0x31, 0x46},
	)},
	{id: "catppuccin_mocha", dark: newBaseColors(
		[3]uint8{0x1e, 0x1e, 0x2e}, [3]uint8{0xcd, 0xd6, 0xf4},
		[3]uint8{0xcb, 0xa6, 0xf7}, [3]uint8{0x31, 0x32, 0x44},
		[3]uint8{0x26, 0x26, 0x3a}, [3]uint8{0x25, 0x25, 0x38},
		[3]uint8{0x3b, 0x3c, 0x50},
	)},
	{id: "solarized_light", light: newBaseColors(
		[3]uint8{0xfd, 0xf6, 0xe3}, [3]uint8{0x58, 0x6e, 0x75},
		[3]uint8{0x2a, 0xa1, 0x98}, [3]uint8{0xee, 0xe8, 0xd5},
		[3]uint8{0xf0, 0xe9, 0xd0}, [3]uint8{0xf4, 0xee, 0xdc},
		[3]uint8{0xdf, 0xd9, 0xc8},
	)},
	{id: "gruvbox_light", light: newBaseColors(
		[3]uint8{0xfb, 0xf1, 0xc7}, [3]uint8{0x3c, 0x38, 0x36},
		[3]uint8{0xd7, 0x99, 0x21}, [3]uint8{0xeb, 0xdb, 0xb2},
		[3]uint8{0xee, 0xde, 0xb5}, [3]uint8{0xf2, 0xe5, 0xbc},
		[3]uint8{0xdd, 0xcd, 0xa5},
	)},
	{id: "light_gray", light: newBaseColors(
		[3]uint8{0xe0, 0xe0, 0xe0}, [3]uint8{0x20, 0x20, 0x20},
		[3]uint8{0x40, 0x40, 0xff}, [3]uint8{0xc0, 0xc0, 0xc0},
		[3]uint8{0xcc, 0xcc, 0xcc}, [3]uint8{0xd5, 0xd5, 0xd5},
		[3]uint8{0xb0, 0xb0, 0xb0},
	)},
	{id: "dark_gray", dark: newBaseColors(
		[3]uint8{0x20, 0x20, 0x20}, [3]uint8{0xd0, 0xd0, 0xd0},
		[3]uint8{0x60, 0x60, 0xff}, [3]uint8{0x30, 0x30, 0x30},
		[3]uint8{0x28, 0x28, 0x28}, [3]uint8{0x28, 0x28, 0x28},
		[3]uint8{0x38, 0x38, 0x38},
	)},
}

// FindPreset returns the preset with the given id, or system default.
func FindPreset(id string) (preset, bool) {
	for _, p := range Presets {
		if p.id == id {
			return p, true
		}
	}
	return Presets[0], false
}

// NewCustomTheme constructs a fyne.Theme from a preset.
func NewCustomTheme(p preset) fyne.Theme { return newTheme(p) }

// ID returns the machine-readable id (for i18n keys).
func (p preset) ID() string { return p.id }
