// Package config loads and validates radkeys.config.toml.
package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/docg1701/radkeys/internal/i18n"
	"github.com/docg1701/radkeys/internal/theme"
)

const (
	ProtocolDIY = "radkeys-diy"
)

const (
	ActionText       = "text"
	ActionCopy       = "copy"
	ActionPaste      = "paste"
	ActionPrev       = "prev"
	ActionHome       = "home"
	ActionNavigate   = "navigate"
	ActionSelectAll  = "select_all"
	ActionSelectLine = "select_line"
	ActionLineStart  = "line_start"
	ActionLineEnd    = "line_end"
	ActionBackspace  = "backspace"
	ActionDelete     = "delete"
	ActionExec       = "exec"
)

// IssueKind identifies a class of validation problem for machine translation.
type IssueKind string

const (
	IssueNoScreens              IssueKind = "no_screens"
	IssueInvalidProtocol        IssueKind = "invalid_protocol"
	IssueUnsupportedLanguage    IssueKind = "unsupported_language"
	IssueUnknownTheme           IssueKind = "unknown_theme"
	IssueColumnsOutOfRange      IssueKind = "columns_out_of_range"
	IssueRowsOutOfRange         IssueKind = "rows_out_of_range"
	IssueEmptyScreenID          IssueKind = "empty_screen_id"
	IssueDuplicateScreenID      IssueKind = "duplicate_screen_id"
	IssueEmptyScreenName        IssueKind = "empty_screen_name"
	IssueEmptyLabel             IssueKind = "empty_label"
	IssueOutOfGridRow           IssueKind = "out_of_grid_row"
	IssueOutOfGridCol           IssueKind = "out_of_grid_col"
	IssueDuplicatePosition      IssueKind = "duplicate_position"
	IssueInvalidAction          IssueKind = "invalid_action"
	IssueNavigateRequiresTarget IssueKind = "navigate_requires_target"
	IssueActionRejectsTarget    IssueKind = "action_rejects_target"
	IssueTextRequiresContent    IssueKind = "text_requires_content"
	IssueExecRequiresContent    IssueKind = "exec_requires_content"
	IssueActionRejectsContent   IssueKind = "action_rejects_content"
	IssueNavigateUnknownTarget  IssueKind = "navigate_unknown_target"
)

// Issue describes one validation problem in a Config.
type Issue struct {
	Kind     IssueKind
	ScreenID string
	Row      int
	Col      int
	Label    string
	Detail   string
}

// ActionList is the canonical ordered list of all button actions. It is the
// single source of truth for both validation and the editor's action picker.
var ActionList = []string{
	ActionText, ActionExec, ActionCopy, ActionPaste, ActionPrev, ActionHome,
	ActionNavigate, ActionSelectAll, ActionSelectLine, ActionLineStart,
	ActionLineEnd, ActionBackspace, ActionDelete,
}

// ActionLabel returns the localized label for an action id.
func ActionLabel(id string) string { return i18n.T("action." + id) }

// ActionLabels returns the localized labels for every action in ActionList,
// in the same order. The editor uses it to populate the action dropdown.
func ActionLabels() []string {
	out := make([]string, len(ActionList))
	for i, id := range ActionList {
		out[i] = i18n.T("action." + id)
	}
	return out
}

// ActionIDFromLabel maps a localized label back to the action id.
// Returns ActionText on miss.
func ActionIDFromLabel(label string) string {
	for _, id := range ActionList {
		if i18n.T("action."+id) == label {
			return id
		}
	}
	return ActionText
}

// Config is the root of radkeys.config.toml.
type Config struct {
	App     App      `toml:"app"`
	Screens []Screen `toml:"screens"`
}

// App holds app-wide settings.
type App struct {
	Name        string `toml:"name"`
	Radiologist string `toml:"radiologist"`
	Language    string `toml:"language"`
	Device      Device `toml:"device"`
	Layout      Layout `toml:"layout"`
	Theme       Theme  `toml:"theme"`
}

// Layout describes the physical keypad dimensions.
type Layout struct {
	Columns int `toml:"columns"` // grid columns (1–6)
	Rows    int `toml:"rows"`    // grid rows (1–6)
}

