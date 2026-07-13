/*
 * RadKeys — RP2040-Zero firmware (composite USB: vendor + keyboard).
 *
 * Two Adafruit_USBD_HID interfaces:
 *   Interface A (vendor): 2-byte IN [row,col] + 2-byte OUT [cmd,arg].
 *   Interface B (keyboard): standard 6KRO HID keyboard.
 *
 * The host writes a 2-byte OUT command to the vendor interface to trigger a
 * device-keyboard keystroke: FIRE_PASTE (Ctrl/Cmd+V) and the editing commands
 * SELECT_ALL / SELECT_LINE / LINE_START / LINE_END / BACKSPACE / DELETE. The
 * keyboard interface injects the keystroke into the currently focused window.
 * The keyboard interface does NOT steal focus — HID keyboard reports go to
 * whatever window already has keyboard focus.
 *
 * Arduino IDE setup:
 *   Board: "Waveshare RP2040 Zero" (earlephilhower core)
 *   USB Stack: "Adafruit TinyUSB"
 *   USB VID: 0x1234  USB PID: 0xABCD  (set in Tools menu — matches DIY_VID/PID)
 */

#include <Adafruit_TinyUSB.h>

#define DIY_VID 0x1234
#define DIY_PID 0xABCD

// Vendor OUT command codes (byte 0 of the 2-byte OUT report).
#define CMD_FIRE_PASTE   0x01
#define CMD_GET_VERSION  0x02
#define CMD_SELECT_ALL    0x03
#define CMD_SELECT_LINE   0x04
#define CMD_LINE_START    0x05
#define CMD_LINE_END      0x06
#define CMD_BACKSPACE     0x07
#define CMD_DELETE        0x08

// Modifier selectors for OS-dependent commands (FIRE_PASTE, SELECT_ALL).
// Byte 1 of the OUT report. Unused (0x00) for the other editing commands;
// SELECT_LINE uses a fixed Shift (not OS-dependent).
#define MOD_CTRL 0x01  // Ctrl — Linux/Windows
#define MOD_GUI  0x02  // GUI/Cmd — macOS

// Firmware version reported in response to CMD_GET_VERSION.
#define FW_VERSION_MAJOR 0x01
#define FW_VERSION_MINOR 0x00

// 6×6 matrix — RP2040-Zero GPIOs
const uint8_t colPins[6] = {6, 7, 8, 9, 10, 11};
const uint8_t rowPins[6] = {0, 1, 2, 3, 4, 5};

// --- Interface A: vendor (IN [row,col] + OUT [cmd,arg]) ---
// TUD_HID_REPORT_DESC_GENERIC_INOUT(2) declares a 2-byte INPUT report and a
// 2-byte OUTPUT report on the vendor usage page (0xFF00). No report ID.
// has_out_endpoint=true allocates a dedicated OUT endpoint so the host can
// write the 2-byte command via the OUT endpoint (fast path).
static const uint8_t desc_vendor[] = {
  TUD_HID_REPORT_DESC_GENERIC_INOUT(2)
};

Adafruit_USBD_HID usb_vendor(desc_vendor, sizeof(desc_vendor),
                             HID_ITF_PROTOCOL_NONE, 2, /*has_out_endpoint=*/true);

// --- Interface B: keyboard (standard 6KRO HID keyboard) ---
// TUD_HID_REPORT_DESC_KEYBOARD() declares the standard keyboard report
// (modifier + reserved + 6 keycodes) plus the LED output report. No report ID.
// has_out_endpoint=false: LED output is received via SET_REPORT control
// transfer; we don't handle it (STALL is fine — we don't need LED state).
static const uint8_t desc_keyboard[] = {
  TUD_HID_REPORT_DESC_KEYBOARD()
};

Adafruit_USBD_HID usb_keyboard(desc_keyboard, sizeof(desc_keyboard),
                               HID_ITF_PROTOCOL_KEYBOARD, 2, /*has_out_endpoint=*/false);

bool prevState[6][6] = {false};

