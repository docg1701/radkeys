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

[app.layout]
columns = 4
rows    = 3

[[layers]]
name = "Início"

[[layers.buttons]]
row = 0
col = 0
label = "RX"
action = "text"
content = "Tórax normal."

[[layers.buttons]]
row = 1
col = 0
label = "Próx"
action = "next"

[[layers.buttons]]
row = 2
col = 0
label = "Home"
action = "home"

[[layers.buttons]]
row = 2
col = 3
label = "Copy"
action = "copy"

[[layers]]
name = "RX Tórax"

[[layers.buttons]]
row = 0
col = 0
label = "Normal"
action = "text"
content = "Radiografia de tórax normal."

[[layers.buttons]]
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
	if cfg.App.Device.Protocol != ProtocolElgato {
		t.Fatalf("protocol = %q, want %q", cfg.App.Device.Protocol, ProtocolElgato)
	}
	if cfg.App.Device.VendorID != 0x0fd9 {
		t.Fatalf("vendor_id = %#x, want 0x0fd9", cfg.App.Device.VendorID)
	}
	if len(cfg.Layers) != 2 {
		t.Fatalf("layers = %d, want 2", len(cfg.Layers))
	}
	if cfg.Layers[0].Name != "Início" {
		t.Fatalf("layers[0].name = %q, want Início", cfg.Layers[0].Name)
	}

	// ButtonAt lookup.
	b, ok := cfg.Layers[0].ButtonAt(0, 0)
	if !ok {
		t.Fatal("ButtonAt(0,0) not found")
	}
	if b.Action != ActionText || b.Content != "Tórax normal." {
		t.Fatalf("button = %+v", b)
	}

	// Missing button.
	_, ok = cfg.Layers[0].ButtonAt(3, 3)
	if ok {
		t.Fatal("ButtonAt(3,3) should not exist")
	}
}

func TestLoadInvalidProtocol(t *testing.T) {
	body := `
[app]
[app.device]
vendor_id = 1
product_id = 2
protocol = "bogus"
[[layers]]
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
protocol = "elgato"
[[layers]]
name = "x"
[[layers.buttons]]
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
protocol = "elgato"
[[layers]]
name = "x"
[[layers.buttons]]
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
protocol = "elgato"
[[layers]]
name = "x"
[[layers.buttons]]
row = 0
col = 0
label = "X"
action = "navigate"
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
vendor_id = 1; product_id = 2; protocol = "elgato"
[app.layout]
columns = 4
rows = 3
[[layers]]
name = "x"
[[layers.buttons]]
row = 5
col = 0
label = "X"
action = "next"
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
vendor_id = 1; product_id = 2; protocol = "elgato"
[app.layout]
columns = 4
rows = 3
[[layers]]
name = "x"
[[layers.buttons]]
row = 0
col = 7
label = "X"
action = "next"
`
	_, err := Load(writeFile(t, "c.toml", body))
	if err == nil {
		t.Fatal("expected error for col out of range")
	}
}

func TestLoadNoLayers(t *testing.T) {
	body := `
[app]
[app.device]
vendor_id = 1; product_id = 2; protocol = "elgato"
`
	_, err := Load(writeFile(t, "c.toml", body))
	if err == nil {
		t.Fatal("expected error for zero layers")
	}
}

func TestLoadEmptyLayerName(t *testing.T) {
	body := `
[app]
[app.device]
vendor_id = 1; product_id = 2; protocol = "elgato"
[[layers]]
name = ""
`
	_, err := Load(writeFile(t, "c.toml", body))
	if err == nil {
		t.Fatal("expected error for empty layer name")
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
	if len(cfg2.Layers) != len(cfg.Layers) {
		t.Fatalf("layer count: %d vs %d", len(cfg2.Layers), len(cfg.Layers))
	}
	if cfg2.Layers[0].Name != cfg.Layers[0].Name {
		t.Fatalf("name mismatch: %q vs %q", cfg2.Layers[0].Name, cfg.Layers[0].Name)
	}
}

func TestButtonAt(t *testing.T) {
	layer := Layer{Name: "test", Buttons: []Button{
		{Row: 0, Col: 0, Label: "A", Action: ActionText, Content: "hello"},
		{Row: 2, Col: 3, Label: "B", Action: ActionCopy},
	}}
	if b, ok := layer.ButtonAt(0, 0); !ok || b.Content != "hello" {
		t.Fatalf("ButtonAt(0,0) = %v, %v", b, ok)
	}
	if _, ok := layer.ButtonAt(1, 1); ok {
		t.Fatal("ButtonAt(1,1) should not exist")
	}
}
