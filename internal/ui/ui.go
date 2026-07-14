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
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	fyneTheme "fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/docg1701/radkeys/internal/assets"
	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/hid"
	"github.com/docg1701/radkeys/internal/i18n"
	themes "github.com/docg1701/radkeys/internal/theme"
)

func Run(cfg *config.Config, configPath string, dev hid.Device, version string, mock bool, deviceOpenErr error) error {
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
		device:     dev,
		a:          a,
		win:        w,
		titleBase:  appName(cfg),
		cols:       cols,
		rows:       rows,
		version:    version,
		mock:       mock,
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
	u.tabs = tabs
	u.status = widget.NewLabel("")
	u.status.Hide()
	content := container.NewBorder(u.status, nil, nil, nil, tabs)
	w.SetContent(content)
	u.renderGrid()

	if u.mock {
		u.setStatus(i18n.T("status.mock_mode"))
	}

	if mock && deviceOpenErr != nil {
		msg := fmt.Sprintf(i18n.T("device.not_found_message"), deviceOpenErr.Error())
		dialog.ShowInformation(i18n.T("device.not_found_title"), msg, w)
	}

	// Fyne detects the OS theme variant asynchronously, so the initial render
	// may use the wrong variant (e.g. dark on a light OS). Re-apply the theme
	// colors when settings change — Fyne fires this once the variant settles,
	// which fixes the system-default "black background on reopen" regression
	// without requiring the user to re-select and save the theme.
	u.a.Settings().AddListener(func(s fyne.Settings) {
		th := s.Theme()
		if _, ok := th.(themes.CustomThemeMarker); ok {
			return // deterministic; no OS settling needed
		}
		v := variantFor(th, u.a.Settings().ThemeVariant())
		if u.previewBg != nil {
			u.previewBg.FillColor = th.Color(fyneTheme.ColorNameBackground, v)
			canvas.Refresh(u.previewBg)
		}
		u.renderGrid()
	})

	// One-shot firmware version check at connect (before the event loop).
	// Warns the user once if the firmware is outdated or its version is
	// unknown. Shown at startup — NOT on the HID event path.
	fwMaj, fwMin, fwErr := dev.Version()
	if hid.FirmwareOutdated(fwMaj, fwMin, fwErr == nil) {
		var fwMsg string
		if fwErr != nil {
			fwMsg = fmt.Sprintf(i18n.T("firmware.unknown_message"), hid.MinFirmwareMajor, hid.MinFirmwareMinor)
		} else {
			fwMsg = fmt.Sprintf(i18n.T("firmware.outdated_message"), fwMaj, fwMin, hid.MinFirmwareMajor, hid.MinFirmwareMinor)
		}
		dialog.ShowInformation(i18n.T("firmware.outdated_title"), fwMsg, w)
	}

	if err := dev.Open(); err != nil {
		return fmt.Errorf("hid: open: %w", err)
	}
	go u.pollHID()
	w.SetOnClosed(func() {
		u.closing.Store(true)
		if err := dev.Close(); err != nil {
			log.Printf("radkeys: device close failed: %v", err)
		}
	})
	w.ShowAndRun()
	return nil
}

type appUI struct {
	cfg         *config.Config
	configPath  string
	current     string   // current screen id
	stack       []string // parent screen ids for prev
	device      hid.Device
	a           fyne.App
	win         fyne.Window
	titleBase   string
	preview     *widget.Label
	previewText string
	version     string
	mock        bool
	closing     atomic.Bool
	status      *widget.Label
	flashTimer  *time.Timer
	tabs        *container.AppTabs
	cols        int
	rows        int
	keypad      *fyne.Container
	previewBg   *canvas.Rectangle
}

