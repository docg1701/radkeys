# RadKeys — RP2040-Zero USB Protocol

> Composite HID device (vendor + keyboard). Matches `firmware/rp2040-zero/diy.ino` exactly.

## Device

| Field | Value |
|-------|-------|
| VID | `0x1234` |
| PID | `0xABCD` |
| USB class | Composite (two independent HID interfaces) |

The device exposes two HID interfaces via TinyUSB's composite descriptor.
The PID must remain stable once the keyboard interface is added — OSes cache
the device driver binding after the first enumeration.

## Interface A — Vendor (usage page `0xFF00`)

| Property | Value |
|----------|-------|
| Report descriptor | `TUD_HID_REPORT_DESC_GENERIC_INOUT(2)` |
| Interface protocol | `HID_ITF_PROTOCOL_NONE` (0) |
| Poll interval | 2 ms |
| OUT endpoint | Yes (`has_out_endpoint=true`) |
| Report ID | None (`report_id=0`) |

`GENERIC_INOUT(2)` declares a 2-byte INPUT report and a 2-byte OUTPUT report
on the vendor usage page. The OUT endpoint lets the host write commands via
the interrupt OUT endpoint (fast path) instead of a `SET_REPORT` control
transfer.

### IN report — `[row, col]`

Sent on key press via the interrupt IN endpoint. 2 bytes, no report ID.

| Byte | Field | Range |
|------|-------|-------|
| 0 | row | 0–5 |
| 1 | col | 0–5 |

- Edge-triggered on press only (no release report).
- Debounce: 30 ms per press.
- Scan rate: ~200 Hz.
- Sent via `usb_vendor.sendReport(0, report, 2)`.

### OUT report — `[cmd, arg]`

Written by the host to the OUT endpoint. 2 bytes, no report ID.

| Byte | Field |
|------|-------|
| 0 | cmd |
| 1 | arg |

#### Command table (cmd byte)

| Value | Name | Description |
|-------|------|-------------|
| `0x00` | — | Reserved (no-op) |
| `0x01` | `FIRE_PASTE` | Inject Ctrl/Cmd+V keystroke via the keyboard interface |
| Other | — | Reserved (no-op) |

#### Modifier table (arg byte, for `FIRE_PASTE` only)

| Value | Modifier | USB HID constant | Firmware mapping |
|-------|----------|------------------|-----------------|
| `0x01` | Ctrl (Linux/Windows) | `KEYBOARD_MODIFIER_LEFTCTRL` | `0x01` |
| `0x02` | GUI/Cmd (macOS) | `KEYBOARD_MODIFIER_LEFTGUI` | `0x08` |
| Other | — | — | Ignored (no keystroke sent) |

When `cmd = FIRE_PASTE` with an unknown arg, the entire command is ignored
(no keystroke is sent).

## Interface B — Keyboard (usage page `0x0001`)

| Property | Value |
|----------|-------|
| Report descriptor | `TUD_HID_REPORT_DESC_KEYBOARD()` |
| Interface protocol | `HID_ITF_PROTOCOL_KEYBOARD` (1) |
| Poll interval | 2 ms |
| OUT endpoint | No (`has_out_endpoint=false`) |
| Report ID | None (`report_id=0`) |

The keyboard descriptor still includes a 1-byte LED Output report (Num/Caps/Scroll Lock); because `has_out_endpoint=false`, it is delivered via `SET_REPORT` control transfer, not a dedicated interrupt OUT endpoint. The firmware does not act on LED state.

Standard 6KRO HID keyboard report:

| Byte | Field |
|------|-------|
| 0 | modifier bitmap |
| 1 | reserved (always 0) |
| 2–7 | keycodes (up to 6 simultaneous) |

### FIRE_PASTE keystroke sequence

The vendor `set_report` callback runs in TinyUSB task/IRQ context, so it does
NOT execute the keystroke directly — it only arms a `volatile pending_paste`
flag (0 = none, 1 = Ctrl, 2 = GUI). The main `loop()` drains the flag and runs
the keystroke sequence where blocking is safe:

1. **Guard:** check `TinyUSBDevice.mounted()`.
2. **Key down:** `usb_keyboard.keyboardReport(0, modifier, {HID_KEY_V, 0, 0, 0, 0, 0})`.
3. **Wait:** `delay(10)` — ≥ 2 ms poll interval, lets the host read the key-down report.
4. **Release all:** `usb_keyboard.keyboardReport(0, 0, {0, 0, 0, 0, 0, 0})`.

`HID_KEY_V = 0x19` (USB HID keycode for 'V'). The flag is a single `volatile
uint8_t`; the benign race between callback-arm and loop-drain is acceptable for
paste (a later command simply supersedes an unprocessed earlier one).

### Focus behavior

The keyboard interface does **NOT** steal focus. A HID keyboard report is
injected into the OS input stream and dispatched to the window that already
has keyboard focus. No window activation, raise, or focus-switch occurs.
This is how all HID macro pads operate — paste goes to the currently focused
target.

## Host write layout (go-hid / HIDAPI)

The host writes the 2-byte OUT command via `Device.Write` (output report).
The first byte is the report ID:

```
[0x00, cmd, arg]
```

| Byte | Content |
|------|---------|
| 0 | Report ID (`0x00` — no report ID in descriptor; consumed by HIDAPI/TinyUSB, **not** passed to the device callback) |
| 1 | `cmd` |
| 2 | `arg` |

The device's `set_report` callback receives `buffer` pointing at byte 1
(`cmd`) with `bufsize = 2`. The report-ID byte is stripped by HIDAPI/TinyUSB.

### Examples

Fire paste with Ctrl (Linux/Windows):

```go
out := []byte{0x00, 0x01, 0x01}
dev.Write(out)
```

Fire paste with Cmd (macOS):

```go
out := []byte{0x00, 0x01, 0x02}
dev.Write(out)
```

> **Do NOT use `SendFeatureReport`.** The descriptor declares an `Output`
> report item, not a `Feature` report. `Write` (output report) is the
> correct call — it uses the OUT endpoint when available, or falls back to
> a `SET_REPORT` control transfer.

## Matrix scan (reference)

| Property | Value |
|----------|-------|
| Rows | GPIO 0–5 (`INPUT_PULLUP`) |
| Columns | GPIO 6–11 (`OUTPUT`) |
| Scan method | Column-scan, row-read |
| Trigger | Edge on press only |
| Debounce | 30 ms |
| Scan rate | ~200 Hz (5 ms loop delay) |
| State tracking | `prevState[6][6]` |

IN report `[row, col]` is sent on press. No report is sent on release.

## Constraints

- **No report IDs.** Neither interface uses a report ID. All `sendReport`
  and `keyboardReport` calls pass `report_id = 0`.
- **Single factory flash.** The device is flashed once and never reflashed.
  All configuration lives in the host app (TOML). The device receives only
  transient commands in RAM (e.g., fire paste) and never persists anything.
- **PID stability.** The PID (`0xABCD`) must not change once the composite
  device is in use — OSes cache the driver binding after first enumeration.