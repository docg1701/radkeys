/*
 * RadKeys — firmware RP2040-Zero.
 * Protocolo: envia [row, col] (2 bytes) via HID vendor-defined (TinyUSB).
 * Grid configurável pelo app — firmware NÃO hardcoded tamanho.
 *
 * Configuração Arduino IDE:
 *   Placa: "Waveshare RP2040 Zero" (core earlephilhower)
 *   USB Stack: "Adafruit TinyUSB"
 */

#include <Adafruit_TinyUSB.h>

#define DIY_VID 0x1234
#define DIY_PID 0xABCD

// Matriz 6×6 — GPIOs do RP2040-Zero
const uint8_t colPins[6] = {6, 7, 8, 9, 10, 11};
const uint8_t rowPins[6] = {0, 1, 2, 3, 4, 5};

// Descritor HID: report ID 1, 2 bytes (row, col)
static const uint8_t desc_hid[] = {
  TUD_HID_REPORT_DESC_GENERIC_INOUT(2)
};

Adafruit_USBD_HID usb_hid(desc_hid, sizeof(desc_hid), HID_ITF_PROTOCOL_VENDOR, 2, false);

bool prevState[6][6] = {false};

void setup() {
  for (int c = 0; c < 6; c++) {
    pinMode(colPins[c], OUTPUT);
    digitalWrite(colPins[c], HIGH);
  }
  for (int r = 0; r < 6; r++) {
    pinMode(rowPins[r], INPUT_PULLUP);
  }

  usb_hid.begin();

  // Espera USB enumerar
  while (!TinyUSBDevice.mounted()) delay(1);
}

void loop() {
  // Só varre quando USB está conectado (evita spam se desconectado)
  if (!TinyUSBDevice.mounted()) {
    delay(100);
    return;
  }

  for (int c = 0; c < 6; c++) {
    digitalWrite(colPins[c], LOW);
    delayMicroseconds(10);

    for (int r = 0; r < 6; r++) {
      bool pressed = (digitalRead(rowPins[r]) == LOW);
      if (pressed && !prevState[r][c]) {
        uint8_t report[2] = {uint8_t(r), uint8_t(c)};
        usb_hid.sendReport(1, report, sizeof(report));
        delay(30); // debounce
      }
      prevState[r][c] = pressed;
    }

    digitalWrite(colPins[c], HIGH);
  }
  delay(5); // ~200 Hz scan
}