// Theme holds the selected preset and optional custom icon path.
type Theme struct {
	Preset string `toml:"preset"`
	Icon   string `toml:"icon"` // optional custom icon path
}

// Device identifies the USB HID custom device to open.
type Device struct {
	VendorID  uint16 `toml:"vendor_id"`
	ProductID uint16 `toml:"product_id"`
	Protocol  string `toml:"protocol"`
}

// Screen is one page of the shortcut deck with an ordered list of buttons.
type Screen struct {
	ID      string   `toml:"id"`
	Name    string   `toml:"name"`
	Buttons []Button `toml:"buttons"`
}

// Button maps a physical (row, col) to an action.
type Button struct {
	Row     int    `toml:"row"`               // 0-based
	Col     int    `toml:"col"`               // 0-based
	Label   string `toml:"label"`             // displayed on the button
	Action  string `toml:"action"`            // text | copy | paste | prev | home | navigate
	Target  string `toml:"target,omitempty"`  // screen id (only when action = "navigate")
	Content string `toml:"content,omitempty"` // report text (only when action = "text")
}

// Load reads, parses and validates the config file at path.
// Parse errors are wrapped with context so the user can fix the file.
func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}
	var c Config
	if err := toml.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("syntax error in %s:\n%w", path, err)
	}
	c.applyDefaults()
	if err := c.validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

// applyDefaults fills in omitted values with the product defaults.
// It must be called before validate() so validation is pure.
func (c *Config) applyDefaults() {
	if c.App.Language == "" {
		c.App.Language = "en"
	}
	if c.App.Theme.Preset == "" {
		c.App.Theme.Preset = "system"
	}
	if c.App.Layout.Columns == 0 {
		c.App.Layout.Columns = 6
	}
	if c.App.Layout.Rows == 0 {
		c.App.Layout.Rows = 6
	}
}

