package editor

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/i18n"
	themes "github.com/docg1701/radkeys/internal/theme"
	"github.com/docg1701/radkeys/internal/widgetutil"
)

// buildAppSettings creates the first tab with app-wide fields.
func (e *Editor) buildAppSettings() fyne.CanvasObject {
	e.appSettings = container.NewVBox(
		e.buildAppearanceGroup(),
		e.buildGridGroup(),
		e.buildDeviceGroup(),
	)
	return container.NewVScroll(container.NewPadded(e.appSettings))
}

// buildAppearanceGroup groups name, radiologist, language, and theme.
func (e *Editor) buildAppearanceGroup() fyne.CanvasObject {
	name := widget.NewEntry()
	name.SetText(e.cfg.App.Name)
	name.OnChanged = e.setAppName

	rad := widget.NewEntry()
	rad.SetText(e.cfg.App.Radiologist)
	rad.OnChanged = e.setRadiologist

	lang := widget.NewSelect(i18n.Supported, nil)
	lang.SetSelected(e.cfg.App.Language)
	lang.OnChanged = e.setAppLanguage

	themeIDs, themeNames := e.themeOptions()
	theme := widget.NewSelect(themeNames, nil)
	theme.SetSelected(i18n.T("theme." + e.cfg.App.Theme.Preset))
	theme.OnChanged = func(choice string) {
		e.setAppTheme(themeIDs[indexOf(themeNames, choice)])
	}
	return widgetutil.Section(i18n.T("settings.group_appearance"),
		container.NewGridWithColumns(2,
			widgetutil.Labeled(i18n.T("editor.app_name"), name),
			widgetutil.Labeled(i18n.T("settings.radiologist"), rad),
		),
		container.NewGridWithColumns(2,
			widgetutil.Labeled(i18n.T("settings.language"), lang),
			widgetutil.Labeled(i18n.T("settings.theme"), theme),
		),
	)
}

// themeOptions returns theme ids and localized names.
func (e *Editor) themeOptions() (ids, names []string) {
	for _, p := range themes.Presets {
		ids = append(ids, p.ID())
		names = append(names, i18n.T("theme."+p.ID()))
	}
	return ids, names
}

// buildGridGroup groups the columns/rows steppers.
func (e *Editor) buildGridGroup() fyne.CanvasObject {
	cols := widget.NewSelect(gridSizes(), nil)
	cols.SetSelected(strconv.Itoa(e.cfg.App.Layout.Columns))
	cols.OnChanged = func(choice string) {
		if v, err := strconv.Atoi(choice); err == nil {
			e.resizeGrid(v, e.cfg.App.Layout.Rows)
		}
	}

	rows := widget.NewSelect(gridSizes(), nil)
	rows.SetSelected(strconv.Itoa(e.cfg.App.Layout.Rows))
	rows.OnChanged = func(choice string) {
		if v, err := strconv.Atoi(choice); err == nil {
			e.resizeGrid(e.cfg.App.Layout.Columns, v)
		}
	}

	return widgetutil.Section(i18n.T("editor.grid_size"),
		container.NewGridWithColumns(2,
			widgetutil.Labeled(i18n.T("settings.columns"), cols),
			widgetutil.Labeled(i18n.T("settings.rows"), rows),
		),
	)
}

// buildDeviceGroup groups VID, PID, and protocol.
func (e *Editor) buildDeviceGroup() fyne.CanvasObject {
	vid := widget.NewEntry()
	vid.SetText(fmt.Sprintf("0x%04x", e.cfg.App.Device.VendorID))
	vid.SetPlaceHolder(i18n.T("editor.hex_format"))
	vid.OnChanged = func(s string) { e.setVendorIDFromEntry(vid, s) }

	pid := widget.NewEntry()
	pid.SetText(fmt.Sprintf("0x%04x", e.cfg.App.Device.ProductID))
	pid.SetPlaceHolder(i18n.T("editor.hex_format"))
	pid.OnChanged = func(s string) { e.setProductIDFromEntry(pid, s) }

	proto := widget.NewSelect([]string{config.ProtocolDIY}, nil)
	proto.SetSelected(e.cfg.App.Device.Protocol)
	proto.OnChanged = e.setProtocol

	return widgetutil.Section(i18n.T("settings.group_device"),
		container.NewGridWithColumns(3,
			widgetutil.Labeled(i18n.T("settings.vid"), vid),
			widgetutil.Labeled(i18n.T("settings.pid"), pid),
			widgetutil.Labeled(i18n.T("settings.protocol"), proto),
		),
	)
}

// gridSizes returns the allowed 1–6 values as strings.
func gridSizes() []string {
	return []string{"1", "2", "3", "4", "5", "6"}
}

// indexOf returns the index of choice in options, or -1.
func indexOf(options []string, choice string) int {
	for i, o := range options {
		if o == choice {
			return i
		}
	}
	return -1
}
