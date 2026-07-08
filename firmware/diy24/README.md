# RadKeys DIY 24 — firmware do dispositivo HID custom

Dispositivo USB de 24 botões que o RadKeys lê diretamente via hidapi
(`protocol = "radkeys-diy"`). **Não envia teclas** — é um HID vendor-defined,
logo não rouba foco do RIS e não interfere em atalhos do app focado.

## Plataforma recomendada: Raspberry Pi Pico (RP2040)

- 26 GPIO (cabe 24 botões direto, sem multiplexador) e HID custom maduro via
  **Adafruit_TinyUSB**. Custo ~R$20-30.
- No Arduino IDE: instale o core da RP2040 (ex.: "Raspberry Pi Pico/RP2040"
  by earlephilhower) e a lib **Adafruit TinyUSB Library** (Library Manager).
- Selecione a placa "Raspberry Pi Pico" e o modo "TinyUSB".
- Abra `diy24.ino`, ajuste `DIY_VID`/`DIY_PID` e a pinagem `PINS`, e grave.

## Pinagem

24 chaves, cada uma de um GPIO (default GPIO0..GPIO23) para GND. O pull-up
interno é habilitado; botão pressionado = LOW = `0x01` no report.

> O índice do botão no RadKeys é a posição no array `PINS`. O botão 0 é o
> `copy` fixo (configurável em `[app.fixed_buttons]`), 1 = `level_up`,
> 2 = `go_home`, e 3..23 = configuráveis.

## Protocolo (casa com `internal/hid/reader_cgo.go`)

- HID **vendor-defined**, usage page `0xFF00`, usage `0x0001`.
- Input report **ID 1**, payload **24 bytes**, 1 byte por botão
  (`0x00` solto, `0x01` pressionado).
- O host (hidapi/hidraw) recebe `[0x01, b0..b23]` (25 bytes). O `diyReader`
  aceita também 24 bytes diretos (sem o report ID) para tolerância de backend.

## VID/PID

`0x1234`/`0xABCD` são placeholders. Defina os seus (qualquer par livre; evite
conflitos com dispositivos conhecidos) e **caso com `radkeys.config.toml`**:

```toml
[app.device]
vendor_id  = 0x1234
product_id = 0xABCD
protocol   = "radkeys-diy"
```

No Linux, crie uma regra udev para acesso a `/dev/hidraw*` sem sudo, por exemplo
`/etc/udev/rules.d/49-radkeys.rules`:

```
KERNEL=="hidraw*", SUBSYSTEM=="hidraw", ATTRS{idVendor}=="1234", ATTRS{idProduct}=="abcd", MODE="0660", GROUP="input"
```

(e adicione seu usuário ao grupo `input`: `sudo adduser $USER input`)

## Alternativa: Arduino Pro Micro (ATmega32U4)

O Pro Micro tem poucos GPIO livres (~12-18); para 24 botões diretos é preciso
um multiplexador (ex.: 74HC4067) ou uma matriz. HID vendor-defined no 32U4
exige descritor raw via a lib `HID-Project` ou o core `HID`. Se usar Pro Micro,
porte o descritor e o `sendReport` para a API da lib escolhida; o protocolo
(report ID 1 + 24 bytes) é o mesmo. O RP2040 é mais simples para 24 botões.

## Teste sem hardware

O RadKeys, sem dispositivo, cai no mock e a UI funciona por clique de mouse
(`./radkeys` com `[app.device]` apontando para VID/PID inexistentes).