// Issues returns every validation problem in the config.
// The first issue is the error returned by Load/validate.
func (c *Config) Issues() []Issue {
	var issues []Issue
	if c.App.Device.Protocol != ProtocolDIY {
		issues = append(issues, Issue{Kind: IssueInvalidProtocol, Detail: c.App.Device.Protocol})
	}
	if !i18n.IsSupported(c.App.Language) {
		issues = append(issues, Issue{Kind: IssueUnsupportedLanguage, Detail: c.App.Language})
	}
	if _, ok := theme.FindPreset(c.App.Theme.Preset); !ok {
		issues = append(issues, Issue{Kind: IssueUnknownTheme, Detail: c.App.Theme.Preset})
	}
	if c.App.Layout.Columns < 1 || c.App.Layout.Columns > 6 {
		issues = append(issues, Issue{Kind: IssueColumnsOutOfRange, Detail: fmt.Sprintf("%d", c.App.Layout.Columns)})
	}
	if c.App.Layout.Rows < 1 || c.App.Layout.Rows > 6 {
		issues = append(issues, Issue{Kind: IssueRowsOutOfRange, Detail: fmt.Sprintf("%d", c.App.Layout.Rows)})
	}
	if len(c.Screens) == 0 {
		issues = append(issues, Issue{Kind: IssueNoScreens})
		return issues
	}

	rows := c.App.Layout.Rows
	cols := c.App.Layout.Columns

	ids := map[string]struct{}{}
	for _, s := range c.Screens {
		if s.ID == "" {
			issues = append(issues, Issue{Kind: IssueEmptyScreenID})
			continue
		}
		if _, dup := ids[s.ID]; dup {
			issues = append(issues, Issue{Kind: IssueDuplicateScreenID, ScreenID: s.ID})
			continue
		}
		ids[s.ID] = struct{}{}
		if s.Name == "" {
			issues = append(issues, Issue{Kind: IssueEmptyScreenName, ScreenID: s.ID})
		}
		occupied := map[[2]int]string{}
		for _, b := range s.Buttons {
			if b.Label == "" {
				issues = append(issues, Issue{Kind: IssueEmptyLabel, ScreenID: s.ID, Row: b.Row, Col: b.Col})
			}
			out := false
			if b.Row < 0 || b.Row >= rows {
				issues = append(issues, Issue{Kind: IssueOutOfGridRow, ScreenID: s.ID, Row: b.Row, Col: b.Col, Label: b.Label})
				out = true
			}
			if b.Col < 0 || b.Col >= cols {
				issues = append(issues, Issue{Kind: IssueOutOfGridCol, ScreenID: s.ID, Row: b.Row, Col: b.Col, Label: b.Label})
				out = true
			}
			if out {
				continue
			}
			pos := [2]int{b.Row, b.Col}
			if other, dup := occupied[pos]; dup {
				issues = append(issues, Issue{Kind: IssueDuplicatePosition, ScreenID: s.ID, Row: b.Row, Col: b.Col, Label: b.Label, Detail: other})
				continue
			}
			occupied[pos] = b.Label
			if !slices.Contains(ActionList, b.Action) {
				issues = append(issues, Issue{Kind: IssueInvalidAction, ScreenID: s.ID, Row: b.Row, Col: b.Col, Label: b.Label, Detail: b.Action})
				continue
			}
			if b.Action == ActionNavigate && b.Target == "" {
				issues = append(issues, Issue{Kind: IssueNavigateRequiresTarget, ScreenID: s.ID, Row: b.Row, Col: b.Col, Label: b.Label})
			}
			if b.Action != ActionNavigate && b.Target != "" {
				issues = append(issues, Issue{Kind: IssueActionRejectsTarget, ScreenID: s.ID, Row: b.Row, Col: b.Col, Label: b.Label, Detail: b.Action})
			}
			if (b.Action == ActionText || b.Action == ActionExec) && b.Content == "" {
				kind := IssueTextRequiresContent
				if b.Action == ActionExec {
					kind = IssueExecRequiresContent
				}
				issues = append(issues, Issue{Kind: kind, ScreenID: s.ID, Row: b.Row, Col: b.Col, Label: b.Label})
			}
			if b.Action != ActionText && b.Action != ActionExec && b.Content != "" {
				issues = append(issues, Issue{Kind: IssueActionRejectsContent, ScreenID: s.ID, Row: b.Row, Col: b.Col, Label: b.Label, Detail: b.Action})
			}
		}
	}
	for _, s := range c.Screens {
		for _, b := range s.Buttons {
			if b.Action == ActionNavigate {
				if _, ok := ids[b.Target]; !ok {
					issues = append(issues, Issue{Kind: IssueNavigateUnknownTarget, ScreenID: s.ID, Row: b.Row, Col: b.Col, Label: b.Label, Detail: b.Target})
				}
			}
		}
	}
	return issues
}

func (c *Config) validate() error {
	if issues := c.Issues(); len(issues) > 0 {
		return issues[0].Error(c.App.Layout.Rows, c.App.Layout.Columns)
	}
	return nil
}

// Error formats an Issue as a human-readable error by looking up its
// formatter in the issueFormatters table.
func (issue Issue) Error(rows, cols int) error {
	formatter, ok := issueFormatters[issue.Kind]
	if !ok {
		return fmt.Errorf("%v", issue)
	}
	return formatter(issue, rows, cols)
}

// issueFormatter formats an Issue into an error, receiving the grid dimensions
// needed for position-related messages.
type issueFormatter func(issue Issue, rows, cols int) error

// issueFormatters maps every IssueKind to its formatter. A single table
// replaces the former nested appError/layoutError/screenError/buttonError
// switch chain.
var issueFormatters = map[IssueKind]issueFormatter{
	IssueInvalidProtocol:        formatInvalidProtocol,
	IssueUnsupportedLanguage:    formatUnsupportedLanguage,
	IssueUnknownTheme:           formatUnknownTheme,
	IssueColumnsOutOfRange:      formatColumnsOutOfRange,
	IssueRowsOutOfRange:         formatRowsOutOfRange,
	IssueNoScreens:              formatNoScreens,
	IssueEmptyScreenID:          formatEmptyScreenID,
	IssueDuplicateScreenID:      formatDuplicateScreenID,
	IssueEmptyScreenName:        formatEmptyScreenName,
	IssueEmptyLabel:             formatEmptyLabel,
	IssueOutOfGridRow:           formatOutOfGridRow,
	IssueOutOfGridCol:           formatOutOfGridCol,
	IssueDuplicatePosition:      formatDuplicatePosition,
	IssueInvalidAction:          formatInvalidAction,
	IssueNavigateRequiresTarget: formatNavigateRequiresTarget,
	IssueActionRejectsTarget:    formatActionRejectsTarget,
	IssueTextRequiresContent:    formatTextRequiresContent,
	IssueExecRequiresContent:    formatExecRequiresContent,
	IssueActionRejectsContent:   formatActionRejectsContent,
	IssueNavigateUnknownTarget:  formatNavigateUnknownTarget,
}

