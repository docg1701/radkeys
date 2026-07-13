package editor

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/docg1701/radkeys/internal/config"
)

const repoFixture = "../../radkeys.config.toml"

func TestRoundtripViaConfigSave(t *testing.T) {
	cfg, err := config.Load(repoFixture)
	if err != nil {
		t.Fatalf("Load fixture: %v", err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "out.toml")
	if err := cfg.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	reloaded, err := config.Load(path)
	if err != nil {
		t.Fatalf("Reload saved: %v", err)
	}
	if !reflect.DeepEqual(reloaded.App, cfg.App) {
		t.Fatalf("app block changed: got %+v, want %+v", reloaded.App, cfg.App)
	}
	if len(reloaded.Screens) != len(cfg.Screens) {
		t.Fatalf("screen count: got %d, want %d", len(reloaded.Screens), len(cfg.Screens))
	}
	for i, want := range cfg.Screens {
		got := reloaded.Screens[i]
		if got.ID != want.ID || got.Name != want.Name {
			t.Fatalf("screen %d: got %+v, want %+v", i, got, want)
		}
		if len(got.Buttons) != len(want.Buttons) {
			t.Fatalf("screen %d buttons: got %d, want %d", i, len(got.Buttons), len(want.Buttons))
		}
		for j, bw := range want.Buttons {
			bg := got.Buttons[j]
			if !reflect.DeepEqual(bg, bw) {
				t.Fatalf("screen %d button %d: got %+v, want %+v", i, j, bg, bw)
			}
		}
	}
}

func TestConfigIssuesSurfacesButtonProblems(t *testing.T) {
	assertIssueKinds(t, invalidButtonsConfig().Issues(), []config.IssueKind{
		config.IssueEmptyLabel,
		config.IssueDuplicatePosition,
		config.IssueOutOfGridRow,
		config.IssueOutOfGridCol,
		config.IssueNavigateUnknownTarget,
		config.IssueTextRequiresContent,
	})
}

func invalidButtonsConfig() *config.Config {
	return &config.Config{
		App: config.App{
			Language: "en",
			Theme:    config.Theme{Preset: "system"},
			Device:   config.Device{VendorID: 0x1234, ProductID: 0xABCD, Protocol: config.ProtocolDIY},
			Layout:   config.Layout{Columns: 4, Rows: 4},
		},
		Screens: []config.Screen{
			{
				ID:   "root",
				Name: "Root",
				Buttons: []config.Button{
					{Row: 0, Col: 0, Label: "", Action: config.ActionText, Content: "has content"},
					{Row: 0, Col: 0, Label: "DupA", Action: config.ActionCopy},
					{Row: 0, Col: 0, Label: "DupB", Action: config.ActionPaste},
					{Row: 5, Col: 5, Label: "Far", Action: config.ActionPrev},
					{Row: 1, Col: 1, Label: "BadNav", Action: config.ActionNavigate, Target: "missing"},
					{Row: 2, Col: 2, Label: "NoContent", Action: config.ActionText},
				},
			},
		},
	}
}

func assertIssueKinds(t *testing.T, issues []config.Issue, want []config.IssueKind) {
	t.Helper()
	got := make(map[config.IssueKind]bool, len(issues))
	for _, issue := range issues {
		got[issue.Kind] = true
	}
	for _, kind := range want {
		if !got[kind] {
			t.Errorf("missing issue kind %q; got %v", kind, got)
		}
	}
}

func TestResizeIsNonDestructive(t *testing.T) {
	cfg := &config.Config{
		App: config.App{
			Language: "en",
			Theme:    config.Theme{Preset: "system"},
			Device:   config.Device{VendorID: 0x1234, ProductID: 0xABCD, Protocol: config.ProtocolDIY},
			Layout:   config.Layout{Columns: 6, Rows: 6},
		},
		Screens: []config.Screen{
			{
				ID:      "root",
				Name:    "Root",
				Buttons: []config.Button{{Row: 5, Col: 5, Label: "Far", Action: config.ActionCopy}},
			},
		},
	}
	if len(cfg.Issues()) != 0 {
		t.Fatalf("initial config invalid: %v", cfg.Issues())
	}
	cfg.App.Layout.Columns = 4
	cfg.App.Layout.Rows = 4
	issues := cfg.Issues()
	if len(issues) == 0 {
		t.Fatal("expected out-of-grid issue after shrink")
	}
	found := false
	for _, issue := range issues {
		if issue.Kind == config.IssueOutOfGridRow || issue.Kind == config.IssueOutOfGridCol {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected out-of-grid issue, got %v", issues)
	}
	if len(cfg.Screens[0].Buttons) != 1 {
		t.Fatalf("button was removed by resize")
	}
	cfg.App.Layout.Columns = 6
	cfg.App.Layout.Rows = 6
	if len(cfg.Issues()) != 0 {
		t.Fatalf("config should be valid after restore: %v", cfg.Issues())
	}
}

func TestDefaultConfigHasOneScreen(t *testing.T) {
	cfg := defaultConfig()
	if len(cfg.Screens) != 1 {
		t.Fatalf("screens = %d, want 1", len(cfg.Screens))
	}
	if cfg.Screens[0].ID != "root" {
		t.Fatalf("first screen id = %q, want root", cfg.Screens[0].ID)
	}
}

func TestStartupPathUsesExecutableDir(t *testing.T) {
	// Fallback returns the relative filename when no executable-dir config exists.
	if got := StartupPath(); got != "radkeys.config.toml" {
		t.Fatalf("StartupPath fallback = %q, want radkeys.config.toml", got)
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
