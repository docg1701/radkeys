package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
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

var Version = "0.16.4"

var flagConfig = flag.String("c", "", "Path to radkeys.config.toml")

func main() {
	flag.StringVar(flagConfig, "config", "", "Path to radkeys.config.toml")
	flag.Parse()

	path := *flagConfig
	if path == "" {
		path = config.StartupPath()
	}
	if err := ensureConfig(path); err != nil {
		log.Fatalf("radkeys: %v", err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		showConfigError(path, err)
		return
	}

	dev, devOpenErr := hid.Open(cfg.App.Device)
	isMock := devOpenErr != nil
	if isMock {
		log.Printf("radkeys: %v", devOpenErr)
		dev = hid.NewMock()
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
