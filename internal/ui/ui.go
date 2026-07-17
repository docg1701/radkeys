// Package ui renders RadKeys: preview on top half, virtual keypad on bottom.
// Three tabs: shortcuts, settings, about.
package ui

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"slices"
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
	"github.com/docg1701/radkeys/internal/widgetutil"
)

func Run(cfg *config.Config, configPath string, dev hid.Device, version string, mock bool) error {
	u, err := buildMainUI(cfg, configPath, dev, version, mock)
	if err != nil {
		return err
	}

	u.a.Settings().AddListener(u.osThemeSettledListener)
	u.checkFirmware(u.win)
	if err := u.startHIDLoop(); err != nil {
		return err
	}
	u.win.ShowAndRun()
	return nil
}

// buildMainUI creates the Fyne app, window, and initial tab layout.
func buildMainUI(cfg *config.Config, configPath string, dev hid.Device, version string, mock bool) (*appUI, error) {
	a := app.New()
	a.Settings().SetTheme(resolveFullTheme(cfg))
	a.SetIcon(fyne.NewStaticResource("icon.png", appIconData(cfg)))
	i18n.SetLanguage(cfg.App.Language)

	w := a.NewWindow(fmt.Sprintf("RadKeys — %s", cfg.App.Radiologist))
	w.Resize(fyne.NewSize(1280, 800))
	w.SetIcon(a.Icon())

	u := &appUI{
		cfg:        cfg,
		configPath: configPath,
		device:     dev,
		a:          a,
		win:        w,
		titleBase:  appName(cfg),
		cols:       cfg.App.Layout.Columns,
		rows:       cfg.App.Layout.Rows,
		version:    version,
		mock:       mock,
		preview:    widget.NewLabel(i18n.T("preview.placeholder")),
		current:    cfg.Screens[0].ID,
	}
	u.preview.Wrapping = fyne.TextWrapWord
	u.preview.TextStyle = fyne.TextStyle{Monospace: true}
	u.keypad = container.NewGridWithColumns(u.cols)
	u.status = widget.NewLabel("")
	u.status.Hide()

	u.rebuildTabs()

	if u.mock {
		u.setStatus(i18n.T("status.mock_mode"))
	}
	return u, nil
}

// osThemeSettledListener re-applies theme colors once Fyne settles the OS
// variant, fixing the system-default "black background on reopen" regression.
func (u *appUI) osThemeSettledListener(s fyne.Settings) {
	th := s.Theme()
	if _, ok := th.(themes.CustomThemeMarker); ok {
		return // deterministic; no OS settling needed
	}
	v := variantFor(th, u.a.Settings().ThemeVariant())
	if u.previewBg != nil {
		u.previewBg.FillColor = th.Color(fyneTheme.ColorNameBackground, v)
		canvas.Refresh(u.previewBg)
	}
	if u.navMap != nil {
		u.navMap.SetTheme(th, v)
	}
	u.renderGrid()
}

// checkFirmware warns the user once if the firmware is outdated or unknown.
// It is called at startup — NOT on the HID event path.
func (u *appUI) checkFirmware(w fyne.Window) {
	fwMaj, fwMin, fwErr := u.device.Version()
	if !hid.FirmwareOutdated(fwMaj, fwMin, fwErr == nil) {
		return
	}
	var fwMsg string
	if fwErr != nil {
		fwMsg = fmt.Sprintf(i18n.T("firmware.unknown_message"), hid.MinFirmwareMajor, hid.MinFirmwareMinor)
	} else {
		fwMsg = fmt.Sprintf(i18n.T("firmware.outdated_message"), fwMaj, fwMin, hid.MinFirmwareMajor, hid.MinFirmwareMinor)
	}
	dialog.ShowInformation(i18n.T("firmware.outdated_title"), fwMsg, w)
}

// startHIDLoop opens the device, starts the event loop, and wires the close
// handler. The pollHID path must preserve the HID_FOCUS_INVARIANT.
func (u *appUI) startHIDLoop() error {
	if err := u.device.Open(); err != nil {
		return fmt.Errorf("hid: open: %w", err)
	}
	go u.pollHID()
	u.win.SetOnClosed(func() {
		u.closing.Store(true)
		if err := u.device.Close(); err != nil {
			log.Printf("radkeys: device close failed: %v", err)
		}
	})
	return nil
}

