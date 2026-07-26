package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
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
[[app.layout.blocks]]
rows = 2
cols = 4
[[app.layout.blocks]]
rows = 3
cols = 4

[[screens]]
id = "root"
name = "Início"

[[screens.buttons]]
block = 1
row = 0
col = 0
label = "RX"
action = "navigate"
target = "rx_torax"

[[screens.buttons]]
block = 1
row = 1
col = 0
label = "Voltar"
action = "prev"

[[screens.buttons]]
block = 0
row = 0
col = 0
label = "Home"
action = "home"

[[screens.buttons]]
block = 0
row = 1
col = 3
label = "Copy"
action = "copy"

[[screens]]
id = "rx_torax"
name = "RX Tórax"

[[screens.buttons]]
block = 1
row = 0
col = 0
label = "Normal"
action = "text"
content = "Radiografia de tórax normal."

[[screens.buttons]]
block = 1
row = 1
col = 0
label = "Voltar"
action = "prev"
`

// base is a minimal valid header (device + one 6×6 block) for fixtures that
// only exercise screen/button rules.
const base = `
[app]
[app.device]
vendor_id = 1
product_id = 2
protocol = "radkeys-diy"
[app.layout]
[[app.layout.blocks]]
rows = 6
cols = 6
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
	b, ok := root.ButtonAt(1, 0, 0)
	if !ok {
		t.Fatal("ButtonAt(1,0,0) not found")
	}
	if b.Action != ActionNavigate || b.Target != "rx_torax" {
		t.Fatalf("button = %+v", b)
	}
}

func TestLoadInvalidProtocol(t *testing.T) {
	_, err := Load(writeFile(t, "c.toml", `
[app]
[app.device]
vendor_id = 1
product_id = 2
protocol = "bogus"
[[screens]]
id = "root"
name = "x"
`))
	if err == nil {
		t.Fatal("expected error for invalid protocol")
	}
}

func TestLoadTextRequiresContent(t *testing.T) {
	_, err := Load(writeFile(t, "c.toml", base+`
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 0
col = 0
label = "X"
action = "text"
content = ""
`))
	if err == nil {
		t.Fatal("expected error for text without content")
	}
}

func TestLoadActionMustNotHaveContent(t *testing.T) {
	_, err := Load(writeFile(t, "c.toml", base+`
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 0
col = 0
label = "X"
action = "copy"
content = "nope"
`))
	if err == nil {
		t.Fatal("expected error for non-text action with content")
	}
}

func TestLoadInvalidAction(t *testing.T) {
	_, err := Load(writeFile(t, "c.toml", base+`
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 0
col = 0
label = "X"
action = "next"
`))
	if err == nil {
		t.Fatal("expected error for invalid action")
	}
}

func TestLoadRowOutOfRange(t *testing.T) {
	_, err := Load(writeFile(t, "c.toml", `
[app]
[app.device]
vendor_id = 1
product_id = 2
protocol = "radkeys-diy"
[app.layout]
[[app.layout.blocks]]
rows = 3
cols = 4
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 3
col = 0
label = "X"
action = "prev"
`))
	if err == nil {
		t.Fatal("expected error for row out of range")
	}
}

func TestLoadColOutOfRange(t *testing.T) {
	_, err := Load(writeFile(t, "c.toml", `
[app]
[app.device]
vendor_id = 1
product_id = 2
protocol = "radkeys-diy"
[app.layout]
[[app.layout.blocks]]
rows = 3
cols = 4
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 0
col = 4
label = "X"
action = "prev"
`))
	if err == nil {
		t.Fatal("expected error for col out of range")
	}
}

func TestLoadNoBlocks(t *testing.T) {
	_, err := Load(writeFile(t, "c.toml", `
[app]
[app.device]
vendor_id = 1
product_id = 2
protocol = "radkeys-diy"
[[screens]]
id = "root"
name = "x"
`))
	if err == nil {
		t.Fatal("expected error for layout without blocks")
	}
}

