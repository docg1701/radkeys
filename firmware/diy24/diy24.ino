/*
 * RadKeys DIY 24 — firmware do dispositivo USB HID custom.
 *
 * Plataforma alvo: Raspberry Pi Pico (RP2040) — 26 GPIO, HID custom maduro.
 * Biblioteca: Adafruit_TinyUSB (instale via Arduino Library Manager).
 * Placa: "Raspberry Pi Pico" com core "Adafruit TinyUSB Library" / "earlephilhower".
 *
 * Protocolo (deve casar com internal/hid/reader_cgo.go — diyReader):
 *   - HID vendor-defined, usage page 0xFF00, usage 0x0001.
 *   - Input report ID 1, payload 24 bytes, 1 byte por botao
 *     (0x00 solto, 0x01 pressionado).
 *   - O hidapi no host le [0x01, b0..b23] (25 bytes).
 *
 * VID/PID: 0x1234 / 0xABCD (placeholders). Edite e registre os seus; o
 * radkeys.config.toml precisa casar com estes vendor_id/product_id.
 *
 * Pinagem: 24 chaves de GPIO0..GPIO23 para GND (INPUT_PULLUP). Ajuste PINS
 * conforme a sua fiação. O botao 0 e o "copy" fixo do RadKeys (indice 0).
 */

#include <Adafruit_TinyUSB.h>

#define DIY_VID 0x1234
#define DIY_PID 0xABCD
#define N_BUTTONS 24

// Descritor HID vendor-defined: report ID 1, 24 bytes de estados (0/1).
uint8_t const desc_hid[] = {
  0x06, 0x00, 0xFF, // Usage Page (Vendor Defined 0xFF00)
  0x09, 0x01,       // Usage (Vendor Defined 1)
  0xA1, 0x01,       // Collection (Application)
  0x85, 0x01,       //   Report ID (1)
  0x15, 0x00,       //   Logical Minimum (0)
  0x25, 0x01,       //   Logical Maximum (1)
  0x75, 0x08,       //   Report Size (8 bits)
  0x95, 0x18,       //   Report Count (24)
  0x09, 0x01,       //   Usage (Vendor Defined 1)
  0x81, 0x02,       //   Input (Data,Var,Abs)
  0xC0              // End Collection
};

Adafruit_USBD_HID usb_hid;

// GPIO dos 24 botoes. Ajuste conforme a fiação.
const uint8_t PINS[N_BUTTONS] = {
  0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
  12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23
};

uint8_t report[N_BUTTONS] = {0};

void setup() {
  TinyUSBDevice.setID(DIY_VID, DIY_PID);
  usb_hid.setReportDescriptor(desc_hid, sizeof(desc_hid));
  usb_hid.begin();

  for (uint8_t i = 0; i < N_BUTTONS; i++) {
    pinMode(PINS[i], INPUT_PULLUP);
  }
}

void loop() {
  if (!TinyUSBDevice.mounted()) {
    return;
  }

  uint8_t next[N_BUTTONS];
  bool changed = false;
  for (uint8_t i = 0; i < N_BUTTONS; i++) {
    next[i] = (digitalRead(PINS[i]) == LOW) ? 0x01 : 0x00; // pull-up: LOW = pressionado
    if (next[i] != report[i]) changed = true;
  }

  if (changed) {
    memcpy(report, next, sizeof(report));
    // sendReport(report_id, payload, len): o host recebe [report_id, payload...].
    usb_hid.sendReport(1, report, sizeof(report));
  }
  delay(5); // debounce leve
}