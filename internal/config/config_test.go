package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
)

const sample = `
[app]
name = "RadKeys"

[app.device]
vendor_id  = 0x1234
product_id = 0xABCD
protocol   = "radkeys-diy"

[app.layout]
columns = 4
rows    = 3

[[screens]]
id = "root"
name = "Início"

[[screens.buttons]]
row = 0
col = 0
label = "RX"
action = "navigate"
target = "rx_torax"

[[screens.buttons]]
row = 1
col = 0
label = "Voltar"
action = "prev"

[[screens.buttons]]
row = 2
col = 0
label = "Home"
action = "home"

[[screens.buttons]]
row = 2
col = 3
label = "Copy"
action = "copy"

[[screens]]
id = "rx_torax"
name = "RX Tórax"

[[screens.buttons]]
row = 0
col = 0
label = "Normal"
action = "text"
content = "Radiografia de tórax normal."

[[screens.buttons]]
row = 1
col = 0
label = "Voltar"
action = "prev"
`

func writeFile(t *testing.T, name, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

func TestLoadOK(t *testing.T) {
	cfg, err := Load(writeFile(t, "radkeys.config.toml", sample))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.App.Device.Protocol != ProtocolDIY {
		t.Fatalf("protocol = %q, want %q", cfg.App.Device.Protocol, ProtocolDIY)
	}
	if len(cfg.Screens) != 2 {
		t.Fatalf("screens = %d, want 2", len(cfg.Screens))
	}
	root, ok := cfg.ScreenByID("root")
	if !ok {
		t.Fatal("root screen not found")
	}
	if root.Name != "Início" {
		t.Fatalf("root name = %q, want Início", root.Name)
	}
	b, ok := root.ButtonAt(0, 0)
	if !ok {
		t.Fatal("ButtonAt(0,0) not found")
	}
	if b.Action != ActionNavigate || b.Target != "rx_torax" {
		t.Fatalf("button = %+v", b)
	}
}

func TestLoadInvalidProtocol(t *testing.T) {
	body := `
[app]
[app.device]
vendor_id = 1
product_id = 2
protocol = "bogus"
[[screens]]
id = "root"
name = "x"
`
	_, err := Load(writeFile(t, "c.toml", body))
	if err == nil {
		t.Fatal("expected error for invalid protocol")
	}
}

func TestLoadTextRequiresContent(t *testing.T) {
	body := `
[app]
[app.device]
vendor_id = 1
product_id = 2
protocol = "radkeys-diy"
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 0
col = 0
label = "X"
action = "text"
content = ""
`
	_, err := Load(writeFile(t, "c.toml", body))
	if err == nil {
		t.Fatal("expected error for text without content")
	}
}

func TestLoadActionMustNotHaveContent(t *testing.T) {
	body := `
[app]
[app.device]
vendor_id = 1
product_id = 2
protocol = "radkeys-diy"
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 0
col = 0
label = "X"
action = "copy"
content = "nope"
`
	_, err := Load(writeFile(t, "c.toml", body))
	if err == nil {
		t.Fatal("expected error for non-text action with content")
	}
}

func TestLoadInvalidAction(t *testing.T) {
	body := `
[app]
[app.device]
vendor_id = 1
product_id = 2
protocol = "radkeys-diy"
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 0
col = 0
label = "X"
action = "next"
`
	_, err := Load(writeFile(t, "c.toml", body))
	if err == nil {
		t.Fatal("expected error for invalid action")
	}
}

func TestLoadRowOutOfRange(t *testing.T) {
	body := `
[app]
[app.device]
vendor_id = 1
product_id = 2
protocol = "radkeys-diy"
[app.layout]
columns = 4
rows = 3
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 5
col = 0
label = "X"
action = "prev"
`
	_, err := Load(writeFile(t, "c.toml", body))
	if err == nil {
		t.Fatal("expected error for row out of range")
	}
}

func TestLoadColOutOfRange(t *testing.T) {
	body := `
[app]
[app.device]
vendor_id = 1
product_id = 2
protocol = "radkeys-diy"
[app.layout]
columns = 4
rows = 3
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 0
col = 7
label = "X"
action = "prev"
`
	_, err := Load(writeFile(t, "c.toml", body))
	if err == nil {
		t.Fatal("expected error for col out of range")
	}
}

func TestLoadAppliesDefaultsOmittedLayoutIsSixBySix(t *testing.T) {
	body := `
[app]
[app.device]
vendor_id = 1
product_id = 2
protocol = "radkeys-diy"
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 5
col = 5
label = "X"
action = "prev"
`
	cfg, err := Load(writeFile(t, "c.toml", body))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.App.Layout.Columns != 6 {
		t.Fatalf("columns = %d, want 6", cfg.App.Layout.Columns)
	}
	if cfg.App.Layout.Rows != 6 {
		t.Fatalf("rows = %d, want 6", cfg.App.Layout.Rows)
	}
}

func TestLoadAppliesDefaultsOmittedLanguageAndTheme(t *testing.T) {
	body := `
[app]
[app.device]
vendor_id = 1
product_id = 2
protocol = "radkeys-diy"
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 0
col = 0
label = "X"
action = "prev"
`
	cfg, err := Load(writeFile(t, "c.toml", body))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.App.Language != "en" {
		t.Fatalf("language = %q, want en", cfg.App.Language)
	}
	if cfg.App.Theme.Preset != "system" {
		t.Fatalf("theme preset = %q, want system", cfg.App.Theme.Preset)
	}
}

func TestValidateDoesNotMutatePopulatedConfig(t *testing.T) {
	cfg := Config{
		App: App{
			Language: "pt-BR",
			Theme:    Theme{Preset: "dracula"},
			Layout:   Layout{Columns: 5, Rows: 4},
			Device:   Device{Protocol: ProtocolDIY},
		},
		Screens: []Screen{{ID: "root", Name: "x"}},
	}
	want := cfg
	if err := cfg.validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if cfg.App.Language != want.App.Language ||
		cfg.App.Theme.Preset != want.App.Theme.Preset ||
		cfg.App.Layout.Columns != want.App.Layout.Columns ||
		cfg.App.Layout.Rows != want.App.Layout.Rows {
		t.Fatalf("validate mutated config: got %+v, want %+v", cfg, want)
	}
}

func TestLoadOmittedLayoutDoesNotBreakSixBySixButtons(t *testing.T) {
	body := `
[app]
[app.device]
vendor_id = 1
product_id = 2
protocol = "radkeys-diy"
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 5
col = 5
label = "X"
action = "prev"
`
	if _, err := Load(writeFile(t, "c.toml", body)); err != nil {
		t.Fatalf("Load: %v", err)
	}
}

func TestLoadNoScreens(t *testing.T) {
	body := `
[app]
[app.device]
vendor_id = 1
product_id = 2
protocol = "radkeys-diy"
`
	_, err := Load(writeFile(t, "c.toml", body))
	if err == nil {
		t.Fatal("expected error for zero screens")
	}
}

func TestLoadEmptyScreenID(t *testing.T) {
	body := `
[app]
[app.device]
vendor_id = 1
product_id = 2
protocol = "radkeys-diy"
[[screens]]
id = ""
name = "x"
`
	_, err := Load(writeFile(t, "c.toml", body))
	if err == nil {
		t.Fatal("expected error for empty screen id")
	}
}

func TestLoadNavigateRequiresTarget(t *testing.T) {
	body := `
[app]
[app.device]
vendor_id = 1
product_id = 2
protocol = "radkeys-diy"
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 0
col = 0
label = "X"
action = "navigate"
`
	_, err := Load(writeFile(t, "c.toml", body))
	if err == nil {
		t.Fatal("expected error for navigate without target")
	}
}

func TestLoadNavigateUnknownTarget(t *testing.T) {
	body := `
[app]
[app.device]
vendor_id = 1
product_id = 2
protocol = "radkeys-diy"
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 0
col = 0
label = "X"
action = "navigate"
target = "missing"
`
	_, err := Load(writeFile(t, "c.toml", body))
	if err == nil {
		t.Fatal("expected error for navigate to unknown target")
	}
}

func TestLoadDuplicateScreenID(t *testing.T) {
	body := `
[app]
[app.device]
vendor_id = 1
product_id = 2
protocol = "radkeys-diy"
[[screens]]
id = "root"
name = "x"
[[screens]]
id = "root"
name = "y"
`
	_, err := Load(writeFile(t, "c.toml", body))
	if err == nil {
		t.Fatal("expected error for duplicate screen id")
	}
}

func TestRoundtrip(t *testing.T) {
	cfg, err := Load(writeFile(t, "radkeys.config.toml", sample))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(cfg); err != nil {
		t.Fatalf("Encode: %v", err)
	}
	var cfg2 Config
	if err := toml.Unmarshal(buf.Bytes(), &cfg2); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if err := cfg2.validate(); err != nil {
		t.Fatalf("validate roundtripped: %v", err)
	}
	if len(cfg2.Screens) != len(cfg.Screens) {
		t.Fatalf("screen count: %d vs %d", len(cfg2.Screens), len(cfg.Screens))
	}
}

func TestButtonAt(t *testing.T) {
	s := Screen{ID: "test", Name: "test", Buttons: []Button{
		{Row: 0, Col: 0, Label: "A", Action: ActionText, Content: "hello"},
		{Row: 2, Col: 3, Label: "B", Action: ActionCopy},
	}}
	if b, ok := s.ButtonAt(0, 0); !ok || b.Content != "hello" {
		t.Fatalf("ButtonAt(0,0) = %v, %v", b, ok)
	}
	if _, ok := s.ButtonAt(1, 1); ok {
		t.Fatal("ButtonAt(1,1) should not exist")
	}
}

func TestConfigSaveWritesFileAndBackup(t *testing.T) {
	cfg, err := Load(writeFile(t, "radkeys.config.toml", sample))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	path := filepath.Join(t.TempDir(), "c.toml")
	// seed an existing file with comments so the .bak backup path is exercised
	if err := os.WriteFile(path, []byte("# my comments\n"), 0o600); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := cfg.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	// the rewritten file must be valid and reloadable
	cfg2, err := Load(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if len(cfg2.Screens) != len(cfg.Screens) {
		t.Fatalf("roundtrip screens: %d vs %d", len(cfg2.Screens), len(cfg.Screens))
	}
	// the backup must preserve the original commented content
	bak, err := os.ReadFile(path + ".bak")
	if err != nil {
		t.Fatalf("backup missing: %v", err)
	}
	if string(bak) != "# my comments\n" {
		t.Fatalf("backup = %q, want the original comments", bak)
	}
}

func TestLoadUnsupportedLanguage(t *testing.T) {
	body := `
[app]
language = "xx"
[app.device]
vendor_id = 1
product_id = 2
protocol = "radkeys-diy"
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 0
col = 0
label = "X"
action = "prev"
`
	_, err := Load(writeFile(t, "c.toml", body))
	if err == nil {
		t.Fatal("expected error for unsupported language")
	}
}

func TestLoadUnknownThemePreset(t *testing.T) {
	body := `
[app]
[app.theme]
preset = "nonexistent"
[app.device]
vendor_id = 1
product_id = 2
protocol = "radkeys-diy"
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 0
col = 0
label = "X"
action = "prev"
`
	_, err := Load(writeFile(t, "c.toml", body))
	if err == nil {
		t.Fatal("expected error for unknown theme preset")
	}
}

func TestLoadLayoutOutOfRange(t *testing.T) {
	body := `
[app]
[app.layout]
columns = 20
rows = 3
[app.device]
vendor_id = 1
product_id = 2
protocol = "radkeys-diy"
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 0
col = 0
label = "X"
action = "prev"
`
	_, err := Load(writeFile(t, "c.toml", body))
	if err == nil {
		t.Fatal("expected error for out-of-range layout columns")
	}
}

func TestLoadDuplicateButtonPosition(t *testing.T) {
	body := `
[app]
[app.layout]
columns = 4
rows = 3
[app.device]
vendor_id = 1
product_id = 2
protocol = "radkeys-diy"
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 1
col = 2
label = "A"
action = "prev"
[[screens.buttons]]
row = 1
col = 2
label = "B"
action = "prev"
`
	_, err := Load(writeFile(t, "c.toml", body))
	if err == nil {
		t.Fatal("expected error for duplicate button position")
	}
}
