# RadKeys — Guia de Montagem do Hardware

> Teclado físico 6×6 (36 teclas) com Raw HID custom, fiação ponto a ponto e conectores Dupont.
> Custo total: ~R$ 55–70 (AliExpress).

---

## 1. Lista de Compras (AliExpress)

### 1.1 Eletrônica

| Item | Qtd | Buscar por |
|------|-----|------------|
| **RP2040-Zero** | 1 | "RP2040-Zero" — ~R$ 10 |
| Push buttons SPST 4 pinos 6×6mm | 36 | "push button 6x6 4 pin 100pcs" |
| Diodos 1N4148 through-hole | 36 | "1N4148 diodo through hole" |
| Cabos flexíveis coloridos AWG 24-26 | kit | "wire kit awg 24" ou use cabo de rede |
| Terminais Dupont fêmea + alicate de crimp | 12+ | "dupont connector kit" |
| Ferro de solda, estanho, alicate de corte | 1 cada | — |
| Abraçadeiras pequenas | ~10 | "zip tie small" |

### 1.2 Estrutura

| Item | Qtd | Nota |
|------|-----|------|
| Case 3D impresso — grade 6×6 | 1 | Furos quadrados 6,2×6,2mm, espaçamento 14mm centro a centro |
| Tampa inferior 3D impressa | 1 | Protege a fiação |
| Parafusos M2 ou M3 | 4-6 | Fechar o case |
| Cabo USB-C | 1 | O RP2040-Zero já tem conector USB-C |

---

## 2. Circuito Elétrico (Matriz 6×6 com Anti-Ghosting)

- **Linhas (Rows)**: 6 saídas digitais
- **Colunas (Columns)**: 6 entradas com pull-up interno

Cada push button é ligado em série com um diodo 1N4148:
- **Ânodo** do diodo → terminal do botão (lado da coluna)
- **Cátodo** (faixa preta) → barramento da linha (fio comum daquela linha)
- O **outro terminal** do botão → barramento da coluna

Esse arranjo garante corrente apenas da linha para a coluna, eliminando ghosting.

### Atribuição de Pinos (RP2040-Zero)

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

| Linha | GPIO | | Coluna | GPIO |
|-------|------|-|--------|------|
| R0 | GP0 | | C0 | GP6 |
| R1 | GP1 | | C1 | GP7 |
| R2 | GP2 | | C2 | GP8 |
| R3 | GP3 | | C3 | GP9 |
| R4 | GP4 | | C4 | GP10 |
| R5 | GP5 | | C5 | GP11 |

> GPIOs GP12–GP22 ficam livres para LED, buzzer ou expansão futura.

---

## 3. Montagem Ponto a Ponto + Dupont

### 3.1 Preparação

1. Encaixe os 36 push buttons no frame 3D.
2. Padronize a orientação: terminais mais longos na horizontal ou vertical — mas **igual para todos**.
3. Identifique para cada botão: terminal da coluna e terminal da linha.
   - Sugestão: terminal superior-esquerdo = coluna, inferior-direito = linha.

### 3.2 Soldar os diodos (anti-ghosting)

Para cada botão:
- Solde o **cátodo** (faixa preta) do diodo no terminal escolhido para **linha**.
- Deixe o **ânodo** livre, apontando para fora do botão.

### 3.3 Barramentos de linha

Para cada uma das 6 linhas:
1. Pegue um fio flexível que percorra todos os 6 botões da linha + ~20cm extra.
2. Descasque pequenos trechos onde o fio encontra cada diodo.
3. Enrole o fio descascado no ânodo de cada diodo e solde.
4. Na ponta final, solde ou crimpe um terminal **Dupont fêmea**.
5. Conecte ao GPIO correspondente (R0→GP0, R1→GP1, ... R5→GP5).

### 3.4 Barramentos de coluna

Para cada uma das 6 colunas:
1. Pegue um fio de **outra cor** (ex: preto/azul).
2. Percorra os terminais de **coluna** dos 6 botões daquela coluna.
3. Descasque, enrole no terminal, solde.
4. Na ponta, terminal **Dupont fêmea**.
5. Conecte ao GPIO (C0→GP6, C1→GP7, ... C5→GP11).

### 3.5 Organização final

1. Agrupe os 12 fios (6 linhas + 6 colunas) com abraçadeiras.
2. Plugue cada Dupont fêmea no pino correto do RP2040-Zero.
3. Fixe o RP2040-Zero no fundo do case com fita dupla face.
4. Feche a tampa inferior.

