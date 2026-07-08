// Package theme defines 12 preset color themes for the RadKeys UI.
// 10 are inspired by popular terminal themes; 2 are gray-only (light/dark).
package theme

// Preset is a named set of UI colors (hex strings without alpha).
type Preset struct {
	Name       string
	Background string
	Button     string
	Fixed      string
}

// Presets is the list of all selectable themes (13). Index 0 is the default.
var Presets = []Preset{
	{"Padrão do sistema", "", "", ""},
	{"Dracula", "#282a36", "#44475a", "#6272a4"},
	{"Solarized Dark", "#002b36", "#073642", "#586e75"},
	{"Monokai", "#272822", "#3e3d32", "#49483e"},
	{"Gruvbox Dark", "#282828", "#3c3836", "#504945"},
	{"Nord", "#2e3440", "#3b4252", "#434c5e"},
	{"One Dark", "#282c34", "#353b41", "#4b5263"},
	{"Tokyo Night", "#1a1b26", "#24283b", "#414868"},
	{"Catppuccin Mocha", "#1e1e2e", "#313244", "#45475a"},
	{"Solarized Light", "#fdf6e3", "#eee8d5", "#93a1a1"},
	{"Gruvbox Light", "#fbf1c7", "#ebdbb2", "#d5c4a1"},
	{"Light Gray", "#e0e0e0", "#c0c0c0", "#a0a0a0"},
	{"Dark Gray", "#202020", "#303030", "#404040"},
}

// PresetNames returns just the names for a dropdown.
func PresetNames() []string {
	out := make([]string, len(Presets))
	for i, p := range Presets {
		out[i] = p.Name
	}
	return out
}

// FindPreset returns the Preset with the given name, or Presets[0] if not found.
func FindPreset(name string) (Preset, bool) {
	for _, p := range Presets {
		if p.Name == name {
			return p, true
		}
	}
	return Presets[0], false
}
