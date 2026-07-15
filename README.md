# RadKeys

Portable shortcut deck for radiology reports — copy pre-written phrases to
the clipboard without stealing focus from the RIS/PACS.

![RadKeys Screenshot](screenshot.png)

[![License](https://img.shields.io/badge/license-OCL%20v1.1-blue)](LICENSE)

## What it is

RadKeys is a companion app for radiologists. You connect a custom keypad
(6×6 = 36 buttons) via USB, and each button inserts a pre-written report
template. No keyboard shortcuts to memorize and no focus stealing from
your RIS/PACS — the radiologist just presses a keypad button.

You write your report templates once in a config file. The app shows them
in a grid that mirrors your physical keypad. Press a physical button → the
phrase appears on screen → press Copy → paste into the RIS. That's it.

Works on Linux, Windows, and macOS. One executable, one config file, zero install.
Everything else (icon, translations, themes) is embedded in the binary.

## Features

- 36 configurable buttons (6×6) organized in navigable screens
- 13 actions: text templates, clipboard, navigation, editing keystrokes, and bash command execution
- Paste via the device's USB keyboard — no focus stealing, no host-side software
- 7 languages, 13 color themes, custom icon
- Single binary per OS (Linux + Windows; macOS builds from source)
- Mock mode — run without hardware, the UI works via mouse clicks
- One-shot firmware version check on connect (warns if outdated)

## How it works

The RP2040-Zero is a **composite USB device** with two HID interfaces:

- **Vendor HID** — sends `[row, col]` button events to the host (background,
  no focus stealing).
- **HID keyboard** — sends Ctrl+V (Linux/Windows) or Cmd+V (macOS) to the
  already-focused window when the host commands a paste.

The app is a **configurator**: all configuration (phrases, button actions) lives
in `radkeys.config.toml`. The device is flashed **once** at the factory and
never reflashed for configuration. Paste goes through the device's keyboard
interface, so the RIS keeps focus — the app never injects keystrokes into the
OS. At connect, the app checks the firmware version once and warns if it is
outdated or unknown.

## Download

Get the latest release from [Releases](../../releases). Each release includes:

| File | Platform |
|------|----------|
| `radkeys-linux-amd64` | Linux x86_64 (main app) |
| `radkeys-windows-amd64.exe` | Windows x86_64 (main app) |
| `radkeys-config-linux-amd64` | Linux x86_64 (config editor) |
| `radkeys-config-windows-amd64.exe` | Windows x86_64 (config editor) |
| `radkeys.config.toml` | Config template (all platforms) |

**macOS**: binary not provided (cross-compile from Linux is impossible — needs
Apple's proprietary SDK). macOS is supported in code: build from source on a
Mac following the instructions below. Paste is cross-platform via the device
(no per-OS keystroke injection), so macOS works the same as Linux/Windows.

Put the binary and `radkeys.config.toml` in the same directory and run.
Use `-c` to specify a different config path:
```bash
./radkeys-linux-amd64 -c ~/my-config.toml
```

Without a hardware device, the app runs in mock mode — the UI works via mouse
clicks.

## Usage

1. Edit `radkeys.config.toml` to add your phrases. The file is heavily
   commented — a human or LLM can read it and generate a custom config
   following the rules in the comments.
2. Connect your USB device with the DIY keypad (RP2040-Zero).
3. Run RadKeys.
4. Press a text button → phrase appears in the preview.
5. Press Copy → phrase goes to the clipboard.
6. Press Paste → the device sends Ctrl+V (Linux/Windows) or Cmd+V (macOS)
   to the focused window (your RIS/PACS) as a USB keyboard. The phrase appears
   at the cursor position. Editing commands (select_all, select_line, line_start,
   line_end, backspace, delete) are also sent by the device keyboard — they
   go to the currently focused window without stealing focus. No host-side
   software is needed — the device is the keyboard. RadKeys never steals focus.

   Press an `exec` button to run an arbitrary bash command (fire-and-forget, user
   permissions).

   To configure the keypad visually, use the `radkeys-config` binary instead of
   hand-editing the TOML file.

The radiologist never touches the keyboard.

## Build from source

### Prerequisites

| Dependency | Linux | Windows | macOS |
|------------|-------|---------|-------|
| **Go** 1.24+ | `sudo apt install golang-go` | [go.dev/dl](https://go.dev/dl/) | [go.dev/dl](https://go.dev/dl/) |
| **GCC** (CGO) | `sudo apt install gcc` | [MinGW-w64](https://www.mingw-w64.org/) | `xcode-select --install` |
| **Fyne** | `sudo apt install libgl1-mesa-dev xorg-dev libxxf86vm-dev` | — | — |
| **HIDAPI** | `sudo apt install libudev-dev` | — | IOKit (system) |

### Build (native)

```bash
# Linux
CGO_ENABLED=1 go build -tags flatpak -o dist/radkeys-linux-amd64 .
CGO_ENABLED=1 go build -tags flatpak -o dist/radkeys-config-linux-amd64 ./cmd/radkeys-config

# Windows (on Windows, or cross-compile from Linux with mingw)
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -o dist/radkeys-windows-amd64.exe .
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -o dist/radkeys-config-windows-amd64.exe ./cmd/radkeys-config

# macOS Intel (on a Mac — cross-compile from Linux is impossible)
CGO_ENABLED=1 go build -o dist/radkeys-macos-amd64 .
CGO_ENABLED=1 go build -o dist/radkeys-config-macos-amd64 ./cmd/radkeys-config

# macOS Apple Silicon (on a Mac)
CGO_ENABLED=1 GOARCH=arm64 go build -o dist/radkeys-macos-arm64 .
CGO_ENABLED=1 GOARCH=arm64 go build -o dist/radkeys-config-macos-arm64 ./cmd/radkeys-config
```

### Cross-compile from Linux (Windows only)

```bash
sudo apt install -y gcc-mingw-w64
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -o dist/radkeys-windows-amd64.exe .
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -o dist/radkeys-config-windows-amd64.exe ./cmd/radkeys-config
```

### Test

```bash
go test ./... -v
```

### Runtime dependencies (end user)

The device enumerates as a standard composite HID device (vendor + keyboard).
No host-side software is needed for paste — the device is the USB keyboard.

| Dependency | Linux | Windows | macOS |
|------------|-------|---------|-------|
| HID access (hidapi) | libudev (system, via systemd) | bundled with the binary | IOKit (system) |

## Hardware

| Option | Device | Keys | Cost |
|--------|--------|------|------|
| DIY | RP2040-Zero + push buttons + 3D case | Up to 36 | ~R$55-70 |

Firmware: [`firmware/rp2040-zero/`](firmware/rp2040-zero/)
Assembly guide: [`BUILD.md`](BUILD.md)
Bill of materials: [`BOM.md`](BOM.md) — links and prices across AliExpress, Mercado Livre, and Shopee

## Configuration

All settings live in `radkeys.config.toml` (TOML, plaintext, shareable).
The file is heavily commented so a human or LLM can understand and edit
everything:
- Radiologist name, language (7 options), color theme (13 presets)
- Device VID/PID and protocol 
- Keypad layout (columns × rows)
- Screens and buttons (phrases organized in a hierarchy)

Use the `radkeys-config` binary (included in each release) to edit the config
visually — never touch TOML syntax. The editor shows the 6×6 grid, lets you
add/remove layers and buttons, and validates everything before saving.

You can also edit the TOML file directly with any text editor.

Note: saving the config strips comments (BurntSushi/toml limitation); a backup
(`radkeys.config.toml.bak`) is created before every save so your comments
survive.

## Contributing

See [`AGENTS.md`](AGENTS.md) for AI agent rules, the dev cycle (test → tag →
CI auto-release), and project conventions.

## License

RadKeys is distributed under the **Open Community License v1.1** ([OCL v1.1](LICENSE)).

The hardware design, firmware, and host software are publicly available for
personal use, community modification, and right-to-repair. Any commercial use,
including internal business use, manufacturing, resale, or SaaS offering,
requires a separate written license — see [COMMERCIAL-LICENSE.md](COMMERCIAL-LICENSE.md).

This is a source-available / community license, not an OSI-approved open-source
license or OSHWA-certified open-source hardware license.