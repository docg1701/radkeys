# RadKeys — Hardware Assembly Guide

> Physical 6×6 (36-key) keypad using 12×12mm push buttons with removable square caps, point-to-point wiring, and Dupont connectors.

**Shopping list**: see [`BOM.md`](BOM.md) for component links, prices, and quantity recommendations across AliExpress, Mercado Livre, and Shopee.

---

## 1. Components

---

## 2. Electrical Circuit (6×6 Matrix with Anti-Ghosting)

- **Rows**: 6 digital outputs
- **Columns**: 6 inputs with internal pull-up

Each push button is wired in series with a 1N4148 diode:
- **Diode anode** → button terminal (column side)
- **Cathode** (black band) → row bus (common wire for that row)
- The **other terminal** of the button → column bus

This arrangement ensures current only flows from row to column, eliminating ghosting.

### Pin Assignment (RP2040-Zero)

```
        Col 0  Col 1  Col 2  Col 3  Col 4  Col 5
        (GP6)  (GP7)  (GP8)  (GP9)  (GP10) (GP11)
Row 0 (GP0)  B0     B1     B2     B3     B4     B5
Row 1 (GP1)  B6     B7     B8     B9     B10    B11
Row 2 (GP2)  B12    B13    B14    B15    B16    B17
Row 3 (GP3)  B18    B19    B20    B21    B22    B23
Row 4 (GP4)  B24    B25    B26    B27    B28    B29
Row 5 (GP5)  B30    B31    B32    B33    B34    B35
```

| Row | GPIO | | Column | GPIO |
|-------|------|-|--------|------|
| R0 | GP0 | | C0 | GP6 |
| R1 | GP1 | | C1 | GP7 |
| R2 | GP2 | | C2 | GP8 |
| R3 | GP3 | | C3 | GP9 |
| R4 | GP4 | | C4 | GP10 |
| R5 | GP5 | | C5 | GP11 |

> GPIOs GP12–GP22 are free for LED, buzzer, or future expansion.

---

## 3. Point-to-Point + Dupont Assembly

### 3.1 Preparation

1. Fit the 36 push buttons (12×12mm) into the 3D frame.
2. Standardize the orientation: longer terminals horizontally or vertically — but **the same for all**.
3. Identify for each button: the column terminal and the row terminal.
   - Suggestion: upper-left terminal = column, lower-right = row.
4. Snap the square cap onto each button stem *after* soldering (caps are removable).

### 3.2 Solder the Diodes (Anti-Ghosting)

For each button:
- Solder the **cathode** (black band) of the diode to the terminal chosen for **row**.
- Leave the **anode** free, pointing outward from the button.

### 3.3 Row Buses

For each of the 6 rows:
1. Take a flexible wire that runs across all 6 buttons in the row + ~20cm extra.
2. Strip small sections where the wire meets each diode.
3. Wrap the stripped wire around the anode of each diode and solder.
4. At the final end, solder or crimp a **female Dupont** terminal.
5. Connect to the corresponding GPIO (R0→GP0, R1→GP1, ... R5→GP5).

### 3.4 Column Buses

For each of the 6 columns:
1. Take a wire of **another color** (e.g., black/blue).
2. Run across the **column** terminals of the 6 buttons in that column.
3. Strip, wrap around the terminal, and solder.
4. At the end, attach a **female Dupont** terminal.
5. Connect to the GPIO (C0→GP6, C1→GP7, ... C5→GP11).

### 3.5 Cable Management

1. Group the 12 wires (6 rows + 6 columns) with zip ties.
2. Plug each female Dupont into the correct RP2040-Zero pin.
3. Fix the RP2040-Zero at the bottom of the case with double-sided tape.
4. Close the bottom cover.

### 3.6 Testing

1. Flash the firmware onto the RP2040-Zero (see section 4).
2. Connect the USB-C cable and run `./radkeys`.
3. Press each button — the UI should show `(row, col)` in the log/terminal.
4. If coordinates are swapped (e.g., physical button 0,3 triggers 3,0), swap the Duponts on the RP2040-Zero.

---

## 4. Firmware

The device is a **composite USB**: vendor HID interface (`[row,col]` events) +
HID keyboard interface (Ctrl/Cmd+V for paste). **Single factory flash** — the
firmware is never rewritten for configuration (all configuration lives in the
App/TOML). The Paste button makes the device send Ctrl/Cmd+V to the focused
window (the RIS) as a USB keyboard.

### 4.1 Code (Flash onto the RP2040-Zero via Arduino IDE)

The firmware is at [`firmware/rp2040-zero/diy.ino`](firmware/rp2040-zero/diy.ino)
— composite USB device (vendor HID `[row,col]` + HID keyboard interface for
paste). See [`PROTOCOL.md`](firmware/rp2040-zero/PROTOCOL.md) for the command
protocol.

### 4.2 Arduino IDE Configuration

1. Install the **Raspberry Pi Pico/RP2040** core (earlephilhower):
   `File → Preferences → Additional Boards Manager URLs`:
   ```
   https://github.com/earlephilhower/arduino-pico/releases/download/global/package_rp2040_index.json
   ```
2. Install the **Adafruit TinyUSB Library** (Library Manager).
3. Select:
   - **Board**: "Waveshare RP2040 Zero"
   - **USB Stack**: "Adafruit TinyUSB"
4. Connect the RP2040-Zero with the **BOOT button pressed** → release after connecting.
5. Port: select the port that appears (UF2 Board).
6. Compile and flash.

### 4.3 VID/PID

The values `0x1234`/`0xABCD` are **prototype placeholders** — `0x1234` is a
commonly reused example Vendor ID and may collide with other USB gadgets.
**Before any clinical/production use**, replace them with your own pair (a PID
under a registered VID, or an allocated open-source PID) and match
`radkeys.config.toml`:

```toml
[app.device]
vendor_id = 0x1234
product_id = 0xABCD
protocol = "radkeys-diy"
```

### 4.4 Linux Permission (udev)

Create `/etc/udev/rules.d/49-radkeys.rules`:

```
KERNEL=="hidraw*", SUBSYSTEM=="hidraw", ATTRS{idVendor}=="1234", ATTRS{idProduct}=="abcd", MODE="0660", GROUP="input"
```

Then: `sudo adduser $USER input` (log out / log in).

---

## 5. Durability and Maintenance

- Well-made solders (shiny and firm) ensure long-lasting contact.
- The Dupont connectors let you disconnect the RP2040-Zero easily for maintenance.
- Push buttons: 100,000 to 500,000 cycles. If one fails, desolder 2 wires and replace it.
- The 3D frame protects the components.

---

## 6. Cost

See [`BOM.md`](BOM.md) for current prices across AliExpress, Mercado Livre, and Shopee, with per-platform summaries.
