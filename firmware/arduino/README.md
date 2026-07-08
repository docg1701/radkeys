# RadKeys DIY 24 — firmware Arduino Pro Micro (ATmega32U4)

Dispositivo USB de 24 botões que o RadKeys lê diretamente via hidapi
(`protocol = "radkeys-diy"`). **Não envia teclas** — é um HID vendor-defined,
logo não rouba foco do RIS e não interfere em atalhos do app focado.

## Plataforma

**Arduino Pro Micro (ATmega32U4)** — barato (~R$15-25 no AliExpress), amplamente
disponível no Brasil, USB nativo HID. É o mesmo microcontrolador usado no
projeto de referência [Mercawa/DIYStreamDeck-HIDKeyboard](https://github.com/Mercawa/DIYStreamDeck-HIDKeyboard).

> Alternativa: [Raspberry Pi Pico (RP2040)](../rp2040/README.md) — 24 GPIO
> diretos (sem matriz), HID custom via Adafruit_TinyUSB. Um pouco mais caro
> (~R$20-30) mas mais simples na fiação.

## Hardware (BOM)

| Item | Qtd | Nota |
|------|-----|------|
| Arduino Pro Micro (ATmega32U4) | 1 | Clone barato funciona. |
| Chaves mecânicas (switches) | 24 | Reaproveite de um teclado chinês barato/quebrado (ex.: switches pretos/azuis de teclado Sentey, como no Mercawa). |
| Keycaps | 24 | Do mesmo teclado reaproveitado. |
| Diodos 1N4148 | 24 | Um por chave, para evitar ghosting na matriz. |
| Caixa 3D printed | 1 | Modelo sugerido: [Stream Deck Macro Keyboard (Printables)](https://www.printables.com/model/269757-stream-deck-macro-keyboard). |
| Cabo USB micro-B | 1 | |
| Fios / solda / ferro | — | |

## Matriz 6×4 (24 botões, 10 pinos)

As 24 chaves são organizadas numa matriz de 6 colunas × 4 linhas. Cada chave
liga uma coluna a uma linha, com um diodo em série (cátodo na linha) para
evitar ghosting.

```
        Col 0  Col 1  Col 2  Col 3  Col 4  Col 5
        (D2)   (D3)   (D4)   (D5)   (D6)   (D7)
Row 0 (D8)  B0     B1     B2     B3     B4     B5
Row 1 (D9)  B6     B7     B8     B9     B10    B11
Row 2 (D10) B12    B13    B14    B15    B16    B17
Row 3 (A0)  B18    B19    B20    B21    B22    B23
```

- **Colunas** (output, LOW ativa): pinos 2, 3, 4, 5, 6, 7.
- **Linhas** (input, pull-up): pinos 8, 9, 10, 14 (A0).
- **Diodo**: ânodo na coluna, cátodo na linha (ou vice-versa; o firmware
  espera LOW na linha quando a coluna está LOW e a chave fechada).

> O índice do botão no RadKeys é `row * 6 + col`. O botão 0 é o `copy` fixo
> (configurável em `[app.fixed_buttons]`), 1 = `level_up`, 2 = `go_home`,
> e 3..23 = configuráveis.

## Protocolo (casa com `internal/hid/reader_cgo.go`)

- HID **vendor-defined**, usage page `0xFF00`, usage `0x0001`.
- Input report **ID 1**, payload **24 bytes**, 1 byte por botão
  (`0x00` solto, `0x01` pressionado).
- O host (hidapi/hidraw) recebe `[0x01, b0..b23]` (25 bytes). O `diyReader`
  aceita também 24 bytes diretos (sem o report ID) para tolerância de backend.

## VID/PID

`0x1234`/`0xABCD` são placeholders. Defina os seus (qualquer par livre; evite
conflitos com dispositivos conhecidos) e **case com `radkeys.config.toml`**:

```toml
[app.device]
vendor_id  = 0x1234
product_id = 0xABCD
protocol   = "radkeys-diy"
```

## Permissão Linux (udev)

Crie `/etc/udev/rules.d/49-radkeys.rules`:

```
KERNEL=="hidraw*", SUBSYSTEM=="hidraw", ATTRS{idVendor}=="1234", ATTRS{idProduct}=="abcd", MODE="0660", GROUP="input"
```

E adicione seu usuário ao grupo `input`: `sudo adduser $USER input` (log out/in).

## Gravação

1. Abra `diy24.ino` no Arduino IDE.
2. Placa: "Arduino Leonardo" ou "Arduino Micro".
3. Porta: a porta serial do Pro Micro.
4. Compile e grave.

> O Pro Micro aparece como um dispositivo HID composto (teclado + vendor-defined).
> O RadKeys lê apenas o report vendor-defined; o teclado padrão não é usado.

## Teste sem hardware

O RadKeys, sem dispositivo, cai no mock e a UI funciona por clique de mouse
(`./radkeys` com `[app.device]` apontando para VID/PID inexistentes).