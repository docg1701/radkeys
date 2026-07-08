// Package ui renders the RadKeys use screen: title, preview, and a grid of
// buttons (3 fixed + configurable). HID events and mouse clicks both drive
// the deck; the window is always-on-top and never grabs focus on HID input.
package ui

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/deck"
	"github.com/docg1701/radkeys/internal/hid"
)

// Run builds the window, wires the HID reader to the deck, and blocks until
// the window closes. configPath is the path to radkeys.config.toml (for the
// editor to save back to).
func Run(cfg *config.Config, configPath string, reader hid.Reader) error {
	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())
	w := a.NewWindow("RadKeys")
	w.SetFixedSize(true)
	w.Resize(fyne.NewSize(960, 640))

	preview := widget.NewRichTextFromMarkdown("*Selecione uma frase para pré‑visualizar.*")
	preview.Wrapping = fyne.TextWrapWord

	title := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	grid := container.NewGridWithColumns(6)

	u := &appUI{cfg: cfg, configPath: configPath, deck: deck.New(cfg), reader: reader, fapp: a, win: w, preview: preview, title: title, grid: grid}
	u.renderScreen()

	editBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), u.openEditor)
	editBtn.Importance = widget.LowImportance

	top := container.NewBorder(nil, nil, title, editBtn)
	content := container.NewBorder(top, u.previewBox(), nil, nil, grid)
	w.SetContent(content)

	// Always-on-top: NOT available in Fyne v2.7.4 (PR #6184 is on develop / v2.8.0,
	// still rc1 as of 2026-07-07). Decision: MVP stays on v2.7.4 stable without
	// always-on-top; re-add below once Fyne v2.8.0 is stable. See
	// research/fyne-always-on-top.md.
	//
	//   if dw, ok := w.(desktop.Window); ok { dw.RequestAlwaysOnTop() } // before Show

	if err := reader.Open(); err != nil {
		return fmt.Errorf("hid: open: %w", err)
	}
	go u.pollHID()
	w.SetOnClosed(func() { _ = reader.Close() })
	w.ShowAndRun()
	return nil
}

type appUI struct {
	cfg        *config.Config
	configPath string
	deck       *deck.Deck
	reader     hid.Reader
	fapp       fyne.App
	win        fyne.Window
	preview    *widget.RichText
	title      *widget.Label
	grid       *fyne.Container
}

func (u *appUI) press(index int) {
	eff := u.deck.Press(index)
	switch eff.Type {
	case deck.EffectCopy:
		u.fapp.Clipboard().SetContent(eff.Text)
	case deck.EffectNavigate:
		u.renderScreen()
	case deck.EffectPreview:
		u.preview.ParseMarkdown(eff.Text)
	}
}

func (u *appUI) renderScreen() {
	s := u.deck.CurrentScreen()
	u.title.SetText(fmt.Sprintf("%s — %s", u.cfg.App.Name, s.Title))

	f := u.cfg.App.FixedButtons
	objs := []fyne.CanvasObject{
		u.fixedBtn("📋 Copiar", f.Copy),
		u.fixedBtn("⬆ Voltar", f.LevelUp),
		u.fixedBtn("⌂ Início", f.GoHome),
	}
	for _, b := range s.Buttons {
		b := b
		btn := widget.NewButtonWithIcon(fmt.Sprintf("%d  %s", b.Index, b.Label), nil, func() { u.press(b.Index) })
		btn.Importance = widget.MediumImportance
		objs = append(objs, btn)
	}
	u.grid.Objects = objs
	u.grid.Refresh()
}

func (u *appUI) fixedBtn(label string, index int) *widget.Button {
	btn := widget.NewButtonWithIcon(label, nil, func() { u.press(index) })
	btn.Importance = widget.HighImportance
	return btn
}

func (u *appUI) previewBox() fyne.CanvasObject {
	sep := canvas.NewRectangle(color.NRGBA{R: 0x40, G: 0x40, B: 0x40, A: 0xFF})
	sep.SetMinSize(fyne.NewSize(1, 2))
	return container.NewBorder(sep, nil, nil, nil, container.NewPadded(u.preview))
}

// pollHID forwards physical button presses to the UI thread via fyne.Do.
func (u *appUI) pollHID() {
	for ev := range u.reader.Events() {
		if !ev.Pressed {
			continue
		}
		idx := ev.Index
		fyne.Do(func() { u.press(idx) })
	}
}
