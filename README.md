# RadKeys

Portable shortcut deck for radiology reports — copy pre-written phrases to the clipboard without stealing focus from the RIS/PACS.

[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

## What it is

A single-binary desktop app for Windows, macOS and Linux. A USB HID custom device (Stream Deck or DIY 24-key pad) sends button presses directly to the app via hidapi — no keyboard keys, no focus stealing, no modifier interference. The app shows a preview on top and a virtual keypad on the bottom; press a button → phrase loads → copy → paste into the RIS.

## Install

Download the executable and `radkeys.config.toml` from the [latest release](../../releases). Put both in the same directory. Run the executable.

```bash
./radkeys        # Linux
radkeys.exe      # Windows
open radkeys     # macOS
```

Without a hardware device, the app runs in mock mode — the UI works via mouse clicks.

## Usage

1. Edit `radkeys.config.toml` to add your phrases (or ask an LLM to read the example file and generate your custom config following the rules in the comments).
2. Connect your USB device (Stream Deck or DIY 24-key pad).
3. Run RadKeys.
4. Press a button → phrase appears in the preview → press Copy → paste in the RIS.

## Hardware

| Option | Device | Keys | Cost |
|--------|--------|------|------|
| Buy ready | Stream Deck / Elgato-compatible clone | 15 / 32 | $$$ |
| DIY (primary) | Arduino Pro Micro + salvaged switches + 3D case | 24 | ~R$30-50 |
| DIY (alt) | Raspberry Pi Pico + switches + 3D case | 24 | ~R$40-60 |

Firmware: [`firmware/arduino/`](firmware/arduino/) · [`firmware/rp2040/`](firmware/rp2040/)

## Configuration

All settings are in `radkeys.config.toml` (TOML, plaintext, shareable). The file is heavily commented so a human or LLM can understand and edit everything. See the bundled example.

## Contributing

See [`AGENTS.md`](AGENTS.md) for AI agent rules and project conventions.

## License

MIT — see [LICENSE](LICENSE).