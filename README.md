# RadKeys

Portable shortcut deck for radiology reports — copy pre-written phrases to
the clipboard without stealing focus from the RIS/PACS.

![RadKeys Screenshot](screenshot.png)

[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

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
| `radkeys-linux-amd64` | Linux x86_64 |
| `radkeys-windows-amd64.exe` | Windows x86_64 |
| `radkeys.config.toml` | Config template (all platforms) |

**macOS**: binary not provided (cross-compile from Linux is impossible — needs
Apple's proprietary SDK). macOS is supported in code: build from source on a
Mac following the instructions below. Paste is cross-platform via the device
(no per-OS keystroke injection), so macOS works the same as Linux/Windows.

Put the binary and `radkeys.config.toml` in the same directory and run.

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
   at the cursor position. No host-side software is needed for Paste — the
   device is the keyboard. RadKeys never steals focus.

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
CGO_ENABLED=1 go build -tags flatpak -o radkeys-linux-amd64 .

# Windows (on Windows, or cross-compile from Linux with mingw)
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -o radkeys-windows-amd64.exe .

# macOS Intel (on a Mac — cross-compile from Linux is impossible)
CGO_ENABLED=1 go build -o radkeys-macos-amd64 .

# macOS Apple Silicon (on a Mac)
CGO_ENABLED=1 GOARCH=arm64 go build -o radkeys-macos-arm64 .
```

### Cross-compile from Linux (Windows only)

```bash
sudo apt install -y gcc-mingw-w64
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -o radkeys-windows-amd64.exe .
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

## Configuration

All settings live in `radkeys.config.toml` (TOML, plaintext, shareable).
The file is heavily commented so a human or LLM can understand and edit
everything:
- Radiologist name, language (7 options), color theme (13 presets)
- Device VID/PID and protocol 
- Keypad layout (columns × rows)
- Screens and buttons (phrases organized in a hierarchy)

Edit the file manually — the UI's "Settings" tab only changes app settings,
not screens/buttons. To add phrases, edit the TOML file directly.

Note: the Settings tab's Save rewrites the file without comments
(BurntSushi/toml does not preserve them); it first copies the previous file
to `radkeys.config.toml.bak` so your comments are not lost.

## Contributing

See [`AGENTS.md`](AGENTS.md) for AI agent rules, the dev cycle (test → tag →
CI auto-release), and project conventions.

## License

MIT — see [LICENSE](LICENSE).