func formatInvalidProtocol(issue Issue, _, _ int) error {
	return fmt.Errorf("[app.device] protocol must be %q, got %q", ProtocolDIY, issue.Detail)
}

func formatUnsupportedLanguage(issue Issue, _, _ int) error {
	return fmt.Errorf("[app] language %q is not supported (use one of: %s)", issue.Detail, strings.Join(i18n.Supported, ", "))
}

func formatUnknownTheme(issue Issue, _, _ int) error {
	ids := make([]string, len(theme.Presets))
	for i, p := range theme.Presets {
		ids[i] = p.ID()
	}
	return fmt.Errorf("[app.theme] preset %q is unknown (use one of: %s)", issue.Detail, strings.Join(ids, ", "))
}

func formatColumnsOutOfRange(issue Issue, _, _ int) error {
	return fmt.Errorf("[app.layout] columns=%s out of range [1,6]", issue.Detail)
}

func formatRowsOutOfRange(issue Issue, _, _ int) error {
	return fmt.Errorf("[app.layout] rows=%s out of range [1,6]", issue.Detail)
}

func formatNoScreens(_ Issue, _, _ int) error {
	return errors.New("no screens defined — add at least one [[screens]]")
}

func formatEmptyScreenID(issue Issue, _, _ int) error {
	return errors.New("screen has empty id")
}

func formatDuplicateScreenID(issue Issue, _, _ int) error {
	return fmt.Errorf("duplicate screen id %q", issue.ScreenID)
}

func formatEmptyScreenName(issue Issue, _, _ int) error {
	return fmt.Errorf("screen %q has empty name", issue.ScreenID)
}

func formatEmptyLabel(issue Issue, _, _ int) error {
	return fmt.Errorf("screen %q, button at (row=%d, col=%d): label is required", issue.ScreenID, issue.Row, issue.Col)
}

func formatOutOfGridRow(issue Issue, rows, _ int) error {
	return fmt.Errorf("screen %q, button %q: row=%d out of range [0,%d)", issue.ScreenID, issue.Label, issue.Row, rows)
}

func formatOutOfGridCol(issue Issue, _, cols int) error {
	return fmt.Errorf("screen %q, button %q: col=%d out of range [0,%d)", issue.ScreenID, issue.Label, issue.Col, cols)
}

func formatDuplicatePosition(issue Issue, _, _ int) error {
	return fmt.Errorf("screen %q: buttons %q and %q both occupy (row=%d, col=%d)", issue.ScreenID, issue.Detail, issue.Label, issue.Row, issue.Col)
}

func formatInvalidAction(issue Issue, _, _ int) error {
	return fmt.Errorf("screen %q, button %q: invalid action %q (use: text, copy, paste, prev, home, navigate, select_all, select_line, line_start, line_end, backspace, delete, exec)", issue.ScreenID, issue.Label, issue.Detail)
}

func formatNavigateRequiresTarget(issue Issue, _, _ int) error {
	return fmt.Errorf("screen %q, button %q: navigate requires target", issue.ScreenID, issue.Label)
}

func formatActionRejectsTarget(issue Issue, _, _ int) error {
	return fmt.Errorf("screen %q, button %q: action %q does not accept target", issue.ScreenID, issue.Label, issue.Detail)
}

func formatTextRequiresContent(issue Issue, _, _ int) error {
	return fmt.Errorf("screen %q, button %q: text requires content", issue.ScreenID, issue.Label)
}

