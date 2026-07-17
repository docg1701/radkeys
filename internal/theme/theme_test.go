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

func TestFindPresetByDisplayNameNotSupported(t *testing.T) {
	_, ok := FindPreset("Dracula")
	if ok {
		t.Fatal("name-based lookup removed — should not find by display name")
	}
}

func TestFindPresetKnown(t *testing.T) {
	p, ok := FindPreset("system")
	if !ok {
		t.Fatal("expected system preset to be found")
	}
	if p.ID() != "system" {
		t.Fatalf("id = %q, want system", p.ID())
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

func TestPresetsUnique(t *testing.T) {
	seen := make(map[string]struct{}, len(Presets))
	for _, p := range Presets {
		if _, ok := seen[p.ID()]; ok {
			t.Fatalf("duplicate preset id %q", p.ID())
		}
		seen[p.ID()] = struct{}{}
	}
}

func presetOrFatal(t *testing.T, id string) preset {
	t.Helper()
	p, ok := FindPreset(id)
	if !ok {
		t.Fatalf("preset %q not found", id)
	}
	return p
}

func TestNewCustomThemeSystem(t *testing.T) {
	th := NewCustomTheme(presetOrFatal(t, "system"))
	if th != theme.DefaultTheme() {
		t.Fatal("system preset should return DefaultTheme")
	}
}

func TestNewCustomThemeDracula(t *testing.T) {
	th := NewCustomTheme(presetOrFatal(t, "dracula"))
	if th == nil {
		t.Fatal("dracula theme is nil")
	}
	if th == theme.DefaultTheme() {
		t.Fatal("dracula should not be DefaultTheme")
	}
}

func TestDarkThemeColorsExist(t *testing.T) {
	th := NewCustomTheme(presetOrFatal(t, "dracula"))
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
	th := NewCustomTheme(presetOrFatal(t, "solarized_light"))
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
	th := NewCustomTheme(presetOrFatal(t, "dracula"))
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

func TestShiftFactorGreaterThanOneSaturates(t *testing.T) {
	mid := color.NRGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xff}
	got := shift(mid, 2.0)
	want := color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	if got != want {
		t.Fatalf("shift(mid, 2.0) = %v, want %v (saturated white)", got, want)
	}
}

func TestShiftFactorLessThanNegativeOneSaturates(t *testing.T) {
	mid := color.NRGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xff}
	got := shift(mid, -2.0)
	want := color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff}
	if got != want {
		t.Fatalf("shift(mid, -2.0) = %v, want %v (saturated black)", got, want)
	}
}

func TestColorConcurrentNoRace(t *testing.T) {
	th := NewCustomTheme(presetOrFatal(t, "dracula"))
	const n = 100
	done := make(chan struct{}, 2)
	for i := 0; i < 2; i++ {
		go func() {
			for j := 0; j < n; j++ {
				_ = th.Color(theme.ColorNamePressed, theme.VariantDark)
				_ = th.Color(theme.ColorNamePressed, theme.VariantLight)
			}
			done <- struct{}{}
		}()
	}
	<-done
	<-done
}

func TestBlendMidpoint(t *testing.T) {
	got := blend(color.NRGBA{0, 0, 0, 255}, color.NRGBA{255, 255, 255, 255}, 0.5)
	if got.R != 127 || got.G != 127 || got.B != 127 || got.A != 255 {
		t.Fatalf("blend midpoint = %v, want (127,127,127,255)", got)
	}
}

func TestSatAddSaturates(t *testing.T) {
	if satAdd(200, 100) != 255 {
		t.Fatal("satAdd(200,100) should saturate to 255")
	}
	if satAdd(100, 50) != 150 {
		t.Fatal("satAdd(100,50) should be 150")
	}
}

func TestSatSubFloors(t *testing.T) {
	if satSub(50, 100) != 0 {
		t.Fatal("satSub(50,100) should floor to 0")
	}
	if satSub(100, 50) != 50 {
		t.Fatal("satSub(100,50) should be 50")
	}
}

func TestContrastOf(t *testing.T) {
	if got := contrastOf(color.NRGBA{255, 255, 255, 255}); got != (color.NRGBA{0, 0, 0, 255}) {
		t.Fatalf("contrastOf(white) = %v, want black", got)
	}
	if got := contrastOf(color.NRGBA{0, 0, 0, 255}); got != (color.NRGBA{255, 255, 255, 255}) {
		t.Fatalf("contrastOf(black) = %v, want white", got)
	}
}

func TestLerp(t *testing.T) {
	if lerp(0, 100, 0.5) != 50 {
		t.Fatal("lerp(0,100,0.5) should be 50")
	}
}

func TestSetAlpha(t *testing.T) {
	got := setAlpha(color.NRGBA{255, 255, 255, 255}, 0x80)
	if got.A != 0x80 {
		t.Fatalf("setAlpha A = %v, want 0x80", got.A)
	}
}
