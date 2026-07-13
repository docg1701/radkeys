package theme

import (
	"image/color"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func TestFindPresetExists(t *testing.T) {
	p, ok := FindPreset("dracula")
	if !ok {
		t.Fatal("dracula preset not found")
	}
	if p.ID() != "dracula" {
		t.Fatalf("id = %q, want dracula", p.ID())
	}
}

func TestFindPresetByName(t *testing.T) {
	p, ok := FindPreset("Dracula")
	if !ok {
		t.Fatal("Dracula preset not found by display name")
	}
	if p.ID() != "dracula" {
		t.Fatalf("id = %q, want dracula", p.ID())
	}
}

func TestFindPresetUnknown(t *testing.T) {
	p, ok := FindPreset("nonexistent")
	if ok {
		t.Fatal("expected not found for nonexistent preset")
	}
	if p.ID() != "system" {
		t.Fatalf("fallback id = %q, want system", p.ID())
	}
}

func TestNewCustomThemeSystem(t *testing.T) {
	th := NewCustomTheme(systemDefault)
	if th != theme.DefaultTheme() {
		t.Fatal("system preset should return DefaultTheme")
	}
}

func TestNewCustomThemeDracula(t *testing.T) {
	th := NewCustomTheme(dracula)
	if th == nil {
		t.Fatal("dracula theme is nil")
	}
	if th == theme.DefaultTheme() {
		t.Fatal("dracula should not be DefaultTheme")
	}
}

func TestDarkThemeColorsExist(t *testing.T) {
	th := NewCustomTheme(dracula)
	for _, name := range []fyne.ThemeColorName{
		theme.ColorNameBackground,
		theme.ColorNameButton,
		theme.ColorNameForeground,
		theme.ColorNamePrimary,
		theme.ColorNameHover,
		theme.ColorNameHeaderBackground,
		theme.ColorNameInputBackground,
		theme.ColorNameSelection,
		theme.ColorNameDisabled,
		theme.ColorNamePlaceHolder,
		theme.ColorNameScrollBar,
		theme.ColorNameSeparator,
	} {
		c := th.Color(name, theme.VariantDark)
		if c == nil {
			t.Fatalf("Color(%s, VariantDark) returned nil", name)
		}
	}
}

func TestLightThemeColorsExist(t *testing.T) {
	th := NewCustomTheme(solarizedLight)
	for _, name := range []fyne.ThemeColorName{
		theme.ColorNameBackground,
		theme.ColorNameButton,
		theme.ColorNameForeground,
		theme.ColorNamePrimary,
		theme.ColorNameHover,
	} {
		c := th.Color(name, theme.VariantLight)
		if c == nil {
			t.Fatalf("Color(%s, VariantLight) returned nil", name)
		}
	}
}

func TestDarkOnlyThemeStaysDarkInLightVariant(t *testing.T) {
	th := NewCustomTheme(dracula)
	darkBg := th.Color(theme.ColorNameBackground, theme.VariantDark)
	lightBg := th.Color(theme.ColorNameBackground, theme.VariantLight)
	if darkBg != lightBg {
		t.Fatal("dark-only theme should use dark colors even when variant is Light")
	}
}

func TestAllPresetsHaveValidThemes(t *testing.T) {
	for _, p := range Presets {
		th := NewCustomTheme(p)
		if th == nil {
			t.Fatalf("preset %q returned nil theme", p.ID())
		}
		// System preset requires a running Fyne app — skip.
		if p.ID() == "system" {
			continue
		}
		if bg := th.Color(theme.ColorNameBackground, theme.VariantDark); bg == nil {
			t.Fatalf("preset %q background is nil", p.ID())
		}
	}
}

// TestShiftNegativeFactorDarkensNotClampsToBlack is a regression test for the
// pressed-color bug: uint8(255 * -0.08) wrapped to 236, satSub saturated to 0,
// so light-theme buttons flashed black when pressed.
func TestShiftNegativeFactorDarkensNotClampsToBlack(t *testing.T) {
	light := color.NRGBA{R: 0xc0, G: 0xc0, B: 0xc0, A: 0xff}
	got := shift(light, -0.08)
	if got.R == 0 && got.G == 0 && got.B == 0 {
		t.Fatalf("shift(light, -0.08) clamped to black %v; want darkened-but-not-zero", got)
	}
	if got.R >= light.R || got.G >= light.G || got.B >= light.B {
		t.Fatalf("shift(light, -0.08) = %v did not darken %v", got, light)
	}
}

func TestShiftPositiveFactorLightens(t *testing.T) {
	dark := color.NRGBA{R: 0x44, G: 0x47, B: 0x5a, A: 0xff}
	got := shift(dark, 0.08)
	if got.R <= dark.R || got.G <= dark.G || got.B <= dark.B {
		t.Fatalf("shift(dark, 0.08) = %v did not lighten %v", got, dark)
	}
}