func TestLoadBlockDimOutOfRange(t *testing.T) {
	_, err := Load(writeFile(t, "c.toml", `
[app]
[app.device]
vendor_id = 1
product_id = 2
protocol = "radkeys-diy"
[app.layout]
[[app.layout.blocks]]
rows = 0
cols = 6
[[screens]]
id = "root"
name = "x"
`))
	if err == nil {
		t.Fatal("expected error for block with rows=0")
	}
}

func TestLoadTooManySlots(t *testing.T) {
	_, err := Load(writeFile(t, "c.toml", `
[app]
[app.device]
vendor_id = 1
product_id = 2
protocol = "radkeys-diy"
[app.layout]
[[app.layout.blocks]]
rows = 6
cols = 6
[[app.layout.blocks]]
rows = 1
cols = 1
[[screens]]
id = "root"
name = "x"
`))
	if err == nil {
		t.Fatal("expected error for blocks totaling more than 36 slots")
	}
}

func TestLoadUnknownBlock(t *testing.T) {
	_, err := Load(writeFile(t, "c.toml", base+`
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
block = 1
row = 0
col = 0
label = "X"
action = "prev"
`))
	if err == nil {
		t.Fatal("expected error for button pointing at a block that does not exist")
	}
}

func TestLoadAppliesDefaultsOmittedLanguageAndTheme(t *testing.T) {
	cfg, err := Load(writeFile(t, "c.toml", base+`
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 0
col = 0
label = "X"
action = "prev"
`))
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
			Layout:   Layout{Blocks: []Block{{Rows: 4, Cols: 5}}},
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
		!reflect.DeepEqual(cfg.App.Layout, want.App.Layout) {
		t.Fatalf("validate mutated config: got %+v, want %+v", cfg, want)
	}
}

func TestLoadNoScreens(t *testing.T) {
	_, err := Load(writeFile(t, "c.toml", base))
	if err == nil {
		t.Fatal("expected error for zero screens")
	}
}

func TestLoadEmptyScreenID(t *testing.T) {
	_, err := Load(writeFile(t, "c.toml", base+`
[[screens]]
id = ""
name = "x"
`))
	if err == nil {
		t.Fatal("expected error for empty screen id")
	}
}

func TestLoadNavigateRequiresTarget(t *testing.T) {
	_, err := Load(writeFile(t, "c.toml", base+`
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 0
col = 0
label = "X"
action = "navigate"
`))
	if err == nil {
		t.Fatal("expected error for navigate without target")
	}
}

func TestLoadNavigateUnknownTarget(t *testing.T) {
	_, err := Load(writeFile(t, "c.toml", base+`
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 0
col = 0
label = "X"
action = "navigate"
target = "missing"
`))
	if err == nil {
		t.Fatal("expected error for navigate to unknown target")
	}
}

func TestLoadDuplicateScreenID(t *testing.T) {
	_, err := Load(writeFile(t, "c.toml", base+`
[[screens]]
id = "root"
name = "x"
[[screens]]
id = "root"
name = "y"
`))
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
		{Block: 0, Row: 0, Col: 0, Label: "A", Action: ActionText, Content: "hello"},
		{Block: 1, Row: 2, Col: 3, Label: "B", Action: ActionCopy},
	}}
	if b, ok := s.ButtonAt(0, 0, 0); !ok || b.Content != "hello" {
		t.Fatalf("ButtonAt(0,0,0) = %v, %v", b, ok)
	}
	if b, ok := s.ButtonAt(1, 2, 3); !ok || b.Label != "B" {
		t.Fatalf("ButtonAt(1,2,3) = %v, %v", b, ok)
	}
	if _, ok := s.ButtonAt(0, 2, 3); ok {
		t.Fatal("ButtonAt(0,2,3) should not exist — block differs")
	}
}

