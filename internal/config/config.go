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
	ActionText  = "text"
	ActionCopy  = "copy"
	ActionPaste = "paste"
	ActionPrev  = "prev"
	ActionNext  = "next"
	ActionHome  = "home"
)

// ValidActions is the set of all supported button actions.
var ValidActions = map[string]bool{
	ActionText:  true,
	ActionCopy:  true,
	ActionPaste: true,
	ActionPrev:  true,
	ActionNext:  true,
	ActionHome:  true,
}

// Config is the root of radkeys.config.toml.
type Config struct {
	App    App     `toml:"app"`
	Layers []Layer `toml:"layers"`
}

// App holds app-wide settings.
type App struct {
	Name        string `toml:"name"`
	Radiologist string `toml:"radiologist"`
	Language    string `toml:"language"`
	Version     string `toml:"version"`
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

// Layer is one page of shortcuts with an ordered list of buttons.
type Layer struct {
	Name    string   `toml:"name"`
	Buttons []Button `toml:"buttons"`
}

// Button maps a physical (row, col) to an action.
type Button struct {
	Row     int    `toml:"row"`               // 0-based
	Col     int    `toml:"col"`               // 0-based
	Label   string `toml:"label"`             // displayed on the button
	Action  string `toml:"action"`            // text | copy | paste | prev | next | home
	Content string `toml:"content,omitempty"` // only when action = "text"
}

// Load reads, parses and validates the config file at path.
func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}
	var c Config
	if err := toml.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}
	if err := c.validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Config) validate() error {
	if c.App.Device.Protocol != ProtocolElgato && c.App.Device.Protocol != ProtocolDIY {
		return fmt.Errorf("config: device.protocol must be %q or %q, got %q",
			ProtocolElgato, ProtocolDIY, c.App.Device.Protocol)
	}
	if c.App.Language == "" {
		c.App.Language = "en"
	}
	// Clamp layout to valid range.
	if c.App.Layout.Columns <= 0 || c.App.Layout.Columns > 6 {
		c.App.Layout.Columns = 4
	}
	if c.App.Layout.Rows <= 0 || c.App.Layout.Rows > 6 {
		c.App.Layout.Rows = 5
	}
	if len(c.Layers) == 0 {
		return fmt.Errorf("config: at least one layer is required")
	}

	rows := c.App.Layout.Rows
	cols := c.App.Layout.Columns

	for i, l := range c.Layers {
		if l.Name == "" {
			return fmt.Errorf("config: layers[%d].name is empty", i)
		}
		for j, b := range l.Buttons {
			if b.Row < 0 || b.Row >= rows {
				return fmt.Errorf("config: layer %q buttons[%d] row %d out of range [0,%d)",
					l.Name, j, b.Row, rows)
			}
			if b.Col < 0 || b.Col >= cols {
				return fmt.Errorf("config: layer %q buttons[%d] col %d out of range [0,%d)",
					l.Name, j, b.Col, cols)
			}
			if !ValidActions[b.Action] {
				return fmt.Errorf("config: layer %q buttons[%d] action %q invalid",
					l.Name, j, b.Action)
			}
			if b.Action == ActionText && b.Content == "" {
				return fmt.Errorf("config: layer %q buttons[%d] text requires content",
					l.Name, j)
			}
			if b.Action != ActionText && b.Content != "" {
				return fmt.Errorf("config: layer %q buttons[%d] action %q must not have content",
					l.Name, j, b.Action)
			}
		}
	}
	return nil
}

// ButtonAt returns the button at (row, col) for a given layer, or (Button{}, false).
func (l Layer) ButtonAt(row, col int) (Button, bool) {
	for _, b := range l.Buttons {
		if b.Row == row && b.Col == col {
			return b, true
		}
	}
	return Button{}, false
}
