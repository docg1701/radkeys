// Package ui renders RadKeys: preview on top half, virtual keypad on bottom.
// Two tabs: "Atalhos" (preview + keypad) and "Ajustes" (settings).
package ui

import (
	"fmt"
	"image/color"
	"os"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/BurntSushi/toml"

	"github.com/docg1701/radkeys/internal/assets"
	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/deck"
	"github.com/docg1701/radkeys/internal/hid"
	"github.com/docg1701/radkeys/internal/i18n"
	themes "github.com/docg1701/radkeys/internal/theme"
)

func Run(cfg *config.Config, configPath string, reader hid.Reader) error {
	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())
	iconRes := fyne.NewStaticResource("icon.png", assets.IconPNG)
	a.SetIcon(iconRes)

	i18n.SetLanguage(cfg.App.Language)
	if cfg.App.ConfigPath == "" {
		cfg.App.ConfigPath = configPath
	}

	title := fmt.Sprintf("RadKeys — %s", cfg.App.Radiologist)
	w := a.NewWindow(title)
	w.Resize(fyne.NewSize(1280, 800))
	w.SetIcon(iconRes)

	cols := cfg.App.Layout.Columns
	if cols <= 0 {
		cols = 4
	}
	rows := cfg.App.Layout.Rows
	if rows <= 0 {
		rows = 5
	}

	thm := resolveTheme(cfg)

	u := &appUI{
		cfg:        cfg,
		configPath: configPath,
		deck:       deck.New(cfg),
		reader:     reader,
		a:          a,
		win:        w,
		titleBase:  "RadKeys",
		cols:       cols,
		rows:       rows,
		thm:        thm,
		preview:    widget.NewLabel(i18n.T("preview.placeholder")),
	}
	u.preview.Wrapping = fyne.TextWrapWord
	u.preview.TextStyle = fyne.TextStyle{Monospace: true}

	u.keypad = container.NewGridWithColumns(cols)

	previewArea := u.previewBox()
	keypadArea := container.NewPadded(u.keypad)
	split := container.NewVSplit(previewArea, keypadArea)
	// No SetOffset — usuário redimensiona livremente.

	tabs := container.NewAppTabs(
		container.NewTabItem(i18n.T("tab.shortcuts"), split),
		container.NewTabItem(i18n.T("tab.settings"), u.buildSettings()),
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
	a          fyne.App
	win        fyne.Window
	titleBase  string
	preview    *widget.Label
	cols       int
	rows       int
	thm        themeColors
	keypad     *fyne.Container
}

type themeColors struct {
	bg     color.NRGBA
	button color.NRGBA
	fixed  color.NRGBA
}

func resolveTheme(cfg *config.Config) themeColors {
	bg := cfg.App.Theme.Background
	btn := cfg.App.Theme.Button
	fix := cfg.App.Theme.Fixed
	if cfg.App.Theme.Preset != "" {
		if p, ok := themes.FindPreset(cfg.App.Theme.Preset); ok && p.Name != "Custom" {
			bg, btn, fix = p.Background, p.Button, p.Fixed
		}
	}
	return themeColors{
		bg:     parseHex(bg, 0x1a, 0x1a, 0x1a),
		button: parseHex(btn, 0x2a, 0x2a, 0x2a),
		fixed:  parseHex(fix, 0x3a, 0x3a, 0x3a),
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
		u.a.Clipboard().SetContent(eff.Text)
	case deck.EffectNavigate:
		u.renderScreen()
	case deck.EffectPreview:
		u.preview.SetText(eff.Text)
	}
}

func (u *appUI) renderScreen() {
	s := u.deck.CurrentScreen()
	f := u.cfg.App.FixedButtons
	fixed := []config.Button{
		{Index: f.Copy, Label: i18n.T("button.copy")},
		{Index: f.LevelUp, Label: i18n.T("button.back")},
		{Index: f.GoHome, Label: i18n.T("button.home")},
	}
	all := append(append([]config.Button{}, fixed...), s.Buttons...)

	totalSlots := u.cols * u.rows
	u.keypad.Objects = u.keypad.Objects[:0]
	for i := 0; i < totalSlots; i++ {
		if i < len(all) {
			b := all[i]
			btn := widget.NewButton(b.Label, func() { u.press(b.Index) })
			u.keypad.Objects = append(u.keypad.Objects, btn)
		} else {
			rect := canvas.NewRectangle(u.thm.button)
			u.keypad.Objects = append(u.keypad.Objects, rect)
		}
	}
	u.keypad.Refresh()
}

