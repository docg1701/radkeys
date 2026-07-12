// Package config loads and validates radkeys.config.toml.
package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

const (
	ProtocolElgato = "elgato"
	ProtocolDIY    = "radkeys-diy"
)

const (
	ActionText     = "text"
	ActionCopy     = "copy"
	ActionPaste    = "paste"
	ActionPrev     = "prev"
	ActionHome     = "home"
	ActionNavigate = "navigate"
)

// ValidActions is the set of all supported button actions.
var ValidActions = map[string]bool{
	ActionText:     true,
	ActionCopy:     true,
	ActionPaste:    true,
	ActionPrev:     true,
	ActionHome:     true,
	ActionNavigate: true,
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
	if err := c.validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Config) validate() error {
	if c.App.Device.Protocol != ProtocolElgato && c.App.Device.Protocol != ProtocolDIY {
		return fmt.Errorf(
			"[app.device] protocol must be %q or %q, got %q",
			ProtocolElgato, ProtocolDIY, c.App.Device.Protocol)
	}
	if c.App.Language == "" {
		c.App.Language = "en"
	}
	if c.App.Layout.Columns <= 0 || c.App.Layout.Columns > 6 {
		c.App.Layout.Columns = 4
	}
	if c.App.Layout.Rows <= 0 || c.App.Layout.Rows > 6 {
		c.App.Layout.Rows = 5
	}
	if len(c.Screens) == 0 {
		return fmt.Errorf("no screens defined — add at least one [[screens]]")
	}

	rows := c.App.Layout.Rows
	cols := c.App.Layout.Columns

	ids := map[string]struct{}{}
	for i, s := range c.Screens {
		if s.ID == "" {
			return fmt.Errorf("screen %d has empty id", i+1)
		}
		if _, dup := ids[s.ID]; dup {
			return fmt.Errorf("duplicate screen id %q", s.ID)
		}
		ids[s.ID] = struct{}{}
		if s.Name == "" {
			return fmt.Errorf("screen %q has empty name", s.ID)
		}
		for j, b := range s.Buttons {
			if b.Row < 0 || b.Row >= rows {
				return fmt.Errorf(
					"screen %q, button %d: row=%d out of range [0,%d)",
					s.ID, j+1, b.Row, rows)
			}
			if b.Col < 0 || b.Col >= cols {
				return fmt.Errorf(
					"screen %q, button %d: col=%d out of range [0,%d)",
					s.ID, j+1, b.Col, cols)
			}
			if !ValidActions[b.Action] {
				return fmt.Errorf(
					"screen %q, button %q: invalid action %q (use: text, copy, paste, prev, home, navigate)",
					s.ID, b.Label, b.Action)
			}
			if b.Action == ActionNavigate && b.Target == "" {
				return fmt.Errorf(
					"screen %q, button %q: navigate requires target",
					s.ID, b.Label)
			}
			if b.Action != ActionNavigate && b.Target != "" {
				return fmt.Errorf(
					"screen %q, button %q: action %q does not accept target",
					s.ID, b.Label, b.Action)
			}
			if b.Action == ActionText && b.Content == "" {
				return fmt.Errorf(
					"screen %q, button %q: text requires content",
					s.ID, b.Label)
			}
			if b.Action != ActionText && b.Content != "" {
				return fmt.Errorf(
					"screen %q, button %q: action %q does not accept content",
					s.ID, b.Label, b.Action)
			}
		}
	}
	// Validate navigate targets exist.
	for _, s := range c.Screens {
		for _, b := range s.Buttons {
			if b.Action == ActionNavigate {
				if _, ok := ids[b.Target]; !ok {
					return fmt.Errorf(
						"screen %q, button %q: target %q does not exist",
						s.ID, b.Label, b.Target)
				}
			}
		}
	}
	return nil
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

// ButtonAt returns the button at (row, col) for the screen, or (Button{}, false).
func (s Screen) ButtonAt(row, col int) (Button, bool) {
	for _, b := range s.Buttons {
		if b.Row == row && b.Col == col {
			return b, true
		}
	}
	return Button{}, false
}
