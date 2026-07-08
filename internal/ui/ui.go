// Package ui renders RadKeys: preview on top half, virtual keypad on bottom.
// Three tabs: "Atalhos" (preview + keypad), "Ajustes" (settings), "Sobre" (about).
package ui

import (
	"fmt"
	"image/color"
	"net/url"
	"os"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
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

	// Apply the preset theme to the entire UI.
	customTheme := resolveFullTheme(cfg)
	a.Settings().SetTheme(customTheme)

	iconRes := fyne.NewStaticResource("icon.png", appIconData(cfg))
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

	thm := resolveKeypadColors(cfg)

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

	tabs := container.NewAppTabs(
		container.NewTabItem(i18n.T("tab.shortcuts"), split),
		container.NewTabItem(i18n.T("tab.settings"), u.buildSettings()),
		container.NewTabItem(i18n.T("tab.about"), u.buildAbout()),
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
	thm        keypadColors
	keypad     *fyne.Container
}

type keypadColors struct {
	bg     color.NRGBA
	button color.NRGBA
	fixed  color.NRGBA
}

func resolveFullTheme(cfg *config.Config) fyne.Theme {
	if cfg.App.Theme.Preset != "" {
		if p, ok := themes.FindPreset(cfg.App.Theme.Preset); ok {
			return themes.NewCustomTheme(p)
		}
	}
	return themes.NewCustomTheme(themes.Presets[0])
}

func resolveKeypadColors(cfg *config.Config) keypadColors {
	bg := cfg.App.Theme.Background
	btn := cfg.App.Theme.Button
	fix := cfg.App.Theme.Fixed
	if cfg.App.Theme.Preset != "" {
		if p, ok := themes.FindPreset(cfg.App.Theme.Preset); ok {
			bg, btn, fix = p.Background, p.Button, p.Fixed
		}
	}
	return keypadColors{
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

func appIconData(cfg *config.Config) []byte {
	if cfg.App.Theme.Icon != "" {
		if data := assets.IconData(cfg.App.Theme.Icon); data != nil {
			return data
		}
	}
	return assets.IconPNG
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

// ---------------------------------------------------------------------------
// Settings tab — modern card-based layout with visual groups
// ---------------------------------------------------------------------------

func (u *appUI) buildSettings() fyne.CanvasObject {
	cfg := u.cfg

	// --- Widgets ---

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
	chooseBtn := widget.NewButton(i18n.T("settings.browse"), func() {
		dialog.NewFileOpen(func(rc fyne.URIReadCloser, err error) {
			if err != nil || rc == nil {
				return
			}
			u.configPath = rc.URI().Path()
			configLbl.SetText(u.configPath)
		}, u.win).Show()
	})
	chooseBtn.Importance = widget.MediumImportance
	configRow := container.NewBorder(nil, nil, nil, chooseBtn, configLbl)

	vidEnt := widget.NewEntry()
	vidEnt.SetText(fmt.Sprintf("0x%04x", cfg.App.Device.VendorID))
	vidEnt.SetMinRowsVisible(1)
	pidEnt := widget.NewEntry()
	pidEnt.SetText(fmt.Sprintf("0x%04x", cfg.App.Device.ProductID))
	pidEnt.SetMinRowsVisible(1)
	protoSel := widget.NewSelect([]string{config.ProtocolElgato, config.ProtocolDIY}, nil)
	protoSel.SetSelected(cfg.App.Device.Protocol)

	// --- Icon gallery ---

	selectedIcon := cfg.App.Theme.Icon
	iconGrid := buildIconGallery(&selectedIcon)

	// --- Save action ---

	save := func() {
		cfg.App.Radiologist = radEnt.Text
		cfg.App.Language = langSel.Selected
		cfg.App.Theme.Icon = selectedIcon
		cfg.App.Theme.Preset = themeSel.Selected
		if v, err := strconv.Atoi(colsEnt.Text); err == nil && v > 0 {
			cfg.App.Layout.Columns = v
		} else {
			cfg.App.Layout.Columns = 1
			colsEnt.SetText("1")
		}
		if v, err := strconv.Atoi(rowsEnt.Text); err == nil && v > 0 {
			cfg.App.Layout.Rows = v
		} else {
			cfg.App.Layout.Rows = 1
			rowsEnt.SetText("1")
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

		iconRes := fyne.NewStaticResource("icon.png", appIconData(cfg))
		u.a.SetIcon(iconRes)
		u.win.SetIcon(iconRes)

		u.a.Settings().SetTheme(resolveFullTheme(cfg))
		u.thm = resolveKeypadColors(cfg)

		tabs := u.win.Content().(*container.AppTabs)
		tabs.Items[0].Text = i18n.T("tab.shortcuts")
		tabs.Items[1].Text = i18n.T("tab.settings")
		tabs.Items[2].Text = i18n.T("tab.about")
		tabs.Items[1].Content = u.buildSettings()
		tabs.Items[2].Content = u.buildAbout()

		if cfg.App.Layout.Columns != u.cols || cfg.App.Layout.Rows != u.rows {
			u.cols = cfg.App.Layout.Columns
			u.rows = cfg.App.Layout.Rows
			u.keypad = container.NewGridWithColumns(u.cols)
			previewArea := u.previewBox()
			keypadArea := container.NewPadded(u.keypad)
			split := container.NewVSplit(previewArea, keypadArea)
			tabs.Items[0] = container.NewTabItem(i18n.T("tab.shortcuts"), split)
		}
		tabs.Refresh()

		u.renderScreen()

	}

	saveBtn := widget.NewButton(i18n.T("settings.save"), save)
	saveBtn.Importance = widget.HighImportance

	// --- Cards ---

	cards := []fyne.CanvasObject{
		makeCard(i18n.T("settings.group_general"), "",
			labeled(i18n.T("settings.radiologist"), radEnt),
		),
		makeCard(i18n.T("settings.group_locale"), "",
			container.NewGridWithColumns(2,
				labeled(i18n.T("settings.language"), langSel),
				labeled(i18n.T("settings.theme"), themeSel),
			),
		),
		makeCard(i18n.T("settings.group_icon"), "",
			iconGrid,
		),
		makeCard(i18n.T("settings.group_layout"), "",
			makeDualField(i18n.T("settings.columns"), colsEnt, i18n.T("settings.rows"), rowsEnt),
		),
		makeCard(i18n.T("settings.group_config"), "",
			configRow,
		),
		makeCard(i18n.T("settings.group_device"), "",
			makeUSBRow(i18n.T("settings.vid"), vidEnt, i18n.T("settings.pid"), pidEnt, i18n.T("settings.protocol"), protoSel),
		),
	}

	content := container.NewVBox(cards...)
	content.Add(container.NewCenter(saveBtn))

	return container.NewVScroll(container.NewPadded(content))
}

// ---------------------------------------------------------------------------
// About tab
// ---------------------------------------------------------------------------

func (u *appUI) buildAbout() fyne.CanvasObject {
	ver := u.cfg.App.Version
	if ver == "" {
		ver = "dev"
	}

	header := widget.NewLabel(fmt.Sprintf("RadKeys — %s", fmt.Sprintf(i18n.T("about.version"), ver)))
	header.TextStyle = fyne.TextStyle{Bold: true}

	desc := widget.NewLabel(i18n.T("about.description"))
	desc.Wrapping = fyne.TextWrapWord

	repoURL, _ := url.Parse("https://github.com/docg1701/radkeys")
	repo := widget.NewHyperlink("github.com/docg1701/radkeys", repoURL)
	repoLine := container.NewHBox(widget.NewLabel(i18n.T("about.repository")), repo)

	author := widget.NewLabel(i18n.T("about.author"))
	author.Wrapping = fyne.TextWrapWord
	license := widget.NewLabel(i18n.T("about.license"))
	license.Wrapping = fyne.TextWrapWord

	stack := widget.NewLabel(i18n.T("about.stack"))
	stack.Wrapping = fyne.TextWrapWord

	langs := widget.NewLabel(i18n.T("about.i18n"))
	langs.Wrapping = fyne.TextWrapWord

	return container.NewVScroll(container.NewPadded(
		container.NewVBox(header, desc, stack, langs, author, license, repoLine),
	))
}

// ---------------------------------------------------------------------------
// Layout helpers for the settings tab
// ---------------------------------------------------------------------------

// labeled returns label above input so the input gets full width.
func labeled(label string, input fyne.CanvasObject) fyne.CanvasObject {
	return container.NewVBox(widget.NewLabel(label), input)
}

func makeCard(title string, subtitle string, items ...fyne.CanvasObject) fyne.CanvasObject {
	body := container.NewVBox(items...)
	return widget.NewCard(title, subtitle, body)
}

func makeDualField(l1 string, i1 fyne.CanvasObject, l2 string, i2 fyne.CanvasObject) fyne.CanvasObject {
	return container.NewGridWithColumns(2,
		labeled(l1, i1),
		labeled(l2, i2),
	)
}

func makeUSBRow(vidLabel string, vidInput fyne.CanvasObject, pidLabel string, pidInput fyne.CanvasObject, protoLabel string, protoInput fyne.CanvasObject) fyne.CanvasObject {
	vidCell := container.NewVBox(widget.NewLabel(vidLabel), vidInput)
	pidCell := container.NewVBox(widget.NewLabel(pidLabel), pidInput)
	topRow := container.NewGridWithColumns(2, vidCell, pidCell)

	protoCell := container.NewVBox(widget.NewLabel(protoLabel), protoInput)
	return container.NewVBox(topRow, protoCell)
}

func buildIconGallery(selected *string) fyne.CanvasObject {
	const cols = 6
	cells := make([]fyne.CanvasObject, 0)

	// Default (built-in) icon.
	cells = append(cells, iconTile("", assets.IconPNG, selected))

	// Embedded Obsidian icons.
	for _, name := range assets.IconNames() {
		data := assets.IconData(name)
		if data == nil {
			continue
		}
		cells = append(cells, iconTile(name, data, selected))
	}

	return container.NewGridWithColumns(cols, cells...)
}

func iconTile(name string, data []byte, selected *string) fyne.CanvasObject {
	res := fyne.NewStaticResource(name+".png", data)
	return widget.NewButtonWithIcon("", res, func() { *selected = name })
}
