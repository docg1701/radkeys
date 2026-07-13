package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/hid"
	"github.com/docg1701/radkeys/internal/ui"
)

var Version = "0.4.0"

const configFileName = "radkeys.config.toml"

func main() {
	path := configPath()
	ensureConfig(path)
	cfg, err := config.Load(path)
	if err != nil {
		showConfigError(path, err)
		return
	}

	reader, err := hid.Open(cfg.App.Device)
	if err != nil {
		log.Printf("radkeys: %v; using mock (click UI buttons)", err)
		reader = hid.NewMock()
	}

	if err := ui.Run(cfg, configPath(), reader, Version); err != nil {
		log.Fatalf("radkeys: %v", err)
	}
}

func showConfigError(configPath string, err error) {
	a := app.New()
	w := a.NewWindow("RadKeys — Config Error")
	w.Resize(fyne.NewSize(700, 400))

	msg := widget.NewLabel(err.Error())
	msg.Wrapping = fyne.TextWrapWord

	editBtn := widget.NewButton("Open file to edit", func() {
		_ = exec.Command("xdg-open", configPath).Start()
	})
	editBtn.Importance = widget.HighImportance

	okBtn := widget.NewButton("Close", func() { w.Close() })

	content := container.NewVBox(
		widget.NewLabel("The configuration file contains an error:\n"),
		msg,
		widget.NewLabel("\nFix the error above and restart RadKeys."),
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

func ensureConfig(path string) {
	if _, err := os.Stat(path); err == nil {
		return
	}
	const tmpl = `[app]
name = "RadKeys"

[app.device]
vendor_id  = 0x1234
product_id = 0xABCD
protocol   = "radkeys-diy"

[app.layout]
columns = 4
rows    = 3

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
	_ = os.WriteFile(path, []byte(tmpl), 0o644)
}