func (u *appUI) previewBox() fyne.CanvasObject {
	bg := canvas.NewRectangle(u.thm.bg)
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

func (u *appUI) buildSettings() fyne.CanvasObject {
	cfg := u.cfg

	radEnt := widget.NewEntry()
	radEnt.SetText(cfg.App.Radiologist)

	langSel := widget.NewSelect(i18n.Supported, nil)
	langSel.SetSelected(cfg.App.Language)

	themeSel := widget.NewSelect(themes.PresetNames(), nil)
	themeSel.SetSelected(cfg.App.Theme.Preset)

	colsEnt := widget.NewEntry()
	colsEnt.SetText(strconv.Itoa(cfg.App.Layout.Columns))

	rowsEnt := widget.NewEntry()
	rowsEnt.SetText(strconv.Itoa(cfg.App.Layout.Rows))

	configLbl := widget.NewLabel(cfg.App.ConfigPath)
	configLbl.Wrapping = fyne.TextTruncate
	chooseBtn := widget.NewButton("Procurar...", func() {
		dialog.NewFileOpen(func(rc fyne.URIReadCloser, err error) {
			if err != nil || rc == nil {
				return
			}
			u.configPath = rc.URI().Path()
			configLbl.SetText(u.configPath)
		}, u.win).Show()
	})

	vidEnt := widget.NewEntry()
	vidEnt.SetText(fmt.Sprintf("0x%04x", cfg.App.Device.VendorID))
	pidEnt := widget.NewEntry()
	pidEnt.SetText(fmt.Sprintf("0x%04x", cfg.App.Device.ProductID))
	protoSel := widget.NewSelect([]string{config.ProtocolElgato, config.ProtocolDIY}, nil)
	protoSel.SetSelected(cfg.App.Device.Protocol)

	save := func() {
		cfg.App.Radiologist = radEnt.Text
		cfg.App.Language = langSel.Selected
		cfg.App.Theme.Preset = themeSel.Selected
		if v, err := strconv.Atoi(colsEnt.Text); err == nil {
			cfg.App.Layout.Columns = v
		}
		if v, err := strconv.Atoi(rowsEnt.Text); err == nil {
			cfg.App.Layout.Rows = v
		}
		if v, err := strconv.ParseUint(strings.TrimPrefix(vidEnt.Text, "0x"), 16, 16); err == nil {
			cfg.App.Device.VendorID = uint16(v)
		}
		if v, err := strconv.ParseUint(strings.TrimPrefix(pidEnt.Text, "0x"), 16, 16); err == nil {
			cfg.App.Device.ProductID = uint16(v)
		}
		cfg.App.Device.Protocol = protoSel.Selected

		cfg.App.ConfigPath = u.configPath

		f, err := os.Create(u.configPath)
		if err != nil {
			dialog.ShowError(fmt.Errorf("salvar: %w", err), u.win)
			return
		}
		defer f.Close()
		if err := toml.NewEncoder(f).Encode(cfg); err != nil {
			dialog.ShowError(fmt.Errorf("TOML: %w", err), u.win)
			return
		}

		i18n.SetLanguage(cfg.App.Language)
		u.win.SetTitle(fmt.Sprintf("%s — %s", u.titleBase, cfg.App.Radiologist))
		u.thm = resolveTheme(cfg)

		// Reconstruir tabs com novo idioma.
		tabs := u.win.Content().(*container.AppTabs)
		tabs.Items[0].Text = i18n.T("tab.shortcuts")
		tabs.Items[1].Text = i18n.T("tab.settings")
		tabs.Items[1].Content = u.buildSettings()

		if cfg.App.Layout.Columns != u.cols || cfg.App.Layout.Rows != u.rows {
			u.cols = cfg.App.Layout.Columns
			u.rows = cfg.App.Layout.Rows
			u.keypad = container.NewGridWithColumns(u.cols)
			previewArea := u.previewBox()
			keypadArea := container.NewPadded(u.keypad)
			split := container.NewVSplit(previewArea, keypadArea)
			useTab := container.NewBorder(nil, nil, nil, nil, split)
			tabs.Items[0] = container.NewTabItem(i18n.T("tab.shortcuts"), useTab)
		}
		tabs.Refresh()

		u.renderScreen()
		dialog.ShowInformation(i18n.T("settings.saved_title"), i18n.T("settings.saved_msg"), u.win)
	}

	form := widget.NewForm(
		widget.NewFormItem("Radiologista", radEnt),
		widget.NewFormItem("Idioma", langSel),
		widget.NewFormItem("Tema", themeSel),
		widget.NewFormItem("Colunas", colsEnt),
		widget.NewFormItem("Linhas", rowsEnt),
		widget.NewFormItem("Arquivo de config", container.NewHBox(configLbl, chooseBtn)),
		widget.NewFormItem("Dispositivo USB", container.NewHBox(widget.NewLabel("VID"), vidEnt, widget.NewLabel("PID"), pidEnt, protoSel)),
	)
	form.SubmitText = i18n.T("settings.save")
	form.OnSubmit = save

	return container.NewVScroll(form)
}
