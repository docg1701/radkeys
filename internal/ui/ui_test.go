package ui

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	fyneTheme "fyne.io/fyne/v2/theme"
)

type dummyColorTheme struct {
	bg color.NRGBA
}

func (t *dummyColorTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	if name == fyneTheme.ColorNameBackground {
		return t.bg
	}
	return color.Black
}

func (t *dummyColorTheme) Font(fyne.TextStyle) fyne.Resource     { return nil }
func (t *dummyColorTheme) Icon(fyne.ThemeIconName) fyne.Resource { return nil }
func (t *dummyColorTheme) Size(fyne.ThemeSizeName) float32       { return 0 }

func TestVariantForDefaultTheme(t *testing.T) {
	// DefaultTheme variant depends on the OS — just ensure it doesn't panic.
	_ = variantFor(fyneTheme.DefaultTheme())
}

func TestVariantForDarkBg(t *testing.T) {
	th := &dummyColorTheme{bg: color.NRGBA{R: 0x10, G: 0x10, B: 0x10, A: 0xff}}
	v := variantFor(th)
	if v != fyneTheme.VariantDark {
		t.Fatalf("dark background should return VariantDark, got %v", v)
	}
}

func TestVariantForLightBg(t *testing.T) {
	th := &dummyColorTheme{bg: color.NRGBA{R: 0xf0, G: 0xf0, B: 0xf0, A: 0xff}}
	v := variantFor(th)
	if v != fyneTheme.VariantLight {
		t.Fatalf("light background should return VariantLight, got %v", v)
	}
}