// appName returns the configured app name, defaulting to "RadKeys".
func appName(cfg *config.Config) string {
	if cfg.App.Name != "" {
		return cfg.App.Name
	}
	return "RadKeys"
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
//
// HID_FOCUS_INVARIANT: the fromUI=false path must NEVER raise, activate, or
// focus the RadKeys window — the only permitted focus grab is the initial
// w.ShowAndRun() at startup. ActionText/Copy/Navigate are silent (preview,
// clipboard, internal state + renderGrid); the device-keyboard actions
// (paste, select_all, select_line, line_start, line_end, backspace, delete)
// delegate the keystroke to the device via fireDeviceCommand so the already-
// focused window (the RIS) receives it without RadKeys taking focus. Do not
// call u.win.Show/ShowAndRun/SetContent/RequestFocus here. The fromUI dialog
// is the exception (the user clicked RadKeys). Enforced statically by
// TestHIDPathDoesNotActivateWindow.
func (u *appUI) press(row, col int, fromUI bool) {
	b, ok := u.currentScreen().ButtonAt(row, col)
	if !ok {
		if row < 0 || row >= u.rows || col < 0 || col >= u.cols {
			u.flashStatus(fmt.Sprintf(i18n.T("status.out_of_grid"), row, col, u.rows, u.cols))
			log.Printf("radkeys: device event out of grid bounds (row=%d, col=%d) for %dx%d", row, col, u.rows, u.cols)
		}
		return
	}
	switch b.Action {
	case config.ActionText:
		u.previewText = b.Content
		u.preview.SetText(b.Content)
	case config.ActionCopy:
		u.a.Clipboard().SetContent(u.previewText)
	case config.ActionPaste:
		u.fireDeviceCommand(config.ActionPaste, hid.CmdFirePaste, byte(hid.ModifierForOS()), fromUI)
	case config.ActionSelectAll:
		u.fireDeviceCommand(config.ActionSelectAll, hid.CmdSelectAll, byte(hid.ModifierForOS()), fromUI)
	case config.ActionSelectLine:
		u.fireDeviceCommand(config.ActionSelectLine, hid.CmdSelectLine, 0x00, fromUI)
	case config.ActionLineStart:
		u.fireDeviceCommand(config.ActionLineStart, hid.CmdLineStart, 0x00, fromUI)
	case config.ActionLineEnd:
		u.fireDeviceCommand(config.ActionLineEnd, hid.CmdLineEnd, 0x00, fromUI)
	case config.ActionBackspace:
		u.fireDeviceCommand(config.ActionBackspace, hid.CmdBackspace, 0x00, fromUI)
	case config.ActionDelete:
		u.fireDeviceCommand(config.ActionDelete, hid.CmdDelete, 0x00, fromUI)
	case config.ActionPrev:
		if len(u.stack) > 0 {
			u.current = u.stack[len(u.stack)-1]
			u.stack = u.stack[:len(u.stack)-1]
		}
	case config.ActionHome:
		u.current = u.cfg.Screens[0].ID
		u.stack = u.stack[:0]
	case config.ActionNavigate:
		if b.Target != u.current {
			u.stack = append(u.stack, u.current)
			u.current = b.Target
		}
	}
	u.renderGrid()
}

// fireDeviceCommand sends a device-keyboard command (paste, select_all,
// select_line, line_start, line_end, backspace, delete) to the firmware's HID
// keyboard interface. On-screen clicks (fromUI) are refused: clicking RadKeys
// gives it focus, so the device's keystroke would land in RadKeys itself — show
// a hint instead. The HID path (fromUI=false) never touches u.win, preserving
// the focus invariant (see TestHIDPathDoesNotActivateWindow).
func (u *appUI) fireDeviceCommand(action string, cmd hid.Command, arg byte, fromUI bool) {
	if fromUI {
		label := i18n.T("button." + action)
		dialog.ShowInformation(label, fmt.Sprintf(i18n.T("device_action.via_keypad_hint"), label), u.win)
		return
	}
	if err := u.device.FireCommand(cmd, arg); err != nil {
		u.flashStatus(fmt.Sprintf(i18n.T("status.device_command_failed"), err))
		log.Printf("radkeys: %s failed: %v", action, err)
	}
}

func (u *appUI) renderGrid() {
	s := u.currentScreen()
	totalSlots := u.cols * u.rows
	u.keypad.Objects = u.keypad.Objects[:0]
	th := u.a.Settings().Theme()
	v := variantFor(th, u.a.Settings().ThemeVariant())

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
	u.previewBg = canvas.NewRectangle(th.Color(fyneTheme.ColorNameBackground, variantFor(th, u.a.Settings().ThemeVariant())))
	scroll := container.NewVScroll(u.preview)
	return container.NewStack(u.previewBg, container.NewPadded(scroll))
}

// pollHID reads keypad events and forwards pressed events to the UI goroutine
// via press(..., fromUI=false). See the HID_FOCUS_INVARIANT in press: this path
// must never raise, activate, or focus the RadKeys window.
func (u *appUI) pollHID() {
	for ev := range u.device.Events() {
		if !ev.Pressed {
			continue
		}
		row, col := ev.Row, ev.Col
		fyne.Do(func() { u.press(row, col, false) })
	}
	if !u.closing.Load() {
		u.flashStatus(i18n.T("status.hid_read_failed"))
	}
}

// setStatus shows a persistent status message at the top of the window.
func (u *appUI) setStatus(msg string) {
	fyne.Do(func() {
		u.status.Text = msg
		u.status.Show()
		u.status.Refresh()
	})
}

// flashStatus shows a status message for 5 seconds, then hides it.
// The body runs on the UI goroutine via fyne.Do so u.flashTimer is
// never accessed concurrently; the timer callback bails out if the
// window is closing so we never touch the UI after shutdown.
func (u *appUI) flashStatus(msg string) {
	fyne.Do(func() {
		if u.flashTimer != nil {
			u.flashTimer.Stop()
		}
		u.status.Text = msg
		u.status.Show()
		u.status.Refresh()
		u.flashTimer = time.AfterFunc(5*time.Second, func() {
			if u.closing.Load() {
				return
			}
			fyne.Do(func() {
				u.status.Hide()
				u.status.Text = ""
			})
		})
	})
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
	vidEnt.Validator = hexUint16Validator
	pidEnt := widget.NewEntry()
	pidEnt.SetText(fmt.Sprintf("0x%04x", cfg.App.Device.ProductID))
	pidEnt.SetMinRowsVisible(1)
	pidEnt.Validator = hexUint16Validator
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
			vidEnt.SetValidationError(nil)
		} else {
			vidEnt.SetValidationError(fmt.Errorf("%s", i18n.T("settings.invalid_hex")))
			u.flashStatus(fmt.Sprintf("%s: %v", i18n.T("settings.vid"), err))
		}
		if v, err := strconv.ParseUint(strings.TrimPrefix(pidEnt.Text, "0x"), 16, 16); err == nil {
			cfg.App.Device.ProductID = uint16(v)
			pidEnt.SetValidationError(nil)
		} else {
			pidEnt.SetValidationError(fmt.Errorf("%s", i18n.T("settings.invalid_hex")))
			u.flashStatus(fmt.Sprintf("%s: %v", i18n.T("settings.pid"), err))
		}
		cfg.App.Device.Protocol = protoSel.Selected

		if err := u.cfg.Save(u.configPath); err != nil {
			dialog.ShowError(err, u.win)
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
			u.previewBg.FillColor = newTheme.Color(fyneTheme.ColorNameBackground, variantFor(newTheme, u.a.Settings().ThemeVariant()))
			canvas.Refresh(u.previewBg)
		}

		// Refresh tab labels and content.
		tabs := u.tabs
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

// variantFor returns the theme variant to use for manual Color() lookups.
// For RadKeys custom themes it is derived from the resolved background color,
// so it needs no app/global state. For the adaptive system/DefaultTheme it
// falls back to the variant supplied by the caller.
func variantFor(th fyne.Theme, fallback fyne.ThemeVariant) fyne.ThemeVariant {
	if _, ok := th.(themes.CustomThemeMarker); ok {
		return variantFromBackground(th)
	}
	return fallback
}

func variantFromBackground(th fyne.Theme) fyne.ThemeVariant {
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

func hexUint16Validator(s string) error {
	_, err := strconv.ParseUint(strings.TrimPrefix(s, "0x"), 16, 16)
	if err != nil {
		return fmt.Errorf("%s", i18n.T("settings.invalid_hex"))
	}
	return nil
}

func labeled(label string, input fyne.CanvasObject) fyne.CanvasObject {
	return container.NewVBox(widget.NewLabel(label), input)
}
