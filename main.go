package main

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const configFileName = "radkeys.config.toml"

func main() {
	a := app.New()
	w := a.NewWindow("RadKeys")
	w.SetFixedSize(true)
	w.Resize(fyne.NewSize(800, 600))

	execPath, err := os.Executable()
	if err != nil {
		fmt.Println("Could not determine executable path:", err)
		os.Exit(1)
	}
	configPath := filepath.Join(filepath.Dir(execPath), configFileName)

	info := widget.NewLabel(fmt.Sprintf("RadKeys prototype\nConfig path: %s", configPath))
	info.Wrapping = fyne.TextWrapWord

	w.SetContent(container.NewVBox(
		widget.NewLabel("RadKeys"),
		info,
	))

	w.ShowAndRun()
}