func TestSlotMath(t *testing.T) {
	l := Layout{Blocks: []Block{{Rows: 2, Cols: 4}, {Rows: 4, Cols: 6}}}
	if got := l.SlotCount(); got != 32 {
		t.Fatalf("SlotCount = %d, want 32", got)
	}
	cases := []struct {
		block, row, col, slot int
	}{
		{0, 0, 0, 0},
		{0, 1, 3, 7},
		{1, 0, 0, 8},
		{1, 3, 5, 31},
	}
	for _, tc := range cases {
		if got := l.SlotOf(tc.block, tc.row, tc.col); got != tc.slot {
			t.Errorf("SlotOf(%d,%d,%d) = %d, want %d", tc.block, tc.row, tc.col, got, tc.slot)
		}
		b, r, c, ok := l.LocateSlot(tc.slot)
		if !ok || b != tc.block || r != tc.row || c != tc.col {
			t.Errorf("LocateSlot(%d) = (%d,%d,%d,%v), want (%d,%d,%d,true)",
				tc.slot, b, r, c, ok, tc.block, tc.row, tc.col)
		}
	}
	if _, _, _, ok := l.LocateSlot(32); ok {
		t.Error("LocateSlot(32) should be unassigned")
	}
	if _, _, _, ok := l.LocateSlot(-1); ok {
		t.Error("LocateSlot(-1) should be unassigned")
	}
	if got := MatrixSlot(0, 0); got != 0 {
		t.Errorf("MatrixSlot(0,0) = %d, want 0", got)
	}
	if got := MatrixSlot(5, 5); got != 35 {
		t.Errorf("MatrixSlot(5,5) = %d, want 35", got)
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
	_, err := Load(writeFile(t, "c.toml", `
[app]
language = "xx"
[app.device]
vendor_id = 1
product_id = 2
protocol = "radkeys-diy"
[app.layout]
[[app.layout.blocks]]
rows = 6
cols = 6
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 0
col = 0
label = "X"
action = "prev"
`))
	if err == nil {
		t.Fatal("expected error for unsupported language")
	}
}

func TestLoadUnknownThemePreset(t *testing.T) {
	_, err := Load(writeFile(t, "c.toml", `
[app]
[app.theme]
preset = "nonexistent"
[app.device]
vendor_id = 1
product_id = 2
protocol = "radkeys-diy"
[app.layout]
[[app.layout.blocks]]
rows = 6
cols = 6
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 0
col = 0
label = "X"
action = "prev"
`))
	if err == nil {
		t.Fatal("expected error for unknown theme preset")
	}
}

func TestLoadDuplicateButtonPosition(t *testing.T) {
	_, err := Load(writeFile(t, "c.toml", base+`
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
`))
	if err == nil {
		t.Fatal("expected error for duplicate button position")
	}
}

func TestLoadNewActionsAccepted(t *testing.T) {
	actions := []string{
		ActionSelectAll, ActionSelectLine, ActionLineStart,
		ActionLineEnd, ActionBackspace, ActionDelete,
	}
	for _, a := range actions {
		body := fmt.Sprintf(base+`
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 0
col = 0
label = "X"
action = %q
`, a)
		if _, err := Load(writeFile(t, "c.toml", body)); err != nil {
			t.Fatalf("action %q: %v", a, err)
		}
	}
}

func TestLoadNewActionRejectsContent(t *testing.T) {
	_, err := Load(writeFile(t, "c.toml", base+`
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 0
col = 0
label = "X"
action = "backspace"
content = "nope"
`))
	if err == nil {
		t.Fatal("expected error for backspace with content")
	}
}

func TestLoadNewActionRejectsTarget(t *testing.T) {
	_, err := Load(writeFile(t, "c.toml", base+`
[[screens]]
id = "root"
name = "x"
[[screens.buttons]]
row = 0
col = 0
label = "X"
action = "delete"
target = "root"
`))
	if err == nil {
		t.Fatal("expected error for delete with target")
	}
}

func TestIssueErrorPerKind(t *testing.T) {
	layout := Layout{Blocks: []Block{{Rows: 6, Cols: 6}}}
	kinds := []IssueKind{
		IssueInvalidProtocol,
		IssueUnsupportedLanguage,
		IssueUnknownTheme,
		IssueNoBlocks,
		IssueBlockDimOutOfRange,
		IssueTooManySlots,
		IssueUnknownBlock,
		IssueNoScreens,
		IssueEmptyScreenID,
		IssueDuplicateScreenID,
		IssueEmptyScreenName,
		IssueEmptyLabel,
		IssueOutOfGridRow,
		IssueOutOfGridCol,
		IssueDuplicatePosition,
		IssueInvalidAction,
		IssueNavigateRequiresTarget,
		IssueActionRejectsTarget,
		IssueTextRequiresContent,
		IssueExecRequiresContent,
		IssueActionRejectsContent,
		IssueNavigateUnknownTarget,
	}
	for _, kind := range kinds {
		issue := Issue{Kind: kind, ScreenID: "root", Block: 0, Row: 1, Col: 2, Label: "X", Detail: "detail"}
		err := issue.Error(layout)
		if err == nil {
			t.Errorf("Issue.Error(%q) = nil, want error", kind)
			continue
		}
		if err.Error() == "" {
			t.Errorf("Issue.Error(%q) returned empty message", kind)
		}
	}

	// Unknown kind falls back to a non-empty default message.
	unknown := Issue{Kind: "totally_unknown", ScreenID: "root", Row: 0, Col: 0}
	if got := unknown.Error(layout).Error(); got == "" {
		t.Fatal("unknown kind should still produce a message")
	}
}

func TestDefaultConfigHasOneScreen(t *testing.T) {
	cfg := DefaultConfig()
	if len(cfg.Screens) != 1 {
		t.Fatalf("screens = %d, want 1", len(cfg.Screens))
	}
	if cfg.Screens[0].ID != "root" {
		t.Fatalf("first screen id = %q, want root", cfg.Screens[0].ID)
	}
	if err := cfg.validate(); err != nil {
		t.Fatalf("default config should be valid: %v", err)
	}
}

func TestStartupPathUsesExecutableDir(t *testing.T) {
	// Clear the env override so the fallback path is deterministic.
	t.Setenv("RADKEYS_CONFIG", "")
	// Fallback returns the relative filename when no executable-dir config exists.
	if got := StartupPath(); got != "radkeys.config.toml" {
		t.Fatalf("StartupPath fallback = %q, want radkeys.config.toml", got)
	}
}

func TestStartupPathHonorsEnvOverride(t *testing.T) {
	t.Setenv("RADKEYS_CONFIG", "/tmp/radkeys-test-override.toml")
	if got := StartupPath(); got != "/tmp/radkeys-test-override.toml" {
		t.Fatalf("StartupPath = %q, want /tmp/radkeys-test-override.toml", got)
	}
}

func TestLoadStartupReturnsDefaultWhenMissing(t *testing.T) {
	cfg, err := LoadStartup(filepath.Join(t.TempDir(), "missing.toml"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if len(cfg.Screens) != 1 {
		t.Fatalf("default config should have one screen, got %d", len(cfg.Screens))
	}
}

func TestLoadStartupReadsExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "radkeys.config.toml")
	body := `[app]
[app.device]
vendor_id = 0x1234
product_id = 0xABCD
protocol = "radkeys-diy"
[app.layout]
[[app.layout.blocks]]
rows = 6
cols = 6
[[screens]]
id = "root"
name = "Home"
`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	cfg, err := LoadStartup(path)
	if err != nil {
		t.Fatalf("LoadStartup: %v", err)
	}
	if cfg.Screens[0].Name != "Home" {
		t.Fatalf("screen name = %q, want Home", cfg.Screens[0].Name)
	}
}

func TestParseHexUint16Valid(t *testing.T) {
	for _, in := range []string{"1234", "0x1234", "0X1234", "ABCD"} {
		v, err := ParseHexUint16(in)
		if err != nil {
			t.Fatalf("ParseHexUint16(%q) unexpected error: %v", in, err)
		}
		want := uint16(0x1234)
		if in == "ABCD" {
			want = 0xABCD
		}
		if v != want {
			t.Fatalf("ParseHexUint16(%q) = 0x%04x, want 0x%04x", in, v, want)
		}
	}
}

func TestParseHexUint16Invalid(t *testing.T) {
	_, err := ParseHexUint16("xyz")
	if err == nil {
		t.Fatal("ParseHexUint16(xyz) expected error, got nil")
	}
}

func TestParseHexUint16Overflow(t *testing.T) {
	_, err := ParseHexUint16("12345")
	if err == nil {
		t.Fatal("ParseHexUint16(12345) expected error, got nil")
	}
}
