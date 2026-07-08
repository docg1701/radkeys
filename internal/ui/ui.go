// Package ui renders the RadKeys use screen: square buttons around a central
// preview area. Layout is driven by [app.layout] in the config so it adapts
// to any DIY device. Two tabs: "Uso" (buttons + preview) and "Config" (editor).
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

func Run(cfg *config.Config, configPath string, reader hid.Reader) error {
	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())
	w := a.NewWindow("RadKeys")
	w.SetFixedSize(true)
	w.Resize(fyne.NewSize(960, 640))

	u := &appUI{
		cfg:        cfg,
		configPath: configPath,
		deck:       deck.New(cfg),
		reader:     reader,
		fapp:       a,
		win:        w,
		preview:    widget.NewRichTextFromMarkdown("*Selecione uma frase.*"),
		title:      widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
	}
	u.preview.Wrapping = fyne.TextWrapWord

	cols := cfg.App.Layout.Columns
	if cols <= 0 {
		cols = 4
	}
	u.topRowC = container.NewGridWithColumns(cols)
	u.bottomRowC = container.NewGridWithColumns(cols)
	u.leftColC = container.NewVBox()
	u.rightColC = container.NewVBox()

	previewCenter := u.previewBox()
	ring := container.NewBorder(u.topRowC, u.bottomRowC, u.leftColC, u.rightColC, previewCenter)
	useTab := container.NewBorder(u.title, nil, nil, nil, ring)

	tabs := container.NewAppTabs(
		container.NewTabItem("Uso", useTab),
		container.NewTabItem("Config", u.buildEditor()),
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
	preview    *widget.RichText
	title      *widget.Label
	topRowC    *fyne.Container
	bottomRowC *fyne.Container
	leftColC   *fyne.Container
	rightColC  *fyne.Container
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
	all := append([]config.Button{
		{Index: f.Copy, Label: "Copiar"},
		{Index: f.LevelUp, Label: "Voltar"},
		{Index: f.GoHome, Label: "Início"},
	}, s.Buttons...)

	cols := u.cfg.App.Layout.Columns
	if cols <= 0 {
		cols = 4
	}

	topN := cols
	if len(all) <= cols {
		topN = len(all)
	}
	bottomN := 0
	if len(all) > cols {
		bottomN = cols
		if bottomN > len(all)-topN {
			bottomN = len(all) - topN
		}
	}
	rest := len(all) - topN - bottomN
	leftN := rest / 2
	rightN := rest - leftN

	u.fillRow(u.topRowC, all[:topN])
	if bottomN > 0 {
		u.fillRow(u.bottomRowC, all[topN:topN+bottomN])
	} else {
		u.bottomRowC.Objects = u.bottomRowC.Objects[:0]
		u.bottomRowC.Refresh()
	}
	rs := topN + bottomN
	u.fillCol(u.leftColC, all[rs:rs+leftN])
	u.fillCol(u.rightColC, all[rs+leftN:rs+leftN+rightN])
}

func (u *appUI) fillRow(c *fyne.Container, btns []config.Button) {
	c.Objects = c.Objects[:0]
	for _, b := range btns {
		c.Objects = append(c.Objects, u.makeBtn(b))
	}
	c.Refresh()
}

func (u *appUI) fillCol(c *fyne.Container, btns []config.Button) {
	c.Objects = c.Objects[:0]
	for _, b := range btns {
		c.Objects = append(c.Objects, u.makeBtn(b))
	}
	c.Refresh()
}

func (u *appUI) makeBtn(b config.Button) *widget.Button {
	return widget.NewButton(b.Label, func() { u.press(b.Index) })
}

func (u *appUI) previewBox() fyne.CanvasObject {
	bg := canvas.NewRectangle(color.NRGBA{R: 0x1a, G: 0x1a, B: 0x1a, A: 0xFF})
	return container.NewStack(bg, container.NewPadded(u.preview))
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
