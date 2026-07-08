// Package config loads and validates radkeys.config.toml.
//
// Exemplo:
//
//	cfg, err := config.Load("radkeys.config.toml")
//	if err != nil { log.Fatal(err) }
//	root, ok := cfg.ScreenByID(cfg.App.Device.VendorID)
package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Device protocol selectors.
const (
	ProtocolElgato = "elgato"
	ProtocolDIY    = "radkeys-diy"
)

// Button actions.
const (
	ActionNavigate = "navigate"
	ActionText     = "text"
)

// Config is the root of radkeys.config.toml.
type Config struct {
	App     App      `toml:"app"`
	Screens []Screen `toml:"screens"`
}

// App holds app-wide settings: device connection and the 3 fixed buttons.
type App struct {
	Name         string       `toml:"name"`
	Version      string       `toml:"version"`
	Device       Device       `toml:"device"`
	FixedButtons FixedButtons `toml:"fixed_buttons"`
}

// Device identifies the USB HID custom device to open.
type Device struct {
	VendorID  uint16 `toml:"vendor_id"`
	ProductID uint16 `toml:"product_id"`
	Protocol  string `toml:"protocol"`
}

// FixedButtons are the indices (0-based) of the 3 global control buttons.
type FixedButtons struct {
	Copy    int `toml:"copy"`
	LevelUp int `toml:"level_up"`
	GoHome  int `toml:"go_home"`
}

// Screen is one page of the shortcut deck.
type Screen struct {
	ID      string   `toml:"id"`
	Title   string   `toml:"title"`
	Buttons []Button `toml:"buttons"`
}

// Button maps a physical button index to an action.
type Button struct {
	Index   int    `toml:"index"`
	Label   string `toml:"label"`
	Action  string `toml:"action"`
	Target  string `toml:"target,omitempty"`
	Content string `toml:"content,omitempty"`
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
	if len(c.Screens) == 0 {
		return fmt.Errorf("config: at least one screen is required")
	}
	ids := map[string]struct{}{}
	for i, s := range c.Screens {
		if s.ID == "" {
			return fmt.Errorf("config: screens[%d].id is empty", i)
		}
		if _, dup := ids[s.ID]; dup {
			return fmt.Errorf("config: duplicate screen id %q", s.ID)
		}
		ids[s.ID] = struct{}{}
		for j, b := range s.Buttons {
			if b.Action != ActionNavigate && b.Action != ActionText {
				return fmt.Errorf("config: screen %q buttons[%d].action %q invalid (want %q or %q)",
					s.ID, j, b.Action, ActionNavigate, ActionText)
			}
			if b.Action == ActionNavigate && b.Target == "" {
				return fmt.Errorf("config: screen %q buttons[%d] navigate requires target", s.ID, j)
			}
			if b.Action == ActionText && b.Content == "" {
				return fmt.Errorf("config: screen %q buttons[%d] text requires content", s.ID, j)
			}
		}
	}
	for _, s := range c.Screens {
		for _, b := range s.Buttons {
			if b.Action == ActionNavigate {
				if _, ok := ids[b.Target]; !ok {
					return fmt.Errorf("config: screen %q button %q navigates to unknown screen %q",
						s.ID, b.Label, b.Target)
				}
			}
		}
	}
	return nil
}

// ScreenByID returns a screen by id.
func (c *Config) ScreenByID(id string) (Screen, bool) {
	for _, s := range c.Screens {
		if s.ID == id {
			return s, true
		}
	}
	return Screen{}, false
}

// IsFixed reports whether index is one of the 3 global fixed buttons.
func (c *Config) IsFixed(index int) bool {
	f := c.App.FixedButtons
	return index == f.Copy || index == f.LevelUp || index == f.GoHome
}
