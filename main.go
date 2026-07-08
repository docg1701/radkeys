// Command radkeys loads radkeys.config.toml, opens the configured USB HID
// custom device, and runs the RadKeys UI. Without a device it falls back to
// the in-process mock (the UI is still drivable by mouse clicks).
package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/docg1701/radkeys/internal/config"
	"github.com/docg1701/radkeys/internal/hid"
	"github.com/docg1701/radkeys/internal/ui"
)

const configFileName = "radkeys.config.toml"

func main() {
	cfg, err := config.Load(configPath())
	if err != nil {
		log.Fatalf("radkeys: %v", err)
	}

	reader, err := hid.Open(cfg.App.Device)
	if err != nil {
		// No hardware: use the mock; the UI still works via mouse clicks.
		log.Printf("radkeys: %v; usando mock (clique nos botões da UI)", err)
		reader = hid.NewMock()
	}

	if err := ui.Run(cfg, reader); err != nil {
		log.Fatalf("radkeys: %v", err)
	}
}

// configPath resolves the config file: $RADKEYS_CONFIG, then the executable
// directory, then the current working directory.
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