### 3.6 Teste

1. Carregue o firmware no RP2040-Zero (ver seção 4).
2. Conecte o USB-C e execute `./radkeys`.
3. Pressione cada botão — a UI deve mostrar `(row, col)` no log/terminal.
4. Se coordenadas estiverem trocadas (ex: botão físico 0,3 aciona 3,0), troque os Duponts no RP2040-Zero.

---

## 4. Firmware

### 4.1 Código (grave no RP2040-Zero via Arduino IDE)

```cpp
/*
 * RadKeys — firmware RP2040-Zero.
 * Protocolo: envia [row, col] (2 bytes) via HID vendor-defined (TinyUSB).
 * Grid configurável pelo app até a matriz física 6×6 — firmware varre os 6 rows/cols.
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

// Descritor HID: report único de 2 bytes (sem report ID).
// TUD_HID_REPORT_DESC_GENERIC_INOUT(2) declara um relatório sem ID;
// sendReport DEVE passar report_id=0 para o TinyUSB não prepend um byte extra.
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
        usb_hid.sendReport(0, report, sizeof(report));
        delay(30); // debounce
      }
      prevState[r][c] = pressed;
    }

    digitalWrite(colPins[c], HIGH);
  }
  delay(5); // ~200 Hz scan
}
```

### 4.2 Configuração no Arduino IDE

1. Instalar o core **Raspberry Pi Pico/RP2040** (earlephilhower):
   `Arquivo → Preferências → URLs adicionais`:
   ```
   https://github.com/earlephilhower/arduino-pico/releases/download/global/package_rp2040_index.json
   ```
2. Instalar a biblioteca **Adafruit TinyUSB Library** (Library Manager).
3. Selecionar:
   - **Placa**: "Waveshare RP2040 Zero"
   - **USB Stack**: "Adafruit TinyUSB"
4. Conectar o RP2040-Zero com **botão BOOT pressionado** → soltar após conectar.
5. Porta: selecionar a porta que aparece (UF2 Board).
6. Compilar e gravar.

### 4.3 VID/PID

Os valores `0x1234`/`0xABCD` são **placeholders de protótipo** — `0x1234` é
um Vendor ID de exemplo muito reusado e pode colidir com outros gadgets USB.
**Antes de qualquer uso clínico/produção**, troque por um par próprio (um PID
sob um VID registrado, ou um PID open-source alocado) e case com
`radkeys.config.toml`:

```toml
[app.device]
vendor_id = 0x1234
product_id = 0xABCD
protocol = "radkeys-diy"
```

### 4.4 Permissão Linux (udev)

Crie `/etc/udev/rules.d/49-radkeys.rules`:

```
KERNEL=="hidraw*", SUBSYSTEM=="hidraw", ATTRS{idVendor}=="1234", ATTRS{idProduct}=="abcd", MODE="0660", GROUP="input"
```

Depois: `sudo adduser $USER input` (fazer logout/login).

### 4.5 Dependência: xdotool (Linux)

O botão "Paste" do RadKeys envia Ctrl+V para a janela focada (o RIS/PACS).
No Linux isso usa o `xdotool`. Instale:

```
sudo apt install xdotool
```

Sem o xdotool, o botão Paste não funciona no Linux. No Windows não é
necessário — usa a API nativa do Windows (keybd_event).

---

## 5. Durabilidade e Manutenção

- Soldas bem feitas (brilhantes e firmes) garantem contato duradouro.
- Os Duponts permitem desconectar o RP2040-Zero facilmente para manutenção.
- Push buttons: 100.000 a 500.000 ciclos. Se algum falhar, dessoldar 2 fios e trocar.
- O frame 3D protege os componentes.

---

## 6. Custo por unidade

| Item | Unitário (lote 1) | Unitário (lote 10+) |
|------|-------------------|---------------------|
| RP2040-Zero | R$ 10 | R$ 10 |
| Push buttons ×36 | R$ 20 | R$ 12 |
| Diodos 1N4148 ×36 | R$ 5 | R$ 3 |
| Fios + Duponts | R$ 12 | R$ 8 |
| Case 3D (filamento) | R$ 5 | R$ 5 |
| Cabo USB-C | R$ 5 | R$ 3 |
| **Total** | **~R$ 57** | **~R$ 41** |