// Pending device-keyboard command armed by the vendor OUT callback, drained in
// loop(). pending_cmd = 0 means none. pending_arg carries the modifier selector
// for OS-dependent commands (FIRE_PASTE, SELECT_ALL); 0x00 for the rest.
// volatile because the callback runs in TinyUSB task/IRQ context while loop()
// reads/clears it. The callback sets arg first, then cmd (commit); loop() reads
// cmd+arg then clears cmd. The benign race (a later command supersedes an
// unprocessed earlier one, or a cmd is clobbered on clear) is acceptable for a
// macro pad driven at human speed — same semantics as the prior pending_paste.
volatile uint8_t pending_cmd = 0;
volatile uint8_t pending_arg = 0;

// Pending version request armed by the vendor OUT callback, drained in loop().
// Separate from pending_cmd: GET_VERSION replies on the vendor IN interface
// (not the keyboard), so it keeps its own flag. volatile for the same reason.
volatile uint8_t pending_version = 0;

// Resolve an OS modifier selector (MOD_CTRL/MOD_GUI) to a HID keyboard modifier
// bitmap. Only called with a known selector (the callback ignores unknown args).
uint8_t modifier_bitmap(uint8_t sel) {
  if (sel == MOD_GUI) return KEYBOARD_MODIFIER_LEFTGUI;  // 0x08
  return KEYBOARD_MODIFIER_LEFTCTRL;                     // 0x01
}

// Send one keystroke via the keyboard interface: modifier + keycode down, wait
// >= 2ms poll interval, release all keys, then wait again so consecutive keys
// in a sequence (e.g. SELECT_LINE = Home then Shift+End) are registered as
// distinct. Guarded by mount state. Called from loop() (NOT from the USB
// callback) so the delay()/report sequence never blocks USB processing.
void send_key(uint8_t modifier, uint8_t keycode) {
  if (!TinyUSBDevice.mounted()) return;

  uint8_t keys[6] = {keycode, 0, 0, 0, 0, 0};
  usb_keyboard.keyboardReport(0, modifier, keys);  // key down: modifier + keycode
  delay(10);                                        // >= 2ms poll interval, let host read key-down
  uint8_t release[6] = {0};
  usb_keyboard.keyboardReport(0, 0, release);      // release all keys + modifiers
  delay(10);                                        // gap before the next key in a sequence
}

// Drain a pending device-keyboard command by sending the matching keystroke(s)
// via the keyboard interface. The report goes to the currently focused window
// (HID keyboard does not steal focus). Unknown cmd: no-op.
void fire_keyboard_command(uint8_t cmd, uint8_t arg) {
  switch (cmd) {
    case CMD_FIRE_PASTE:
      send_key(modifier_bitmap(arg), HID_KEY_V);
      break;
    case CMD_SELECT_ALL:
      send_key(modifier_bitmap(arg), HID_KEY_A);
      break;
    case CMD_SELECT_LINE:
      // Home (jump to line start), then Shift+End (select to line end) = whole line.
      send_key(0, HID_KEY_HOME);
      send_key(KEYBOARD_MODIFIER_LEFTSHIFT, HID_KEY_END);
      break;
    case CMD_LINE_START:
      send_key(0, HID_KEY_HOME);
      break;
    case CMD_LINE_END:
      send_key(0, HID_KEY_END);
      break;
    case CMD_BACKSPACE:
      send_key(0, HID_KEY_BACKSPACE);
      break;
    case CMD_DELETE:
      send_key(0, HID_KEY_DELETE);
      break;
    default:
      break;
  }
}

// GET_REPORT callback for the vendor interface. The host reads [row,col] via
// the interrupt IN endpoint (sendReport), not via GET_REPORT control, so we
// STALL (return 0) — no cached state is needed.
uint16_t get_report_callback(uint8_t report_id, hid_report_type_t report_type,
                             uint8_t *buffer, uint16_t reqlen) {
  (void)report_id; (void)report_type; (void)buffer; (void)reqlen;
  return 0;
}