type appUI struct {
	cfg             *config.Config
	configPath      string
	current         string   // current screen id
	stack           []string // parent screen ids for prev
	device          hid.Device
	a               fyne.App
	win             fyne.Window
	titleBase       string
	preview         *widget.Label
	previewText     string
	version         string
	mock            bool
	closing         atomic.Bool
	status          *widget.Label
	flashTimer      *time.Timer
	tabs            *container.AppTabs
	cols            int
	rows            int
	keypad          *fyne.Container
	previewBg       *canvas.Rectangle // created once in buildMainUI, mutated only in applySettings
	navMap          *mapWidget
	mapVisible      bool            // true when panel is shown
	mapScroll       *fyne.Container // cached Max wrapper around navMap
	breadcrumbLabel *widget.Label
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
	if def, ok := deviceCommands[b.Action]; ok {
		u.fireDeviceCommand(b.Action, def.cmd, def.arg(), fromUI)
	} else {
		switch b.Action {
		case config.ActionText:
			u.previewText = b.Content
			u.preview.SetText(b.Content)
		case config.ActionExec:
			if err := runExec(b.Content); err != nil {
				log.Printf("radkeys: exec failed: %v", err)
			}
		case config.ActionCopy:
			u.a.Clipboard().SetContent(u.previewText)
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
	}
	u.renderGrid()
}

// deviceCommand describes a keystroke command sent to the device's HID
// keyboard interface. arg is a thunk so that OS-dependent modifiers are
// resolved at call time (paste/select_all need Ctrl on Linux/Windows and
// Cmd on macOS).
type deviceCommand struct {
	cmd hid.Command
	arg func() byte
}

// deviceCommands maps device-keyboard actions to the command and dynamic
// modifier sent to the firmware. It replaces the long switch in press().
var deviceCommands = map[string]deviceCommand{
	config.ActionPaste:      {cmd: hid.CmdFirePaste, arg: func() byte { return byte(hid.ModifierForOS()) }},
	config.ActionSelectAll:  {cmd: hid.CmdSelectAll, arg: func() byte { return byte(hid.ModifierForOS()) }},
	config.ActionSelectLine: {cmd: hid.CmdSelectLine, arg: func() byte { return 0x00 }},
	config.ActionLineStart:  {cmd: hid.CmdLineStart, arg: func() byte { return 0x00 }},
	config.ActionLineEnd:    {cmd: hid.CmdLineEnd, arg: func() byte { return 0x00 }},
	config.ActionBackspace:  {cmd: hid.CmdBackspace, arg: func() byte { return 0x00 }},
	config.ActionDelete:     {cmd: hid.CmdDelete, arg: func() byte { return 0x00 }},
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

// runExec starts a shell command in the background. It picks the right
// shell per OS: cmd /c on Windows, bash -c elsewhere.
func runExec(command string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("bash", "-c", command)
	}
	return cmd.Start()
}

func (u *appUI) renderGrid() {
	if u.navMap != nil {
		u.navMap.SetCurrentScreen(u.current)
	}
	if u.breadcrumbLabel != nil {
		u.breadcrumbLabel.SetText(u.breadcrumb())
	}
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

// settingsWidgets holds the editable controls created by buildSettings.
type settingsWidgets struct {
	radEnt         *widget.Entry
	langSel        *widget.Select
	themeSel       *widget.Select
	themeIDs       []string
	colsEnt        *widget.Entry
	rowsEnt        *widget.Entry
	configLbl      *widget.Label
	chooseBtn      *widget.Button
	vidEnt         *widget.Entry
	pidEnt         *widget.Entry
	protoSel       *widget.Select
	iconPreview    *canvas.Image
	iconBrowseBtn  *widget.Button
	customIconPath *string
}

func (u *appUI) buildSettings() fyne.CanvasObject {
	w := u.buildSettingsWidgets()
	return u.buildSettingsSections(w)
}

func (u *appUI) buildSettingsWidgets() *settingsWidgets {
	cfg := u.cfg
	w := &settingsWidgets{}

	w.radEnt = widget.NewEntry()
	w.radEnt.SetText(cfg.App.Radiologist)

	w.langSel = widget.NewSelect(i18n.Supported, nil)
	w.langSel.SetSelected(cfg.App.Language)

	ids, names := u.themeOptions()
	w.themeIDs = ids
	w.themeSel = widget.NewSelect(names, nil)
	w.themeSel.SetSelectedIndex(slices.Index(w.themeIDs, cfg.App.Theme.Preset))

	w.colsEnt = widget.NewEntry()
	w.colsEnt.SetText(strconv.Itoa(cfg.App.Layout.Columns))
	w.rowsEnt = widget.NewEntry()
	w.rowsEnt.SetText(strconv.Itoa(cfg.App.Layout.Rows))

	w.configLbl = widget.NewLabel(u.configPath)
	w.configLbl.Wrapping = fyne.TextWrapWord
	w.chooseBtn = widget.NewButton(i18n.T("settings.browse"), func() {
		showFileDialog(u.win, []string{".toml"}, func(path string) {
			cfg, err := config.Load(path)
			if err != nil {
				dialog.ShowError(err, u.win)
				return
			}
			u.configPath = path
			u.cfg = cfg
			i18n.SetLanguage(cfg.App.Language)
			u.applySettings(cfg)
			u.rebuildTabs()
		})
	})
	w.chooseBtn.Importance = widget.MediumImportance

	w.vidEnt = widget.NewEntry()
	w.vidEnt.SetText(fmt.Sprintf("0x%04x", cfg.App.Device.VendorID))
	w.vidEnt.SetMinRowsVisible(1)
	w.vidEnt.Validator = hexUint16Validator
	w.pidEnt = widget.NewEntry()
	w.pidEnt.SetText(fmt.Sprintf("0x%04x", cfg.App.Device.ProductID))
	w.pidEnt.SetMinRowsVisible(1)
	w.pidEnt.Validator = hexUint16Validator

	w.protoSel = widget.NewSelect([]string{config.ProtocolDIY}, nil)
	w.protoSel.SetSelected(cfg.App.Device.Protocol)

	customIconPath := cfg.App.Theme.Icon
	w.customIconPath = &customIconPath
	w.iconPreview = canvas.NewImageFromResource(fyne.NewStaticResource("icon.png", appIconData(cfg)))
	w.iconPreview.SetMinSize(fyne.NewSize(48, 48))
	w.iconPreview.FillMode = canvas.ImageFillContain
	w.iconBrowseBtn = widget.NewButton(i18n.T("settings.browse"), func() {
		showFileDialog(u.win, []string{".png"}, func(path string) {
			*w.customIconPath = path
			data, err := os.ReadFile(path)
			if err != nil {
				log.Printf("radkeys: cannot read icon %q: %v", path, err)
				return
			}
			w.iconPreview.Resource = fyne.NewStaticResource("custom.png", data)
			w.iconPreview.Refresh()
		})
	})
	w.iconBrowseBtn.Importance = widget.MediumImportance

	return w
}

func (u *appUI) buildSettingsSections(w *settingsWidgets) fyne.CanvasObject {
	sections := container.NewVBox(
		widgetutil.Section(i18n.T("settings.group_config"),
			container.NewGridWithColumns(3,
				container.NewBorder(nil, nil, widget.NewLabel(i18n.T("settings.config_file")), nil, w.configLbl),
				w.chooseBtn,
				widget.NewLabel(""),
			),
		),
		widgetutil.Section(i18n.T("settings.group_appearance"),
			container.NewGridWithColumns(3,
				widgetutil.Labeled(i18n.T("settings.radiologist"), w.radEnt),
				widgetutil.Labeled(i18n.T("settings.language"), w.langSel),
				widgetutil.Labeled(i18n.T("settings.theme"), w.themeSel),
			),
			container.NewGridWithColumns(3,
				widget.NewLabel(i18n.T("settings.icon")),
				w.iconPreview,
				w.iconBrowseBtn,
			),
		),
		widgetutil.Section(i18n.T("settings.group_device"),
			container.NewGridWithColumns(3,
				widgetutil.Labeled(i18n.T("settings.columns"), w.colsEnt),
				widgetutil.Labeled(i18n.T("settings.rows"), w.rowsEnt),
				widget.NewLabel(""),
			),
			container.NewGridWithColumns(3,
				widgetutil.Labeled(i18n.T("settings.vid"), w.vidEnt),
				widgetutil.Labeled(i18n.T("settings.pid"), w.pidEnt),
				widgetutil.Labeled(i18n.T("settings.protocol"), w.protoSel),
			),
		),
	)

	footer := container.NewGridWithColumns(3,
		widget.NewLabel(""),
		widget.NewButton(i18n.T("settings.save"), u.makeSaveHandler(w)),
		widget.NewLabel(""),
	)

	return container.NewVScroll(container.NewPadded(container.NewVBox(sections, footer)))
}

func (u *appUI) makeSaveHandler(w *settingsWidgets) func() {
	return func() {
		cfg := u.cfg
		cfg.App.Radiologist = w.radEnt.Text
		cfg.App.Language = w.langSel.Selected
		cfg.App.Theme.Icon = *w.customIconPath
		if selIdx := w.themeSel.SelectedIndex(); selIdx >= 0 && selIdx < len(w.themeIDs) {
			cfg.App.Theme.Preset = w.themeIDs[selIdx]
		}
		if v, err := strconv.Atoi(w.colsEnt.Text); err == nil && v > 0 && v <= 6 {
			cfg.App.Layout.Columns = v
		} else {
			cfg.App.Layout.Columns = 1
			w.colsEnt.SetText("1")
		}
		if v, err := strconv.Atoi(w.rowsEnt.Text); err == nil && v > 0 && v <= 6 {
			cfg.App.Layout.Rows = v
		} else {
			cfg.App.Layout.Rows = 1
			w.rowsEnt.SetText("1")
		}
		if v, err := config.ParseHexUint16(w.vidEnt.Text); err == nil {
			cfg.App.Device.VendorID = v
			w.vidEnt.SetValidationError(nil)
		} else {
			w.vidEnt.SetValidationError(fmt.Errorf("%s", i18n.T("settings.invalid_hex")))
			u.flashStatus(fmt.Sprintf("%s: %v", i18n.T("settings.vid"), err))
		}
		if v, err := config.ParseHexUint16(w.pidEnt.Text); err == nil {
			cfg.App.Device.ProductID = v
			w.pidEnt.SetValidationError(nil)
		} else {
			w.pidEnt.SetValidationError(fmt.Errorf("%s", i18n.T("settings.invalid_hex")))
			u.flashStatus(fmt.Sprintf("%s: %v", i18n.T("settings.pid"), err))
		}
		cfg.App.Device.Protocol = w.protoSel.Selected

		if err := u.cfg.Save(u.configPath); err != nil {
			dialog.ShowError(err, u.win)
			return
		}

		u.applySettings(u.cfg)
		u.rebuildTabs()
	}
}

// applySettings re-applies language, theme, icon, title, and grid layout after
// the user saves the settings tab.
func (u *appUI) applySettings(cfg *config.Config) {
	u.cfg = cfg
	u.navMap = nil // force rebuild on next shortcutsTab — graph must reflect current config
	i18n.SetLanguage(cfg.App.Language)
	u.win.SetTitle(fmt.Sprintf("%s — %s", u.titleBase, cfg.App.Radiologist))

	iconRes := fyne.NewStaticResource("icon.png", appIconData(cfg))
	u.a.SetIcon(iconRes)
	u.win.SetIcon(iconRes)

	newTheme := resolveFullTheme(cfg)
	u.a.Settings().SetTheme(newTheme)
	v := variantFor(newTheme, u.a.Settings().ThemeVariant())
	if u.previewBg != nil {
		u.previewBg.FillColor = newTheme.Color(fyneTheme.ColorNameBackground, v)
		canvas.Refresh(u.previewBg)
	}
	if u.navMap != nil {
		u.navMap.SetTheme(newTheme, v)
	}

	if cfg.App.Layout.Columns != u.cols || cfg.App.Layout.Rows != u.rows {
		u.cols = cfg.App.Layout.Columns
		u.rows = cfg.App.Layout.Rows
		u.keypad = container.NewGridWithColumns(u.cols)
	}

	u.current = cfg.Screens[0].ID
	u.stack = u.stack[:0]
	u.renderGrid()
}

// rebuildTabs creates a fresh AppTabs container and replaces the window
// content. This avoids the fragile pattern of mutating tabs.Items after
// SetContent.
func (u *appUI) rebuildTabs() {
	selectedIdx := 0
	if u.tabs != nil {
		selectedIdx = u.tabs.SelectedIndex()
	}

	previewArea := u.previewBox()
	keypadArea := container.NewPadded(u.keypad)
	main := container.NewVSplit(previewArea, keypadArea)
	shortcuts := u.shortcutsTab(main)

	u.tabs = container.NewAppTabs(
		container.NewTabItem(i18n.T("tab.shortcuts"), shortcuts),
		container.NewTabItem(i18n.T("tab.settings"), u.buildSettings()),
		container.NewTabItem(i18n.T("tab.about"), u.buildAbout()),
	)
	u.tabs.SelectIndex(selectedIdx)
	u.win.SetContent(container.NewBorder(u.headerBar(), nil, nil, nil, u.tabs))
}

// shortcutsTab returns the shortcuts tab content: the preview/keypad
// vertical split, optionally with the map panel on the right.
func (u *appUI) shortcutsTab(main *container.Split) fyne.CanvasObject {
	if u.navMap == nil {
		u.navMap = newMapWidget(u.cfg)
		u.mapScroll = container.NewStack(u.navMap)
	}
	th, v := u.a.Settings().Theme(), u.a.Settings().ThemeVariant()
	u.navMap.SetTheme(th, v)

	icon := fyneTheme.NavigateNextIcon()
	if !u.mapVisible {
		icon = fyneTheme.NavigateBackIcon()
	}
	toggle := widget.NewButtonWithIcon("", icon, func() {
		u.toggleMap()
	})
	toggle.Importance = widget.LowImportance

	if u.mapVisible {
		left := container.NewBorder(nil, nil, nil, toggle, main)
		// Enforce that the left panel never shrinks below 50% of the
		// window width, so the map panel is capped at 50% even when
		// the user drags the HSplit divider.
		guard := newMinWidthBox(left, func() float32 {
			return float32(u.win.Canvas().Size().Width) * 0.5
		})
		split := container.NewHSplit(guard, u.mapScroll)
		split.Offset = u.mapSplitOffset()
		return split
	}
	return container.NewBorder(nil, nil, nil, toggle, main)
}

// toggleMap collapses or expands the side panel in-place (no full tab rebuild).
func (u *appUI) toggleMap() {
	u.mapVisible = !u.mapVisible
	if u.tabs == nil {
		return
	}
	previewArea := u.previewBox()
	keypadArea := container.NewPadded(u.keypad)
	main := container.NewVSplit(previewArea, keypadArea)
	items := u.tabs.Items
	if len(items) > 0 {
		items[0].Content = u.shortcutsTab(main)
		u.tabs.Refresh()
	}
}

// headerBar is the top-of-window row: breadcrumb fills the center,
// device-status message sits on the right. Border layout keeps them
// apart naturally — no separator needed.
func (u *appUI) headerBar() fyne.CanvasObject {
	u.breadcrumbLabel = widget.NewLabel(u.breadcrumb())
	u.breadcrumbLabel.TextStyle = fyne.TextStyle{Italic: true}
	return container.NewBorder(nil, nil, nil, u.status, u.breadcrumbLabel)
}

// breadcrumb returns the ">"-separated path of screen names from the back
// stack to the current screen, e.g. "Home > RM > Medicina Interna > Abdome".
// Unknown ids fall back to the raw id (never empty — visible in the UI).
func (u *appUI) breadcrumb() string {
	names := make([]string, 0, len(u.stack)+1)
	idToName := make(map[string]string, len(u.cfg.Screens))
	for _, s := range u.cfg.Screens {
		idToName[s.ID] = s.Name
	}
	for _, id := range u.stack {
		if name, ok := idToName[id]; ok {
			names = append(names, name)
		} else {
			names = append(names, id)
		}
	}
	cur := u.current
	if name, ok := idToName[cur]; ok {
		names = append(names, name)
	} else {
		names = append(names, cur)
	}
	return strings.Join(names, " > ")
}

// themeOptions returns theme IDs and their localized names for the settings
// dropdown.
func (u *appUI) themeOptions() (ids, names []string) {
	for _, p := range themes.Presets {
		ids = append(ids, p.ID())
		names = append(names, i18n.T("theme."+p.ID()))
	}
	return ids, names
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

// variantFor returns the theme variant to use for manual Color() lookups.
// For RadKeys custom themes it is derived from the resolved background color,
// so it needs no app/global state. For the adaptive system/DefaultTheme it
// falls back to the variant supplied by the caller.
const mapOffsetExpanded = 0.75

// mapSplitOffset returns the HSplit offset that gives the map its natural
// width, capped at 50% of the window width so the main content never gets
// squeezed below half the window.
func (u *appUI) mapSplitOffset() float64 {
	winW := float64(u.win.Canvas().Size().Width)
	if winW <= 0 || u.navMap == nil {
		return mapOffsetExpanded
	}
	mapW := float64(u.navMap.MinSize().Width)
	if mapW > winW*0.5 {
		mapW = winW * 0.5
	}
	offset := 1.0 - mapW/winW
	if offset < 0.5 {
		return 0.5
	}
	return offset
}

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
	_, err := config.ParseHexUint16(s)
	return err
}

// minWidthBox wraps a child widget and enforces a floor on its MinSize
// width. The getMinW callback is invoked on every MinSize() call so the
// constraint stays in sync with the window size.
type minWidthBox struct {
	widget.BaseWidget
	child   fyne.CanvasObject
	getMinW func() float32
}

func newMinWidthBox(child fyne.CanvasObject, getMinW func() float32) *minWidthBox {
	b := &minWidthBox{child: child, getMinW: getMinW}
	b.ExtendBaseWidget(b)
	return b
}

func (b *minWidthBox) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(b.child)
}

func (b *minWidthBox) MinSize() fyne.Size {
	s := b.child.MinSize()
	if minW := b.getMinW(); s.Width < minW {
		s.Width = minW
	}
	return s
}
