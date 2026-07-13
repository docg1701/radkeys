package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/docg1701/radkeys/internal/assets"
	"github.com/docg1701/radkeys/internal/editor"
	"github.com/docg1701/radkeys/internal/i18n"
	themes "github.com/docg1701/radkeys/internal/theme"
)

func main() {
	a := app.NewWithID("com.docg1701.radkeys-config")
	a.Settings().SetTheme(themes.NewCustomTheme(themes.Presets[0]))
	a.SetIcon(fyne.NewStaticResource("icon.png", assets.IconPNG))

	path := editor.StartupPath()
	cfg, err := editor.LoadStartup(path)
	if err != nil {
		log.Printf("radkeys-config: %v; starting with new config", err)
	}
	if p := a.Preferences().String("lastFile"); p != "" {
		path = p
		if loaded, err := editor.LoadStartup(p); err == nil {
			cfg = loaded
		}
	}

	i18n.SetLanguage(cfg.App.Language)

	w := a.NewWindow(i18n.T("editor.title"))
	w.Resize(fyne.NewSize(1100, 760))
	w.SetIcon(fyne.NewStaticResource("icon.png", assets.IconPNG))

	ed := editor.NewEditor(a, w, cfg, path)
	ed.Run()
}
