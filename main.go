package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/hid"
	"github.com/docg1701/radkeys/internal/i18n"
	"github.com/docg1701/radkeys/internal/ui"
)

var Version = "0.12.0"

const configFileName = "radkeys.config.toml"

func main() {
	path := configPath()
	if err := ensureConfig(path); err != nil {
		log.Fatalf("radkeys: %v", err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		showConfigError(path, err)
		return
	}

	dev, err := hid.Open(cfg.App.Device)
	isMock := false
	if err != nil {
		log.Printf("radkeys: %v; using mock (click UI buttons)", err)
		dev = hid.NewMock()
		isMock = true
	}

	if err := ui.Run(cfg, path, dev, Version, isMock); err != nil {
		log.Fatalf("radkeys: %v", err)
	}
}

func showConfigError(configPath string, err error) {
	a := app.New()
	w := a.NewWindow(i18n.T("error.config_title"))
	w.Resize(fyne.NewSize(700, 400))

	msg := widget.NewLabel(err.Error())
	msg.Wrapping = fyne.TextWrapWord

	editBtn := widget.NewButton(i18n.T("error.open_file"), func() {
		if err := openConfigEditor(configPath); err != nil {
			dialog.ShowError(err, w)
		}
	})
	editBtn.Importance = widget.HighImportance

	okBtn := widget.NewButton(i18n.T("button.close"), func() { w.Close() })

	content := container.NewVBox(
		widget.NewLabel(i18n.T("error.config_message")+"\n"),
		msg,
		widget.NewLabel("\n"+i18n.T("error.config_fix")),
		editBtn,
		okBtn,
	)

	w.SetContent(content)
	w.ShowAndRun()
}

func configPath() string {
	if p := os.Getenv("RADKEYS_CONFIG"); p != "" {
		return p
	}
	if exec, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exec), configFileName)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return configFileName
}

func ensureConfig(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	const tmpl = `[app]
name = "RadKeys"

[app.device]
vendor_id  = 0x1234
product_id = 0xABCD
protocol   = "radkeys-diy"

[app.layout]
columns = 6
rows    = 6

[app.theme]
preset = "system"

[[screens]]
id = "root"
name = "Home"

[[screens.buttons]]
row = 0
col = 0
label = "Example"
action = "text"
content = "Example phrase."
`
	if err := os.WriteFile(path, []byte(tmpl), 0o600); err != nil {
		return fmt.Errorf("cannot create default config %s: %w", path, err)
	}
	return nil
}

// openConfigEditor opens the config file in the platform's default editor.
func openConfigEditor(path string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", path).Start()
	case "windows":
		return exec.Command("cmd", "/c", "start", "", path).Start()
	default:
		return exec.Command("xdg-open", path).Start()
	}
}
