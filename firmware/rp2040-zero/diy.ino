/*
 * RadKeys — RP2040-Zero firmware (composite USB: vendor + keyboard).
 *
 * Two Adafruit_USBD_HID interfaces:
 *   Interface A (vendor): 2-byte IN [row,col] + 2-byte OUT [cmd,arg].
 *   Interface B (keyboard): standard 6KRO HID keyboard for Ctrl/Cmd+V paste.
 *
 * The host writes a 2-byte OUT command to the vendor interface to trigger
 * FIRE_PASTE, which injects a Ctrl+V (or Cmd+V) keystroke via the keyboard
 * interface into the currently focused window. The keyboard interface does
 * NOT steal focus — HID keyboard reports go to whatever window already has
 * keyboard focus.
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
#define CMD_FIRE_PASTE 0x01

// Modifier selectors for FIRE_PASTE (byte 1 of the OUT report).
#define MOD_CTRL 0x01  // Ctrl — Linux/Windows
#define MOD_GUI  0x02  // GUI/Cmd — macOS

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

// Pending paste armed by the vendor OUT callback, drained in loop().
// 0 = none, 1 = Ctrl (Linux/Windows), 2 = GUI/Cmd (macOS). volatile because the
// callback runs in TinyUSB task/IRQ context while loop() reads/clears it.
volatile uint8_t pending_paste = 0;

// Inject Ctrl/Cmd+V via the keyboard interface, then release all keys.
// The report goes to the currently focused window (HID keyboard does not
// steal focus). Guarded by mount state. Called from loop() (NOT from the USB
// callback) so the delay()/report sequence never blocks USB processing.
void fire_paste(uint8_t modifier) {
  if (!TinyUSBDevice.mounted()) return;

  uint8_t keys[6] = {HID_KEY_V, 0, 0, 0, 0, 0};
  usb_keyboard.keyboardReport(0, modifier, keys);  // key down: modifier + V
  delay(10);                                        // >= 2ms poll interval, let host read key-down
  uint8_t release[6] = {0};
  usb_keyboard.keyboardReport(0, 0, release);      // release all keys + modifiers
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
// task/IRQ context, so it ONLY arms pending_paste — it must NOT block (no
// delay) or send HID reports here. loop() drains the flag and fires the
// keystroke where blocking is safe.
void set_report_callback(uint8_t report_id, hid_report_type_t report_type,
                         uint8_t const *buffer, uint16_t bufsize) {
  (void)report_id;
  if (report_type != HID_REPORT_TYPE_OUTPUT || bufsize < 2) return;
  if (buffer[0] != CMD_FIRE_PASTE) return;  // only FIRE_PASTE; others reserved/no-op

  if (buffer[1] == MOD_CTRL) {
    pending_paste = 1;
  } else if (buffer[1] == MOD_GUI) {
    pending_paste = 2;
  }
  // Unknown arg: leave pending_paste unchanged (no-op).
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

  // Drain a pending paste armed by the vendor OUT callback. Done here (not
  // in the USB callback) so the keystroke delay()/reports never block USB.
  uint8_t job = pending_paste;
  if (job != 0) {
    pending_paste = 0;
    uint8_t mod = (job == 1) ? KEYBOARD_MODIFIER_LEFTCTRL   // 0x01
                             : KEYBOARD_MODIFIER_LEFTGUI;   // 0x08
    fire_paste(mod);
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
