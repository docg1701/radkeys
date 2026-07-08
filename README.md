# RadKeys

Portable shortcut deck for radiology reports — copy pre-written phrases to
the clipboard without stealing focus from the RIS/PACS.

![RadKeys Screenshot](screenshot.png)

[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

## What it is

A single-binary desktop app for **Linux and Windows**. (macOS: build from
source — cross-compile is impossible with CGO.) A USB HID custom device
(Stream Deck or DIY 24-key pad) sends button presses directly to the app via
hidapi — no keyboard keys, no focus stealing, no modifier interference.
The app shows a preview on top and a virtual keypad on the bottom; press a
button → phrase loads → copy → paste into the RIS.

Release = **1 executable + 1 config file**. Icon, translations (7 languages),
and color themes (13 presets, including system default) are all embedded in
the binary.

## Download

Get the latest release from [Releases](../../releases). Each release includes:

| File | Platform |
|------|----------|
| `radkeys-linux-amd64` | Linux x86_64 |
| `radkeys-windows-amd64.exe` | Windows x86_64 |
| `radkeys.config.toml` | Config template (all platforms) |

**macOS**: not provided (cross-compile from Linux is impossible — needs Apple's
proprietary SDK). Build from source on a Mac following the instructions below.

Put the binary and `radkeys.config.toml` in the same directory and run.

Without a hardware device, the app runs in mock mode — the UI works via mouse
clicks.

## Usage

1. Edit `radkeys.config.toml` to add your phrases. The file is heavily
   commented — a human or LLM can read it and generate a custom config
   following the rules in the comments.
2. Connect your USB device (Stream Deck / Elgato-compatible clone, or the
   DIY 24-key pad).
3. Run RadKeys.
4. Press a button → phrase appears in the preview → press Copy → paste in
   the RIS (Ctrl+V).

The radiologist never touches the keyboard except to paste.

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
CGO_ENABLED=1 go build -o radkeys-linux-amd64 .

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

## Hardware

| Option | Device | Keys | Cost |
|--------|--------|------|------|
| Buy ready | Stream Deck / Elgato-compatible clone | 15 / 32 | $$$ |
| DIY (primary) | Arduino Pro Micro + salvaged switches + 3D case | 24 | ~R$30-50 |
| DIY (alt) | Raspberry Pi Pico + switches + 3D case | 24 | ~R$40-60 |

Firmware: [`firmware/arduino/`](firmware/arduino/) · [`firmware/rp2040/`](firmware/rp2040/)

## Configuration

All settings live in `radkeys.config.toml` (TOML, plaintext, shareable).
The file is heavily commented so a human or LLM can understand and edit
everything:
- Radiologist name, language (7 options), color theme (13 presets)
- Device VID/PID and protocol (Elgato or DIY)
- Keypad layout (columns × rows)
- Screens and buttons (phrases organized in a hierarchy)

Edit the file manually — the UI's "Ajustes" tab only changes app settings,
not screens/buttons. To add phrases, edit the TOML file directly.

## Contributing

See [`AGENTS.md`](AGENTS.md) for AI agent rules, the dev cycle (test → tag →
CI auto-release), and project conventions.

## License

MIT — see [LICENSE](LICENSE).