func formatExecRequiresContent(issue Issue, _, _ int) error {
	return fmt.Errorf("screen %q, button %q: exec requires content", issue.ScreenID, issue.Label)
}

func formatActionRejectsContent(issue Issue, _, _ int) error {
	return fmt.Errorf("screen %q, button %q: action %q does not accept content", issue.ScreenID, issue.Label, issue.Detail)
}

func formatNavigateUnknownTarget(issue Issue, _, _ int) error {
	return fmt.Errorf("screen %q, button %q: target %q does not exist", issue.ScreenID, issue.Label, issue.Detail)
}

// ScreenByID returns the screen with the given id.
func (c *Config) ScreenByID(id string) (Screen, bool) {
	for _, s := range c.Screens {
		if s.ID == id {
			return s, true
		}
	}
	return Screen{}, false
}

// Save writes the config to path as TOML. The existing file is first copied to
// path+".bak" because BurntSushi/toml does not preserve comments on encode,
// so the user's commented master survives in the backup.
func (c *Config) Save(path string) error {
	if existing, err := os.ReadFile(path); err == nil {
		if err := os.WriteFile(path+".bak", existing, 0o600); err != nil {
			log.Printf("radkeys: cannot write backup %s.bak: %v", path, err)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("save: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("radkeys: config close failed: %v", err)
		}
	}()
	if err := toml.NewEncoder(f).Encode(c); err != nil {
		return fmt.Errorf("TOML: %w", err)
	}
	return nil
}

// ButtonAt returns the button at (row, col) for the screen, or (Button{}, false).
func (s Screen) ButtonAt(row, col int) (Button, bool) {
	for _, b := range s.Buttons {
		if b.Row == row && b.Col == col {
			return b, true
		}
	}
	return Button{}, false
}

// ButtonIndex returns the index of the button at (row, col), or (-1, false).
// Mirrors ButtonAt but returns the slice position so callers can mutate.
func (s Screen) ButtonIndex(row, col int) (int, bool) {
	for i, b := range s.Buttons {
		if b.Row == row && b.Col == col {
			return i, true
		}
	}
	return -1, false
}

// DropdownLabel formats the screen for picker UIs as "id — name".
func (s Screen) DropdownLabel() string {
	return s.ID + " — " + s.Name
}

// ParseHexUint16 parses a hexadecimal string as a 16-bit unsigned integer.
// A leading "0x" or "0X" is stripped before parsing.
func ParseHexUint16(s string) (uint16, error) {
	v, err := strconv.ParseUint(strings.TrimPrefix(strings.TrimPrefix(s, "0x"), "0X"), 16, 16)
	if err != nil {
		return 0, err
	}
	return uint16(v), nil
}

// DefaultConfigFile is the file name the host looks for next to the binary
// and in the working directory when no --config flag or env var is set.
const DefaultConfigFile = "radkeys.config.toml"

// StartupPath resolves the config file path the host should open at launch.
// Order: $RADKEYS_CONFIG, <exe-dir>/DefaultConfigFile, DefaultConfigFile.
func StartupPath() string {
	if p := os.Getenv("RADKEYS_CONFIG"); p != "" {
		return p
	}
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), DefaultConfigFile)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return DefaultConfigFile
}

// DefaultConfig returns a blank starter config with one empty screen.
func DefaultConfig() *Config {
	return &Config{
		App: App{
			Name:     "RadKeys",
			Language: "en",
			Device:   Device{VendorID: 0x1234, ProductID: 0xABCD, Protocol: ProtocolDIY},
			Layout:   Layout{Columns: 6, Rows: 6},
			Theme:    Theme{Preset: "system"},
		},
		Screens: []Screen{{ID: "root", Name: "Home"}},
	}
}

// LoadStartup loads the config at path or returns a fresh default and a
// non-nil error describing the failure. The host binary uses this at boot.
func LoadStartup(path string) (*Config, error) {
	if _, err := os.Stat(path); err != nil {
		return DefaultConfig(), fmt.Errorf("no config found at %s", path)
	}
	cfg, err := Load(path)
	if err != nil {
		return DefaultConfig(), err
	}
	return cfg, nil
}
