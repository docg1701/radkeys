// Package deck holds the runtime navigation state machine that maps a physical
// button press to an effect (navigate, text, copy, level_up, go_home).
package deck

import "github.com/docg1701/radkeys/internal/config"

// EffectType is the kind of action produced by a button press.
type EffectType int

const (
	EffectNone     EffectType = iota // nothing to do
	EffectCopy                       // copy the preview to the clipboard
	EffectNavigate                   // the active screen changed
	EffectPreview                    // a text was loaded into the preview
)

// Effect is the result of pressing a button, to be applied by the UI layer.
type Effect struct {
	Type EffectType
	Text string
}

// Deck navigates the config screens.
type Deck struct {
	cfg     *config.Config
	current string   // active screen id
	preview string   // current text loaded for copy
	stack   []string // parent screen ids, for level_up
}

// New creates a Deck starting at the first screen.
func New(cfg *config.Config) *Deck {
	return &Deck{cfg: cfg, current: cfg.Screens[0].ID}
}

// CurrentScreen returns the active screen.
func (d *Deck) CurrentScreen() config.Screen {
	if s, ok := d.cfg.ScreenByID(d.current); ok {
		return s
	}
	return d.cfg.Screens[0]
}

// Preview returns the text currently loaded for copying.
func (d *Deck) Preview() string { return d.preview }

// Press interprets a physical button press at index and mutates state.
func (d *Deck) Press(index int) Effect {
	f := d.cfg.App.FixedButtons
	switch index {
	case f.Copy:
		return Effect{Type: EffectCopy, Text: d.preview}
	case f.LevelUp:
		d.levelUp()
		return Effect{Type: EffectNavigate}
	case f.GoHome:
		d.current = d.cfg.Screens[0].ID
		d.stack = d.stack[:0]
		return Effect{Type: EffectNavigate}
	}

	b, ok := d.buttonByIndex(index)
	if !ok {
		return Effect{Type: EffectNone}
	}
	switch b.Action {
	case config.ActionNavigate:
		d.stack = append(d.stack, d.current)
		d.current = b.Target
		return Effect{Type: EffectNavigate}
	case config.ActionText:
		d.preview = b.Content
		return Effect{Type: EffectPreview, Text: b.Content}
	}
	return Effect{Type: EffectNone}
}

func (d *Deck) buttonByIndex(index int) (config.Button, bool) {
	for _, b := range d.CurrentScreen().Buttons {
		if b.Index == index {
			return b, true
		}
	}
	return config.Button{}, false
}

func (d *Deck) levelUp() {
	if len(d.stack) == 0 {
		d.current = d.cfg.Screens[0].ID
		return
	}
	last := d.stack[len(d.stack)-1]
	d.stack = d.stack[:len(d.stack)-1]
	d.current = last
}
