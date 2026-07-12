// Package theme — custom Fyne theme where every ThemeColorName is explicitly
// resolved from the preset's base colours. No fallback to DefaultTheme.
//
// "Padrão do sistema" delegates entirely to theme.DefaultTheme().
package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

var _ fyne.Theme = (*RadKeysTheme)(nil)

// BaseColours are the minimum colours a preset must define.
// All other ThemeColorName values are derived from these.
type BaseColours struct {
	Bg      color.NRGBA // page background
	Fg      color.NRGBA // primary text
	Primary color.NRGBA // accent (tabs, links, selection tint)
	Button  color.NRGBA // button surface
	Header  color.NRGBA // tab/header bar background
	Input   color.NRGBA // text entry background
	Hover   color.NRGBA // hover highlight
}

// RadKeysTheme resolves every ThemeColorName from BaseColours.
type RadKeysTheme struct {
	light   *BaseColours
	dark    *BaseColours
	isLight bool // cached per-variant
}

// NewCustomTheme returns a fyne.Theme for the given base colours.
// The system preset should use DefaultTheme directly.
func newCustomTheme(light, dark *BaseColours) fyne.Theme {
	return &RadKeysTheme{light: light, dark: dark}
}

func (t *RadKeysTheme) Color(name fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	base := t.dark
	t.isLight = false
	if v == theme.VariantLight {
		base = t.light
		t.isLight = true
	}
	if base == nil {
		return theme.DefaultTheme().Color(name, v)
	}

	switch name {

	// ── structural ──────────────────────────────────────────
	case theme.ColorNameBackground:
		return base.Bg
	case theme.ColorNameButton:
		return base.Button
	case theme.ColorNameHeaderBackground:
		return base.Header
	case theme.ColorNameInputBackground:
		return base.Input
	case theme.ColorNameMenuBackground:
		return base.Button
	case theme.ColorNameOverlayBackground:
		return setAlpha(base.Fg, 0xCC)

	// ── text ────────────────────────────────────────────────
	case theme.ColorNameForeground:
		return base.Fg
	case theme.ColorNameDisabled:
		return blend(base.Bg, base.Fg, 0.50)
	case theme.ColorNamePlaceHolder:
		return blend(base.Bg, base.Fg, 0.60)
	case theme.ColorNameHyperlink:
		return base.Primary

	// ── interactive ─────────────────────────────────────────
	case theme.ColorNameHover:
		return base.Hover
	case theme.ColorNamePressed:
		return shiftBrightness(base.Button, t.variantSign()*-0.08)
	case theme.ColorNameFocus:
		return setAlpha(base.Primary, 0x5c)
	case theme.ColorNameSelection:
		return setAlpha(base.Primary, 0x40)
	case theme.ColorNameDisabledButton:
		return blend(base.Button, base.Bg, 0.50)

	// ── decoration ──────────────────────────────────────────
	case theme.ColorNamePrimary:
		return base.Primary
	case theme.ColorNameScrollBar:
		return blend(base.Bg, base.Fg, 0.14)
	case theme.ColorNameScrollBarBackground:
		return blend(base.Bg, base.Fg, 0.05)
	case theme.ColorNameSeparator:
		return blend(base.Bg, base.Fg, 0.12)
	case theme.ColorNameShadow:
		return setAlpha(base.Fg, 0x10)
	case theme.ColorNameInputBorder:
		return blend(base.Bg, base.Fg, 0.20)

	// ── semantic (standard, NOT derived from preset) ────────
	case theme.ColorNameError:
		return color.NRGBA{R: 0xd3, G: 0x2f, B: 0x2f, A: 0xff}
	case theme.ColorNameSuccess:
		return color.NRGBA{R: 0x38, G: 0x8e, B: 0x3c, A: 0xff}
	case theme.ColorNameWarning:
		return color.NRGBA{R: 0xf5, G: 0x7c, B: 0x00, A: 0xff}

	// ── foreground-on-semantic ──────────────────────────────
	case theme.ColorNameForegroundOnPrimary:
		return contrastOf(base.Primary)
	case theme.ColorNameForegroundOnError:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	case theme.ColorNameForegroundOnSuccess:
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	case theme.ColorNameForegroundOnWarning:
		return color.NRGBA{R: 0x1a, G: 0x1a, B: 0x1a, A: 0xff}
	}

	// Unreachable — all ThemeColorName values are handled above.
	return base.Bg
}

func (t *RadKeysTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *RadKeysTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *RadKeysTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

// variantSign returns +1 for light themes, -1 for dark.
func (t *RadKeysTheme) variantSign() float64 {
	if t.isLight {
		return 1
	}
	return -1
}

// ---------------------------------------------------------------------------
// colour operations
// ---------------------------------------------------------------------------

// shiftBrightness adds (factor * 255) to all RGB channels. Positive = lighter.
func shiftBrightness(c color.NRGBA, factor float64) color.NRGBA {
	d := uint8(255 * factor)
	if factor >= 0 {
		return color.NRGBA{R: satAdd(c.R, d), G: satAdd(c.G, d), B: satAdd(c.B, d), A: c.A}
	}
	return color.NRGBA{R: satSub(c.R, d), G: satSub(c.G, d), B: satSub(c.B, d), A: c.A}
}

func blend(a, b color.NRGBA, t float64) color.NRGBA {
	return color.NRGBA{
		R: lerp(a.R, b.R, t), G: lerp(a.G, b.G, t), B: lerp(a.B, b.B, t), A: a.A,
	}
}

func setAlpha(c color.NRGBA, a uint8) color.NRGBA {
	c.A = a
	return c
}

// contrastOf returns white or black, whichever contrasts more with c.
func contrastOf(c color.NRGBA) color.NRGBA {
	if luminance(c) > 0.45 {
		return color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff}
	}
	return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
}

func luminance(c color.NRGBA) float64 {
	return 0.2126*float64(c.R)/255.0 + 0.7152*float64(c.G)/255.0 + 0.0722*float64(c.B)/255.0
}

func lerp(a, b uint8, t float64) uint8 {
	return uint8(float64(a)*(1-t) + float64(b)*t)
}

func satAdd(a, b uint8) uint8 {
	if s := uint16(a) + uint16(b); s <= 255 {
		return uint8(s)
	}
	return 255
}

func satSub(a, b uint8) uint8 {
	if b > a {
		return 0
	}
	return a - b
}
