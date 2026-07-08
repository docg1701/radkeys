/*
 * RadKeys DIY 24 — firmware para Arduino Pro Micro (ATmega32U4).
 *
 * Hardware: Arduino Pro Micro + 24 chaves em matriz 6 colunas × 4 linhas
 * (10 pinos GPIO) + caixa 3D printed + cabo USB.
 *
 * As chaves podem ser reaproveitadas de um teclado barato chinês (ex.:
 * switches mecânicos de teclado quebrado), como no projeto de referência
 * Mercawa/DIYStreamDeck-HIDKeyboard.
 *
 * Protocolo (casa com internal/hid/reader_cgo.go — diyReader):
 *   - HID vendor-defined, usage page 0xFF00, usage 0x0001.
 *   - Input report ID 1, payload 24 bytes, 1 byte por botão
 *     (0x00 solto, 0x01 pressionado).
 *   - O hidapi no host lê [0x01, b0..b23] (25 bytes).
 *
 * VID/PID: 0x1234 / 0xABCD (placeholders). Edite e case com
 * radkeys.config.toml (vendor_id / product_id / protocol = "radkeys-diy").
 *
 * Placa: "Arduino Leonardo" ou "Arduino Micro" no Arduino IDE.
 * NÃO requer bibliotecas externas — usa apenas o core HID do 32U4.
 */

#include <HID.h>

#define DIY_VID 0x1234
#define DIY_PID 0xABCD
#define COLS 6
#define ROWS 4
#define N_BUTTONS (COLS * ROWS) // 24

// Descritor HID vendor-defined: report ID 1, 24 bytes de estados (0/1).
static const uint8_t PROGMEM desc_hid[] = {
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

// Matriz 6×4: colunas (output) e linhas (input com pull-up).
// Ajuste os pinos conforme a sua fiação.
const uint8_t colPins[COLS] = {2, 3, 4, 5, 6, 7};
const uint8_t rowPins[ROWS] = {8, 9, 10, 14}; // 14 = A0

uint8_t report[N_BUTTONS] = {0};

void setup() {
  // Colunas: output, HIGH (não ativas).
  for (uint8_t c = 0; c < COLS; c++) {
    pinMode(colPins[c], OUTPUT);
    digitalWrite(colPins[c], HIGH);
  }
  // Linhas: input com pull-up (HIGH = solto, LOW = pressionado).
  for (uint8_t r = 0; r < ROWS; r++) {
    pinMode(rowPins[r], INPUT_PULLUP);
  }

  // HID vendor-defined: anexa o descritor custom ao dispositivo HID.
  HID().AppendDescriptor(new HIDSubDescriptor(desc_hid, sizeof(desc_hid)));

  // Inicializa o USB HID. O PC reconhece como dispositivo composto
  // (teclado + vendor-defined). O RadKeys lê o report vendor-defined.
  // O boot keyboard padrão do 32U4 não interfere (não enviamos teclas).
}

void loop() {
  uint8_t next[N_BUTTONS];
  bool changed = false;

  // Varredura da matriz: ativa uma coluna por vez, lê as 4 linhas.
  for (uint8_t c = 0; c < COLS; c++) {
    digitalWrite(colPins[c], LOW);
    delayMicroseconds(10); // estabiliza
    for (uint8_t r = 0; r < ROWS; r++) {
      uint8_t idx = r * COLS + c;
      next[idx] = (digitalRead(rowPins[r]) == LOW) ? 0x01 : 0x00;
      if (next[idx] != report[idx]) changed = true;
    }
    digitalWrite(colPins[c], HIGH);
  }

  if (changed) {
    memcpy(report, next, sizeof(report));
    // Envia o report vendor-defined (report ID 1 + 24 bytes).
    HID().SendReport(1, report, sizeof(report));
  }
  delay(5); // ~200 Hz de scan, debounce leve
}