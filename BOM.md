# RadKeys — Bill of Materials

> Prices in BRL (July 2026). Verified links for AliExpress, Mercado Livre, and Shopee (Brazil).

## 1. RP2040-Zero (×1)

| Platform | Price | Link |
|----------|-------|------|
| **AliExpress** Waveshare original | ~R$ 12.89 | [aliexpress.com/item/1005009682636404](https://pt.aliexpress.com/item/1005009682636404.html) |
| **AliExpress** Tenstar (generic) | ~R$ 5.99 | [aliexpress.com](https://pt.aliexpress.com/wholesale?SearchText=RP2040-Zero) |
| **Shopee** | ~R$ 13.40 | [shopee.com.br/list/Raspberry/Pi%20Zero](https://shopee.com.br/list/Raspberry/Pi%20Zero) |
| **Mercado Livre** (pre-soldered pins) | ~R$ 42 | [lista.mercadolivre.com.br/rp2040-zero](https://lista.mercadolivre.com.br/rp2040-zero) |

> Mercado Livre is more expensive (~R$ 42) but delivers in 2–3 days. AliExpress: free shipping, 15–30 days. Shopee: domestic shipping, ~7–15 days.

## 2. Push buttons 6×6mm 4-pin — 36 units

Buy a 100-pack (spares for future replacement).

| Platform | Price (100 pcs) | Link |
|----------|-----------------|------|
| **Shopee** | ~R$ 15.06 | [shopee.com.br](https://shopee.com.br/list/Bot%C3%A3o/Liga%20Desliga) |
| **Mercado Livre** | ~R$ 20–30 | [lista.mercadolivre.com.br](https://lista.mercadolivre.com.br) |
| **AliExpress** | ~R$ 8–12 | [aliexpress.com](https://pt.aliexpress.com/wholesale?SearchText=push+button+6x6+4+pin+100pcs) |

> Pick 7mm or 9mm height (taller = more comfortable). 5mm is too low for a keypad.

## 3. 1N4148 through-hole diodes — 100-pack

| Platform | Price | Link |
|----------|-------|------|
| **Mercado Livre** | R$ 27.80 | [produto.mercadolivre.com.br/MLB-1725170449](https://produto.mercadolivre.com.br/MLB-1725170449-1-kit-com-100-unidades-diodo-de-sinal-1n4148-_JM) |
| **Shopee** | ~R$ 13.50 | [shopee.com.br](https://shopee.com.br/Kit-100-Unidades-Diodo-1N4148-i.277189279.15378905686) |
| **AliExpress** | ~R$ 5–8 | [aliexpress.com](https://pt.aliexpress.com/wholesale?SearchText=1N4148+diode+100pcs) |

## 4. Wiring — M-F Dupont jumpers + Ethernet cable for buses

6 jumpers per device (cut in half → 12 wires: 6 rows + 6 columns).
Bus wire: any old Ethernet cable (8 AWG 24 solid-core wires, free).

| Platform | Item | Price | Devices covered |
|----------|------|-------|----------------|
| **Mercado Livre** | 60 M-F jumpers 20 cm | ~R$ 12 | 10 |
| **Shopee** | 20–40 M-F jumpers kit | ~R$ 8–15 | 3–6 |
| **AliExpress** | 40–65 M-F jumpers kit | ~R$ 5–10 | 6–10 |

## 5. Zip ties — 100 pcs, 2.5×100mm

| Platform | Price | Link |
|----------|-------|------|
| **Shopee** | ~R$ 10 | [shopee.com.br](https://shopee.com.br/Abra%C3%A7adeira-de-NYLON-2-5x100mm-100-unidades-BRANCA-i.696093200.19699179805) |
| **Mercado Livre** | ~R$ 10 | [produto.mercadolivre.com.br/MLB-3790199176](https://produto.mercadolivre.com.br/MLB-3790199176-kit-abracadeira-de-nylon-100-pecas-25x100-mm-branco-_JM) |

## 6. USB-C cable (×1)

| Platform | Price | Link |
|----------|-------|------|
| **Shopee** | ~R$ 12 | [shopee.com.br](https://shopee.com.br/Cabo-USB-Tipo-C-Carregamento-R%C3%A1pido-Turbo-Cabo-de-Dados-1-Metro-i.1219810528.23293487068) |
| **Mercado Livre** | ~R$ 19 | [produto.mercadolivre.com.br/MLB-5240845278](https://produto.mercadolivre.com.br/MLB-5240845278-cabo-tipo-c-1m-compativel-usb-c-carregador-e-dados-_JM) |

> Any USB-C data cable works. The RP2040-Zero has an on-board USB-C port.

## 7. 3D-printed case (optional)

| Material | 1 kg spool | Notes |
|----------|-----------|-------|
| PETG | ~R$ 66–110 | Strong, heat-resistant, no enclosure needed |
| ABS | ~R$ 55–123 | Requires heated enclosure, more warp-prone |

One spool covers multiple devices (exact count TBD — model not finalized).
Alternatively, hire a printing service (~R$ 25–45 per case on Shopee/ML).

STL and FreeCAD files: [`firmware/rp2040-zero/case/`](firmware/rp2040-zero/case/)

## Summary — per-device cost

### 🛒 AliExpress (lowest price, free shipping, ~20–30 days)
| Item | Kit qty | Kit price | Per device | Devices/kit |
|------|---------|-----------|------------|-------------|
| RP2040-Zero | 1 | R$ 6 | R$ 6.00 | 1 |
| Push buttons 6×6 | 100 pcs | R$ 10 | R$ 3.60 | 2.7 |
| 1N4148 diodes | 100 pcs | R$ 6 | R$ 2.16 | 2.7 |
| M-F Dupont jumpers | 40 pcs | R$ 8 | R$ 1.20 | 6.6 |
| Zip ties | 100 pcs | R$ 5 | R$ 0.50 | 10 |
| USB-C cable | 1 | R$ 8 | R$ 8.00 | 1 |
| **Total per device** | | | **~R$ 21** | |

### 📦 Shopee (mid-range, ~7–15 days)
| Item | Kit qty | Kit price | Per device | Devices/kit |
|------|---------|-----------|------------|-------------|
| RP2040-Zero | 1 | R$ 14 | R$ 14.00 | 1 |
| Push buttons 6×6 | 100 pcs | R$ 15 | R$ 5.40 | 2.7 |
| 1N4148 diodes | 100 pcs | R$ 13.50 | R$ 4.86 | 2.7 |
| M-F Dupont jumpers | 30 pcs | R$ 10 | R$ 2.00 | 5 |
| Zip ties | 100 pcs | R$ 10 | R$ 1.00 | 10 |
| USB-C cable | 1 | R$ 12 | R$ 12.00 | 1 |
| **Total per device** | | | **~R$ 39** | |

### 🏠 Mercado Livre (highest price, fast delivery ~2–5 days)
| Item | Kit qty | Kit price | Per device | Devices/kit |
|------|---------|-----------|------------|-------------|
| RP2040-Zero (pre-soldered) | 1 | R$ 42 | R$ 42.00 | 1 |
| Push buttons 6×6 | 100 pcs | R$ 25 | R$ 9.00 | 2.7 |
| 1N4148 diodes | 100 pcs | R$ 27.80 | R$ 10.01 | 2.7 |
| M-F Dupont jumpers | 60 pcs | R$ 12 | R$ 1.20 | 10 |
| Zip ties | 100 pcs | R$ 10 | R$ 1.00 | 10 |
| USB-C cable | 1 | R$ 19 | R$ 19.00 | 1 |
| **Total per device** | | | **~R$ 82** | |

> Case: PETG ~R$ 66–110/spool or ABS ~R$ 55–123/spool. Add ~R$ 5–15 per device.
> Tools (soldering iron, cutting pliers) not included — builder's responsibility.
