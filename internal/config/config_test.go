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
version = "0.0.0-test"

[app.device]
vendor_id  = 0x0fd9
product_id = 0x0063
protocol   = "elgato"

[app.fixed_buttons]
copy     = 0
level_up = 1
go_home  = 2

[[screens]]
id = "root"
title = "Início"
buttons = [
  { index = 3, label = "RX", action = "navigate", target = "rx_torax" },
  { index = 4, label = "Normal", action = "text", content = "Tórax normal." },
]

[[screens]]
id = "rx_torax"
title = "RX Tórax"
buttons = [
  { index = 3, label = "Normal", action = "text", content = "Radiografia de tórax normal." },
]
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
	if cfg.App.Device.Protocol != ProtocolElgato {
		t.Fatalf("protocol = %q, want %q", cfg.App.Device.Protocol, ProtocolElgato)
	}
	if cfg.App.Device.VendorID != 0x0fd9 {
		t.Fatalf("vendor_id = %#x, want 0x0fd9", cfg.App.Device.VendorID)
	}
	if cfg.App.FixedButtons.Copy != 0 || cfg.App.FixedButtons.LevelUp != 1 || cfg.App.FixedButtons.GoHome != 2 {
		t.Fatalf("fixed buttons = %+v", cfg.App.FixedButtons)
	}
	root, ok := cfg.ScreenByID("root")
	if !ok {
		t.Fatal("root screen not found")
	}
	if len(root.Buttons) != 2 {
		t.Fatalf("root buttons = %d, want 2", len(root.Buttons))
	}
	if !cfg.IsFixed(0) || !cfg.IsFixed(1) || !cfg.IsFixed(2) {
		t.Fatal("IsFixed should be true for 0,1,2")
	}
	if cfg.IsFixed(3) {
		t.Fatal("IsFixed should be false for 3")
	}
}

func TestLoadInvalidProtocol(t *testing.T) {
	body := `
[app]
[app.device]
vendor_id = 1
product_id = 2
protocol = "bogus"
[app.fixed_buttons]
copy = 0
level_up = 1
go_home = 2
[[screens]]
id = "root"
title = "x"
`
	_, err := Load(writeFile(t, "c.toml", body))
	if err == nil {
		t.Fatal("expected error for invalid protocol")
	}
}

func TestLoadNavigateUnknownTarget(t *testing.T) {
	body := `
[app]
[app.device]
vendor_id = 1
product_id = 2
protocol = "elgato"
[app.fixed_buttons]
copy = 0
level_up = 1
go_home = 2
[[screens]]
id = "root"
title = "x"
buttons = [ { index = 3, label = "RX", action = "navigate", target = "missing" } ]
`
	_, err := Load(writeFile(t, "c.toml", body))
	if err == nil {
		t.Fatal("expected error for navigate to unknown target")
	}
}

func TestLoadTextRequiresContent(t *testing.T) {
	body := `
[app]
[app.device]
vendor_id = 1
product_id = 2
protocol = "elgato"
[app.fixed_buttons]
copy = 0
level_up = 1
go_home = 2
[[screens]]
id = "root"
title = "x"
buttons = [ { index = 3, label = "X", action = "text", content = "" } ]
`
	_, err := Load(writeFile(t, "c.toml", body))
	if err == nil {
		t.Fatal("expected error for text without content")
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
	// Parse back.
	var cfg2 Config
	if err := toml.Unmarshal(buf.Bytes(), &cfg2); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if err := cfg2.validate(); err != nil {
		t.Fatalf("validate roundtripped: %v", err)
	}
	if cfg2.App.Device.Protocol != cfg.App.Device.Protocol {
		t.Fatalf("protocol mismatch: %q vs %q", cfg2.App.Device.Protocol, cfg.App.Device.Protocol)
	}
	if len(cfg2.Screens) != len(cfg.Screens) {
		t.Fatalf("screen count: %d vs %d", len(cfg2.Screens), len(cfg.Screens))
	}
}
