// Package theme defines 13 preset color themes for the RadKeys UI.
// Display names are translated via i18n keys (theme.<id>).
package theme

// Preset is a named set of UI colors (hex strings without alpha).
type Preset struct {
	ID         string // machine-readable key (i18n: theme.<id>)
	Name       string // fallback display name (English)
	Background string
	Button     string
	Fixed      string
}

// Presets is the list of all selectable themes (13). Index 0 is the default.
var Presets = []Preset{
	{"system", "System default", "", "", ""},
	{"dracula", "Dracula", "#282a36", "#44475a", "#6272a4"},
	{"solarized_dark", "Solarized Dark", "#002b36", "#073642", "#586e75"},
	{"monokai", "Monokai", "#272822", "#3e3d32", "#49483e"},
	{"gruvbox_dark", "Gruvbox Dark", "#282828", "#3c3836", "#504945"},
	{"nord", "Nord", "#2e3440", "#3b4252", "#434c5e"},
	{"one_dark", "One Dark", "#282c34", "#353b41", "#4b5263"},
	{"tokyo_night", "Tokyo Night", "#1a1b26", "#24283b", "#414868"},
	{"catppuccin_mocha", "Catppuccin Mocha", "#1e1e2e", "#313244", "#45475a"},
	{"solarized_light", "Solarized Light", "#fdf6e3", "#eee8d5", "#93a1a1"},
	{"gruvbox_light", "Gruvbox Light", "#fbf1c7", "#ebdbb2", "#d5c4a1"},
	{"light_gray", "Light Gray", "#e0e0e0", "#c0c0c0", "#a0a0a0"},
	{"dark_gray", "Dark Gray", "#202020", "#303030", "#404040"},
}

// FindPreset returns the Preset with the given ID, or Presets[0] if not found.
// Also accepts legacy display names for backward compatibility.
func FindPreset(id string) (Preset, bool) {
	for _, p := range Presets {
		if p.ID == id || p.Name == id {
			return p, true
		}
	}
	return Presets[0], false
}
