# RadKeys

Open source shortcut deck for radiology reports.

RadKeys is a portable, single-binary desktop app for Windows, macOS and Linux.
It lets radiologists navigate a visual hierarchy of shortcuts and copy
pre-written report phrases to the clipboard **without stealing focus** from the
RIS/PACS window.

## How it works

A USB HID custom device (Stream Deck / clone, or a DIY 24-key pad) sends
button presses directly to the app via hidapi — **no keyboard keys, no
modifiers, no focus stealing**. The app shows a grid of buttons (3 fixed:
Copy / Up / Home, plus configurable per screen) and a text preview. Press a
button → phrase loads in the preview → press Copy → paste into the RIS.

## Status

Early prototype. Compiles and runs on Linux (amd64). Windows/macOS builds
pending cross-compile CI.

## Quick start (Linux)

```bash
# prerequisites
sudo apt install -y golang-go libgl1-mesa-dev xorg-dev libudev-dev

# build & run (mock mode — no hardware needed, UI works via mouse clicks)
go build -o radkeys .
./radkeys
```

Without a device, the app falls back to an in-process mock; the UI is fully
drivable by mouse clicks.

## Hardware

| Option | Device | Buttons | Cost |
|--------|--------|---------|------|
| **Buy ready** | Stream Deck / Elgato-compatible clone | 15 / 32 | $$$ |
| **DIY (primary)** | Arduino Pro Micro + salvaged keyboard switches + 3D printed case | 24 | ~R$30-50 |
| **DIY (alt)** | Raspberry Pi Pico (RP2040) + switches + 3D case | 24 | ~R$40-60 |

Firmware and build instructions: [`firmware/arduino/`](firmware/arduino/) (primary)
and [`firmware/rp2040/`](firmware/rp2040/) (alternative).

## Configuration

Edit `radkeys.config.toml` (TOML, plaintext). See the bundled example.

## Architecture

See [`brief.md`](brief.md) for the full technical brief (v1.3).

```
internal/config   TOML parser + validation
internal/deck     navigation state machine
internal/hid      HID reader (hidapi / mock)
internal/ui       Fyne UI (grid + preview + clipboard)
```

## License

MIT — see [LICENSE](./LICENSE).