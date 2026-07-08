// Package ui renders RadKeys: preview on top half, virtual keypad on bottom
// half. Layout (columns/rows) and colors are fully configurable via TOML.
package ui

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

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

func Run(cfg *config.Config, configPath string, reader hid.Reader) error {
	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())
	w := a.NewWindow("RadKeys")
	w.Resize(fyne.NewSize(1280, 800))

	cols := cfg.App.Layout.Columns
	if cols <= 0 {
		cols = 4
	}
	rows := cfg.App.Layout.Rows
	if rows <= 0 {
		rows = 5
	}

	u := &appUI{
		cfg:        cfg,
		configPath: configPath,
		deck:       deck.New(cfg),
		reader:     reader,
		fapp:       a,
		win:        w,
		cols:       cols,
		rows:       rows,
		thm:        parseTheme(cfg),
		preview:    widget.NewLabel("Selecione uma frase."),
		title:      widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	}
	u.preview.Wrapping = fyne.TextWrapWord
	u.preview.TextStyle = fyne.TextStyle{Monospace: true}

	// Bottom half: virtual keypad (grid of buttons).
	u.keypad = container.NewGridWithColumns(cols)

	previewArea := u.previewBox()
	keypadArea := container.NewPadded(u.keypad)

	// Split: top = preview (50%), bottom = keypad (50%).
	split := container.NewVSplit(previewArea, keypadArea)
	split.SetOffset(0.5)

	useTab := container.NewBorder(u.title, nil, nil, nil, split)

	tabs := container.NewAppTabs(
		container.NewTabItem("Atalhos", useTab),
		container.NewTabItem("Editar", u.buildEditor()),
	)
	w.SetContent(tabs)
	u.renderScreen()

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
	preview    *widget.Label
	title      *widget.Label
	cols       int
	rows       int
	thm        themeColors
	keypad     *fyne.Container
}

type themeColors struct {
	bg      color.NRGBA
	preview color.NRGBA
	button  color.NRGBA
	fixed   color.NRGBA
}

func parseTheme(cfg *config.Config) themeColors {
	t := cfg.App.Theme
	return themeColors{
		bg:      parseHex(t.Background, 0x1a, 0x1a, 0x1a),
		preview: parseHex(t.Background, 0x1a, 0x1a, 0x1a),
		button:  parseHex(t.Button, 0x2a, 0x2a, 0x2a),
		fixed:   parseHex(t.Fixed, 0x3a, 0x3a, 0x3a),
	}
}

func parseHex(s string, dr, dg, db uint8) color.NRGBA {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return color.NRGBA{R: dr, G: dg, B: db, A: 0xFF}
	}
	r, _ := strconv.ParseUint(s[0:2], 16, 8)
	g, _ := strconv.ParseUint(s[2:4], 16, 8)
	b, _ := strconv.ParseUint(s[4:6], 16, 8)
	return color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 0xFF}
}

func (u *appUI) press(index int) {
	eff := u.deck.Press(index)
	switch eff.Type {
	case deck.EffectCopy:
		u.fapp.Clipboard().SetContent(eff.Text)
	case deck.EffectNavigate:
		u.renderScreen()
	case deck.EffectPreview:
		u.preview.SetText(eff.Text)
	}
}

func (u *appUI) renderScreen() {
	s := u.deck.CurrentScreen()
	u.title.SetText(fmt.Sprintf("%s — %s", u.cfg.App.Name, s.Title))

	f := u.cfg.App.FixedButtons
	fixed := []config.Button{
		{Index: f.Copy, Label: "Copiar"},
		{Index: f.LevelUp, Label: "Voltar"},
		{Index: f.GoHome, Label: "Início"},
	}

	all := append(append([]config.Button{}, fixed...), s.Buttons...)
	// Fill the grid with available buttons; empty slots get a placeholder.
	totalSlots := u.cols * u.rows
	u.keypad.Objects = u.keypad.Objects[:0]
	for i := 0; i < totalSlots; i++ {
		if i < len(all) {
			u.keypad.Objects = append(u.keypad.Objects, u.makeBtn(all[i]))
		} else {
			u.keypad.Objects = append(u.keypad.Objects, u.emptySlot())
		}
	}
	u.keypad.Refresh()
}

func (u *appUI) makeBtn(b config.Button) fyne.CanvasObject {
	btn := widget.NewButton(b.Label, func() { u.press(b.Index) })
	return container.NewGridWrap(fyne.NewSize(120, 80), btn)
}

func (u *appUI) emptySlot() fyne.CanvasObject {
	rect := canvas.NewRectangle(u.thm.button)
	return container.NewGridWrap(fyne.NewSize(120, 80), rect)
}

func (u *appUI) previewBox() fyne.CanvasObject {
	bg := canvas.NewRectangle(u.thm.preview)
	scroll := container.NewVScroll(u.preview)
	return container.NewStack(bg, container.NewPadded(scroll))
}

func (u *appUI) pollHID() {
	for ev := range u.reader.Events() {
		if !ev.Pressed {
			continue
		}
		idx := ev.Index
		fyne.Do(func() { u.press(idx) })
	}
}