// SET_REPORT callback for the vendor interface. TinyUSB strips the report-ID
// byte, so buffer points to [cmd, arg] with bufsize == 2. Runs in TinyUSB
// task/IRQ context, so it ONLY arms pending flags — it must NOT block (no
// delay) or send HID reports here. loop() drains the flags and fires the
// keystrokes where blocking is safe.
void set_report_callback(uint8_t report_id, hid_report_type_t report_type,
                         uint8_t const *buffer, uint16_t bufsize) {
  (void)report_id;
  if (report_type != HID_REPORT_TYPE_OUTPUT || bufsize < 2) return;

  uint8_t cmd = buffer[0];
  if (cmd == CMD_GET_VERSION) {
    pending_version = 1;
    return;
  }

  // OS-dependent commands honor the modifier arg; unknown arg -> no-op (matches
  // the FIRE_PASTE semantics). The other editing commands ignore the arg byte.
  switch (cmd) {
    case CMD_FIRE_PASTE:
    case CMD_SELECT_ALL:
      if (buffer[1] == MOD_CTRL || buffer[1] == MOD_GUI) {
        pending_arg = buffer[1];
        pending_cmd = cmd;
      }
      break;
    case CMD_SELECT_LINE:
    case CMD_LINE_START:
    case CMD_LINE_END:
    case CMD_BACKSPACE:
    case CMD_DELETE:
      pending_arg = 0x00;
      pending_cmd = cmd;
      break;
    default:
      break;  // unknown cmd: no-op
  }
}

void setup() {
  // Initialize the TinyUSB stack (defensive — earlephilhower auto-inits, but
  // the dual-interface pattern benefits from an explicit begin before the
  // HID interfaces register).
  if (!TinyUSBDevice.isInitialized()) TinyUSBDevice.begin(0);

  for (int c = 0; c < 6; c++) {
    pinMode(colPins[c], OUTPUT);
    digitalWrite(colPins[c], HIGH);
  }
  for (int r = 0; r < 6; r++) {
    pinMode(rowPins[r], INPUT_PULLUP);
  }

  // Vendor interface: register OUT command callback before begin.
  usb_vendor.setReportCallback(get_report_callback, set_report_callback);
  usb_vendor.begin();

  usb_keyboard.begin();

  // Wait for USB enumeration.
  while (!TinyUSBDevice.mounted()) delay(1);
}

void loop() {
  // Only scan when USB is connected (avoids spam if disconnected).
  if (!TinyUSBDevice.mounted()) {
    delay(100);
    return;
  }

  // Drain a pending device-keyboard command armed by the vendor OUT callback.
  // Done here (not in the USB callback) so the keystroke delay()/reports never
  // block USB.
  uint8_t cmd = pending_cmd;
  if (cmd != 0) {
    uint8_t arg = pending_arg;
    pending_cmd = 0;
    fire_keyboard_command(cmd, arg);
  }

  // Drain a pending version request armed by the vendor OUT callback.
  // Sends a one-shot 2-byte IN report [FW_VERSION_MAJOR, FW_VERSION_MINOR].
  // Done in loop() (not in the callback) to keep the callback non-blocking.
  if (pending_version) {
    pending_version = 0;
    uint8_t ver[2] = {FW_VERSION_MAJOR, FW_VERSION_MINOR};
    usb_vendor.sendReport(0, ver, sizeof(ver));
  }

  for (int c = 0; c < 6; c++) {
    digitalWrite(colPins[c], LOW);
    delayMicroseconds(10);

    for (int r = 0; r < 6; r++) {
      bool pressed = (digitalRead(rowPins[r]) == LOW);
      if (pressed && !prevState[r][c]) {
        uint8_t report[2] = {uint8_t(r), uint8_t(c)};
        usb_vendor.sendReport(0, report, sizeof(report));
        delay(30); // debounce
      }
      prevState[r][c] = pressed;
    }

    digitalWrite(colPins[c], HIGH);
  }
  delay(5); // ~200 Hz scan
}
