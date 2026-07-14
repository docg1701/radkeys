# RadKeys — Bill of Materials

> Prices in BRL (July 2026). Product links for AliExpress, Mercado Livre, and Shopee (Brazil).

## 1. RP2040-Zero (Waveshare, ×1)

The firmware targets the **Waveshare RP2040-Zero** (2 MB flash, USB-C,
29 GPIO). Clones with the same pinout work. Do NOT buy the "RP2040-Zero-M"
(mini variant — different pin count).

| Platform | Price | Link |
|----------|-------|------|
| AliExpress (Waveshare) | ~R$ 12.89 | [aliexpress.com/item/1005009682636404](https://pt.aliexpress.com/item/1005009682636404.html) |
| AliExpress (generic) | ~R$ 5.99 | [aliexpress.com](https://pt.aliexpress.com/wholesale?SearchText=waveshare+RP2040-Zero) |
| Shopee | ~R$ 25 | [shopee.com.br/list/Raspberry/Pi%20Zero](https://shopee.com.br/list/Raspberry/Pi%20Zero) |
| Mercado Livre | ~R$ 42 | [lista.mercadolivre.com.br/rp2040-zero](https://lista.mercadolivre.com.br/rp2040-zero) |

## 2. Push buttons 6×6mm 4-pin (×100)

Buy a 100-pack — spares for 2–3 devices.

| Platform | Price | Link |
|----------|-------|------|
| Shopee | ~R$ 15 | [shopee.com.br/list/Botão/Liga%20Desliga](https://shopee.com.br/list/Bot%C3%A3o/Liga%20Desliga) |
| Mercado Livre | ~R$ 25 | [lista.mercadolivre.com.br](https://lista.mercadolivre.com.br) — search "push button 6x6 100" |
| AliExpress | ~R$ 10 | [aliexpress.com](https://pt.aliexpress.com/wholesale?SearchText=push+button+6x6+4+pin+100pcs) |

## 3. 1N4148 through-hole diodes (×100)

| Platform | Price | Link |
|----------|-------|------|
| Shopee | ~R$ 13.50 | [shopee.com.br/Kit-100-Unidades-Diodo-1N4148](https://shopee.com.br/Kit-100-Unidades-Diodo-1N4148-i.277189279.15378905686) |
| Mercado Livre | R$ 27.80 | [produto.mercadolivre.com.br/MLB-1725170449](https://produto.mercadolivre.com.br/MLB-1725170449-1-kit-com-100-unidades-diodo-de-sinal-1n4148-_JM) |
| AliExpress | ~R$ 6 | [aliexpress.com](https://pt.aliexpress.com/wholesale?SearchText=1N4148+diode+100pcs) |

## 4. Wiring — M-F Dupont jumpers + Ethernet bus wire

6 M-F jumpers per device (cut in half → 12 wires). Bus wire: any old Ethernet
cable (8 solid-color AWG 24 wires, free).

| Platform | Item | Price | Devices |
|----------|------|-------|---------|
| Mercado Livre | 60 M-F 20 cm | ~R$ 12 | 10 |
| Shopee | 40 M-F 20 cm | ~R$ 15 | 6 |
| AliExpress | 65 M-F 20 cm | ~R$ 8 | 10 |

| Platform | Link |
|----------|------|
| Mercado Livre | [produto.mercadolivre.com.br/MLB-1020024323](https://produto.mercadolivre.com.br/MLB-1020024323-kit-cabo-jumper-wire-20cm-60-pecas-p-arduino-veja-anuncio-_JM) |
| Shopee | [shopee.com.br/list/Conector/Femea](https://shopee.com.br/list/Conector/Femea) |
| AliExpress | [aliexpress.com](https://pt.aliexpress.com/wholesale?SearchText=dupont+jumper+male+female+20cm+kit) |

## 5. Zip ties (×100, 2.5×100mm)

| Platform | Price | Link |
|----------|-------|------|
| Shopee | ~R$ 10 | [shopee.com.br/Abraçadeira-de-NYLON-2-5x100mm](https://shopee.com.br/Abra%C3%A7adeira-de-NYLON-2-5x100mm-100-unidades-BRANCA-i.696093200.19699179805) |
| Mercado Livre | ~R$ 10 | [produto.mercadolivre.com.br/MLB-3790199176](https://produto.mercadolivre.com.br/MLB-3790199176-kit-abracadeira-de-nylon-100-pecas-25x100-mm-branco-_JM) |

## 6. USB-C cable (×1)

| Platform | Price | Link |
|----------|-------|------|
| Shopee | ~R$ 12 | [shopee.com.br/Cabo-USB-Tipo-C](https://shopee.com.br/Cabo-USB-Tipo-C-Carregamento-R%C3%A1pido-Turbo-Cabo-de-Dados-1-Metro-i.1219810528.23293487068) |
| Mercado Livre | ~R$ 19 | [produto.mercadolivre.com.br/MLB-5240845278](https://produto.mercadolivre.com.br/MLB-5240845278-cabo-tipo-c-1m-compativel-usb-c-carregador-e-dados-_JM) |

## 7. 3D-printed case (optional)

| Material | 1 kg spool | Notes |
|----------|-----------|-------|
| PETG | ~R$ 66–110 | No enclosure needed, ideal for final build |
| ABS | ~R$ 55–123 | Requires heated enclosure |

One spool covers multiple devices (exact count TBD).
Hire a printing service: ~R$ 25–45 per case on Shopee/ML.

STL and FreeCAD files: [`firmware/rp2040-zero/case/`](firmware/rp2040-zero/case/)

## Per-device cost

### 🛒 AliExpress (~20–30 days)
| Item | Kit | Price | Per device | Covers |
|------|-----|-------|------------|--------|
| RP2040-Zero Waveshare | 1 pc | R$ 13 | R$ 13.00 | 1 |
| Push buttons | 100 pcs | R$ 10 | R$ 3.60 | 2.7 |
| 1N4148 diodes | 100 pcs | R$ 6 | R$ 2.16 | 2.7 |
| M-F jumpers | 65 pcs | R$ 8 | R$ 0.74 | 10.8 |
| Zip ties | 100 pcs | R$ 5 | R$ 0.50 | 10 |
| USB-C cable | 1 pc | R$ 8 | R$ 8.00 | 1 |
| **Total per device** | | | **~R$ 28** | |

### 📦 Shopee (~7–15 days)
| Item | Kit | Price | Per device | Covers |
|------|-----|-------|------------|--------|
| RP2040-Zero | 1 pc | R$ 25 | R$ 25.00 | 1 |
| Push buttons | 100 pcs | R$ 15 | R$ 5.40 | 2.7 |
| 1N4148 diodes | 100 pcs | R$ 13.50 | R$ 4.86 | 2.7 |
| M-F jumpers | 40 pcs | R$ 15 | R$ 2.25 | 6.6 |
| Zip ties | 100 pcs | R$ 10 | R$ 1.00 | 10 |
| USB-C cable | 1 pc | R$ 12 | R$ 12.00 | 1 |
| **Total per device** | | | **~R$ 51** | |

### 🏠 Mercado Livre (~2–5 days)
| Item | Kit | Price | Per device | Covers |
|------|-----|-------|------------|--------|
| RP2040-Zero | 1 pc | R$ 42 | R$ 42.00 | 1 |
| Push buttons | 100 pcs | R$ 25 | R$ 9.00 | 2.7 |
| 1N4148 diodes | 100 pcs | R$ 28 | R$ 10.08 | 2.7 |
| M-F jumpers | 60 pcs | R$ 12 | R$ 1.20 | 10 |
| Zip ties | 100 pcs | R$ 10 | R$ 1.00 | 10 |
| USB-C cable | 1 pc | R$ 19 | R$ 19.00 | 1 |
| **Total per device** | | | **~R$ 82** | |

> Case: add ~R$ 5–15 per device for filament, or ~R$ 25–45 for a printing service.
> Tools (soldering iron, cutters) not included — builder's responsibility.
