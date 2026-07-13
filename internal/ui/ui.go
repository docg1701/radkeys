// Package ui renders RadKeys: preview on top half, virtual keypad on bottom.
// Three tabs: shortcuts, settings, about.
package ui

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	fyneTheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/BurntSushi/toml"

	"github.com/docg1701/radkeys/internal/assets"
	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/hid"
	"github.com/docg1701/radkeys/internal/i18n"
	"github.com/docg1701/radkeys/internal/keystroke"
	themes "github.com/docg1701/radkeys/internal/theme"
)

func Run(cfg *config.Config, configPath string, reader hid.Reader, version string) error {
	a := app.New()

	customTheme := resolveFullTheme(cfg)
	a.Settings().SetTheme(customTheme)

	iconRes := fyne.NewStaticResource("icon.png", appIconData(cfg))
	a.SetIcon(iconRes)

	i18n.SetLanguage(cfg.App.Language)

	title := fmt.Sprintf("RadKeys — %s", cfg.App.Radiologist)
	w := a.NewWindow(title)
	w.Resize(fyne.NewSize(1280, 800))
	w.SetIcon(iconRes)

	cols := cfg.App.Layout.Columns
	rows := cfg.App.Layout.Rows

	u := &appUI{
		cfg:        cfg,
		configPath: configPath,
		reader:     reader,
		a:          a,
		win:        w,
		titleBase:  "RadKeys",
		cols:       cols,
		rows:       rows,
		version:    version,
		preview:    widget.NewLabel(i18n.T("preview.placeholder")),
		current:    cfg.Screens[0].ID,
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
	u.renderGrid()

	if err := reader.Open(); err != nil {
		return fmt.Errorf("hid: open: %w", err)
	}
	go u.pollHID()
	w.SetOnClosed(func() { _ = reader.Close() })
	w.ShowAndRun()
	return nil
}

type appUI struct {
	cfg         *config.Config
	configPath  string
	current     string   // current screen id
	stack       []string // parent screen ids for prev
	reader      hid.Reader
	a           fyne.App
	win         fyne.Window
	titleBase   string
	preview     *widget.Label
	previewText string
	version     string
	cols        int
	rows        int
	keypad      *fyne.Container
	previewBg   *canvas.Rectangle
}

func resolveFullTheme(cfg *config.Config) fyne.Theme {
	if cfg.App.Theme.Preset != "" {
		if p, ok := themes.FindPreset(cfg.App.Theme.Preset); ok {
			return themes.NewCustomTheme(p)
		}
	}
	return themes.NewCustomTheme(themes.Presets[0])
}

func (u *appUI) currentScreen() config.Screen {
	s, ok := u.cfg.ScreenByID(u.current)
	if !ok {
		return u.cfg.Screens[0]
	}
	return s
}

// press handles a button press at physical (row, col). fromUI reports whether
// the press came from an on-screen button click (which gives RadKeys focus)
// versus the physical HID keypad (which preserves the RIS focus).
func (u *appUI) press(row, col int, fromUI bool) {
	b, ok := u.currentScreen().ButtonAt(row, col)
	if !ok {
		return
	}
	switch b.Action {
	case config.ActionText:
		u.previewText = b.Content
		u.preview.SetText(b.Content)
	case config.ActionCopy:
		u.a.Clipboard().SetContent(u.previewText)
	case config.ActionPaste:
		// Paste is driven by the physical keypad so the RIS keeps focus.
		// A UI click gives RadKeys focus, so sending Ctrl+V would paste into
		// RadKeys itself — refuse and explain instead.
		if fromUI {
			dialog.ShowInformation(i18n.T("button.paste"), i18n.T("paste.via_keypad_hint"), u.win)
			return
		}
		if err := keystroke.SendCtrlV(); err != nil {
			log.Printf("radkeys: paste failed: %v", err)
		}
	case config.ActionPrev:
		if len(u.stack) > 0 {
			u.current = u.stack[len(u.stack)-1]
			u.stack = u.stack[:len(u.stack)-1]
		}
	case config.ActionHome:
		u.current = u.cfg.Screens[0].ID
		u.stack = u.stack[:0]
	case config.ActionNavigate:
		u.stack = append(u.stack, u.current)
		u.current = b.Target
	}
	u.renderGrid()
}

func (u *appUI) renderGrid() {
	s := u.currentScreen()
	totalSlots := u.cols * u.rows
	u.keypad.Objects = u.keypad.Objects[:0]
	th := u.a.Settings().Theme()
	v := variantFor(th)

	for i := 0; i < totalSlots; i++ {
		r := i / u.cols
		c := i % u.cols
		if b, ok := s.ButtonAt(r, c); ok {
			row := r
			col := c
			btn := widget.NewButton(b.Label, func() { u.press(row, col, true) })
			u.keypad.Objects = append(u.keypad.Objects, btn)
		} else {
			rect := canvas.NewRectangle(th.Color(fyneTheme.ColorNameButton, v))
			u.keypad.Objects = append(u.keypad.Objects, rect)
		}
	}
	u.keypad.Refresh()
}

func appIconData(cfg *config.Config) []byte {
	if cfg.App.Theme.Icon != "" {
		data, err := os.ReadFile(cfg.App.Theme.Icon)
		if err == nil {
			return data
		}
		log.Printf("radkeys: cannot read icon %q: %v", cfg.App.Theme.Icon, err)
	}
	return assets.IconPNG
}

func (u *appUI) previewBox() fyne.CanvasObject {
	th := u.a.Settings().Theme()
	u.previewBg = canvas.NewRectangle(th.Color(fyneTheme.ColorNameBackground, variantFor(th)))
	scroll := container.NewVScroll(u.preview)
	return container.NewStack(u.previewBg, container.NewPadded(scroll))
}

func (u *appUI) pollHID() {
	for ev := range u.reader.Events() {
		if !ev.Pressed {
			continue
		}
		row, col := ev.Row, ev.Col
		fyne.Do(func() { u.press(row, col, false) })
	}
}

// ---------------------------------------------------------------------------
// Settings tab — modern card-based layout with visual groups
// ---------------------------------------------------------------------------

func (u *appUI) buildSettings() fyne.CanvasObject {
	cfg := u.cfg

	radEnt := widget.NewEntry()
	radEnt.SetText(cfg.App.Radiologist)

	langSel := widget.NewSelect(i18n.Supported, nil)
	langSel.SetSelected(cfg.App.Language)

	themeIDs := make([]string, len(themes.Presets))
	themeNames := make([]string, len(themes.Presets))
	for i, p := range themes.Presets {
		themeIDs[i] = p.ID()
		themeNames[i] = i18n.T("theme." + p.ID())
	}
	themeSel := widget.NewSelect(themeNames, nil)
	for i, id := range themeIDs {
		if id == cfg.App.Theme.Preset {
			themeSel.SetSelectedIndex(i)
			break
		}
	}

	colsEnt := widget.NewEntry()
	colsEnt.SetText(strconv.Itoa(cfg.App.Layout.Columns))

	rowsEnt := widget.NewEntry()
	rowsEnt.SetText(strconv.Itoa(cfg.App.Layout.Rows))

	configLbl := widget.NewLabel(u.configPath)
	configLbl.Wrapping = fyne.TextWrapWord
	chooseBtn := widget.NewButton(i18n.T("settings.browse"), func() {
		showFileDialog(u.win, []string{".toml"}, func(path string) {
			u.configPath = path
			configLbl.SetText(path)
		})
	})
	chooseBtn.Importance = widget.MediumImportance

	vidEnt := widget.NewEntry()
	vidEnt.SetText(fmt.Sprintf("0x%04x", cfg.App.Device.VendorID))
	vidEnt.SetMinRowsVisible(1)
	pidEnt := widget.NewEntry()
	pidEnt.SetText(fmt.Sprintf("0x%04x", cfg.App.Device.ProductID))
	pidEnt.SetMinRowsVisible(1)
	protoSel := widget.NewSelect([]string{config.ProtocolDIY}, nil)
	protoSel.SetSelected(cfg.App.Device.Protocol)

	customIconPath := cfg.App.Theme.Icon
	iconPreview := canvas.NewImageFromResource(fyne.NewStaticResource("icon.png", appIconData(cfg)))
	iconPreview.SetMinSize(fyne.NewSize(48, 48))
	iconPreview.FillMode = canvas.ImageFillContain

	iconBrowseBtn := widget.NewButton(i18n.T("settings.browse"), func() {
		showFileDialog(u.win, []string{".png"}, func(path string) {
			customIconPath = path
			data, err := os.ReadFile(path)
			if err != nil {
				log.Printf("radkeys: cannot read icon %q: %v", path, err)
				return
			}
			iconPreview.Resource = fyne.NewStaticResource("custom.png", data)
			iconPreview.Refresh()
		})
	})
	iconBrowseBtn.Importance = widget.MediumImportance

	save := func() {
		cfg.App.Radiologist = radEnt.Text
		cfg.App.Language = langSel.Selected
		cfg.App.Theme.Icon = customIconPath
		selIdx := themeSel.SelectedIndex()
		if selIdx >= 0 && selIdx < len(themeIDs) {
			cfg.App.Theme.Preset = themeIDs[selIdx]
		}
		if v, err := strconv.Atoi(colsEnt.Text); err == nil && v > 0 && v <= 6 {
			cfg.App.Layout.Columns = v
		} else {
			cfg.App.Layout.Columns = 1
			colsEnt.SetText("1")
		}
		if v, err := strconv.Atoi(rowsEnt.Text); err == nil && v > 0 && v <= 6 {
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

		f, err := os.Create(u.configPath)
		if err != nil {
			dialog.ShowError(fmt.Errorf("save: %w", err), u.win)
			return
		}
		defer func() { _ = f.Close() }()
		if err := toml.NewEncoder(f).Encode(cfg); err != nil {
			dialog.ShowError(fmt.Errorf("TOML: %w", err), u.win)
			return
		}

		i18n.SetLanguage(cfg.App.Language)
		u.win.SetTitle(fmt.Sprintf("%s — %s", u.titleBase, cfg.App.Radiologist))

		iconRes := fyne.NewStaticResource("icon.png", appIconData(cfg))
		u.a.SetIcon(iconRes)
		u.win.SetIcon(iconRes)

		newTheme := resolveFullTheme(cfg)
		u.a.Settings().SetTheme(newTheme)
		if u.previewBg != nil {
			u.previewBg.FillColor = newTheme.Color(fyneTheme.ColorNameBackground, variantFor(newTheme))
			canvas.Refresh(u.previewBg)
		}

		// Refresh tab labels and content.
		tabs := u.win.Content().(*container.AppTabs)
		tabs.Items[0].Text = i18n.T("tab.shortcuts")
		tabs.Items[1].Text = i18n.T("tab.settings")
		tabs.Items[2].Text = i18n.T("tab.about")
		tabs.Items[1].Content = u.buildSettings()
		tabs.Items[2].Content = u.buildAbout()
		tabs.Refresh()

		if cfg.App.Layout.Columns != u.cols || cfg.App.Layout.Rows != u.rows {
			u.cols = cfg.App.Layout.Columns
			u.rows = cfg.App.Layout.Rows
			u.keypad = container.NewGridWithColumns(u.cols)
			previewArea := u.previewBox()
			keypadArea := container.NewPadded(u.keypad)
			split := container.NewVSplit(previewArea, keypadArea)
			tabs.Items[0] = container.NewTabItem(i18n.T("tab.shortcuts"), split)
			tabs.Items[0].Content = split
		}
		tabs.Refresh()

		u.current = u.cfg.Screens[0].ID
		u.stack = u.stack[:0]
		u.renderGrid()
	}

	saveBtn := widget.NewButton(i18n.T("settings.save"), save)
	saveBtn.Importance = widget.HighImportance

	sections := container.NewVBox(
		section(i18n.T("settings.group_config"),
			container.NewGridWithColumns(3,
				container.NewBorder(nil, nil, widget.NewLabel(i18n.T("settings.config_file")), nil, configLbl),
				chooseBtn,
				widget.NewLabel(""),
			),
		),
		section(i18n.T("settings.group_appearance"),
			container.NewGridWithColumns(3,
				labeled(i18n.T("settings.radiologist"), radEnt),
				labeled(i18n.T("settings.language"), langSel),
				labeled(i18n.T("settings.theme"), themeSel),
			),
			container.NewGridWithColumns(3,
				widget.NewLabel(i18n.T("settings.icon")),
				iconPreview,
				iconBrowseBtn,
			),
		),
		section(i18n.T("settings.group_device"),
			container.NewGridWithColumns(3,
				labeled(i18n.T("settings.columns"), colsEnt),
				labeled(i18n.T("settings.rows"), rowsEnt),
				widget.NewLabel(""),
			),
			container.NewGridWithColumns(3,
				labeled(i18n.T("settings.vid"), vidEnt),
				labeled(i18n.T("settings.pid"), pidEnt),
				labeled(i18n.T("settings.protocol"), protoSel),
			),
		),
	)

	footer := container.NewGridWithColumns(3,
		widget.NewLabel(""),
		saveBtn,
		widget.NewLabel(""),
	)

	content := container.NewVBox(sections, footer)
	return container.NewVScroll(container.NewPadded(content))
}

// ---------------------------------------------------------------------------
// About tab
// ---------------------------------------------------------------------------

func (u *appUI) buildAbout() fyne.CanvasObject {
	ver := u.version
	if ver == "" {
		ver = "dev"
	}

	header := widget.NewLabel(fmt.Sprintf("RadKeys — %s", fmt.Sprintf(i18n.T("about.version"), ver)))
	header.TextStyle = fyne.TextStyle{Bold: true}

	desc := widget.NewLabel(i18n.T("about.description"))
	desc.Wrapping = fyne.TextWrapWord

	repoURL, _ := url.Parse("https://github.com/docg1701/radkeys")
	repo := widget.NewHyperlink("github.com/docg1701/radkeys", repoURL)
	repoLine := container.NewHBox(
		widget.NewLabel(i18n.T("about.author")),
		widget.NewLabel("|"),
		widget.NewLabel(i18n.T("about.license")),
		widget.NewLabel("|"),
		widget.NewLabel(i18n.T("about.repository")),
		repo,
	)

	stack := widget.NewLabel(i18n.T("about.stack"))
	stack.Wrapping = fyne.TextWrapWord

	langs := widget.NewLabel(i18n.T("about.i18n"))
	langs.Wrapping = fyne.TextWrapWord

	return container.NewVScroll(container.NewPadded(
		container.NewVBox(header, desc, stack, langs, repoLine),
	))
}

// ---------------------------------------------------------------------------
// Layout helpers
// ---------------------------------------------------------------------------

func section(title string, rows ...fyne.CanvasObject) fyne.CanvasObject {
	header := widget.NewLabel(title)
	header.TextStyle = fyne.TextStyle{Bold: true}
	items := []fyne.CanvasObject{header}
	items = append(items, rows...)
	return container.NewVBox(items...)
}

func variantFor(th fyne.Theme) fyne.ThemeVariant {
	if th == fyneTheme.DefaultTheme() {
		return fyne.CurrentApp().Settings().ThemeVariant()
	}
	bg := th.Color(fyneTheme.ColorNameBackground, fyneTheme.VariantDark)
	r, g, b, _ := bg.RGBA()
	if 0.2126*float64(r)+0.7152*float64(g)+0.0722*float64(b) > 0xffff*0.45 {
		return fyneTheme.VariantLight
	}
	return fyneTheme.VariantDark
}

func showFileDialog(parent fyne.Window, exts []string, onSelect func(path string)) {
	fd := dialog.NewFileOpen(func(rc fyne.URIReadCloser, err error) {
		if err != nil || rc == nil {
			return
		}
		onSelect(rc.URI().Path())
	}, parent)
	fd.SetFilter(storage.NewExtensionFileFilter(exts))
	fd.Resize(fyne.NewSize(900, 650))
	fd.Show()
}

func labeled(label string, input fyne.CanvasObject) fyne.CanvasObject {
	return container.NewVBox(widget.NewLabel(label), input)
}
