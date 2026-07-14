# RadKeys — Bill of Materials

Prices in BRL (July 2026). Each link points to a specific product page.

## 1. RP2040-Zero — Waveshare, 2 MB Flash, USB-C, 29 GPIO (×1)

Do NOT buy "RP2040-Zero-M" — different pin count, firmware-incompatible.

| Platform | Price | Product page |
|----------|-------|--------------|
| AliExpress original | ~R$ 12.89 | [aliexpress.com/item/1005009682636404](https://pt.aliexpress.com/item/1005009682636404.html) |
| AliExpress generic | ~R$ 6.99 | [aliexpress.com/item/1005011886603196](https://pt.aliexpress.com/item/1005011886603196.html) |
| Mercado Livre | ~R$ 42 | [produto.mercadolivre.com.br/MLB-7167143866](https://produto.mercadolivre.com.br/MLB-7167143866--bby-placa-compativel-rp2040-zero-para-raspberry-pi-pico--_JM) |
| Shopee | ~R$ 25 | [shopee.com.br/list/Raspberry/Pi%20Zero](https://shopee.com.br/list/Raspberry/Pi%20Zero) |

## 2. Push buttons 6×6mm 4-pin, 7–9mm height (×100)

| Platform | Price | Product page |
|----------|-------|--------------|
| AliExpress | ~R$ 8.84 | [aliexpress.com/item/1005008683932324](https://pt.aliexpress.com/item/1005008683932324.html) |
| Mercado Livre | ~R$ 25 | [produto.mercadolivre.com.br/MLB-4444543388](https://produto.mercadolivre.com.br/MLB-4444543388-100-x-push-button-micro-chave-4-pinos-6x6-_JM) |
| Shopee | ~R$ 15 | [shopee.com.br/list/Botão/Liga%20Desliga](https://shopee.com.br/list/Bot%C3%A3o/Liga%20Desliga) |

## 3. 1N4148 through-hole diodes DO-35 (×100)

| Platform | Price | Product page |
|----------|-------|--------------|
| AliExpress | ~R$ 1.27 | [aliexpress.com/item/32968119428](https://pt.aliexpress.com/item/32968119428.html) |
| Shopee | ~R$ 13.50 | [shopee.com.br/Kit-100-Unidades-Diodo-1N4148-i.277189279.15378905686](https://shopee.com.br/Kit-100-Unidades-Diodo-1N4148-i.277189279.15378905686) |
| Mercado Livre | R$ 27.80 | [produto.mercadolivre.com.br/MLB-1725170449](https://produto.mercadolivre.com.br/MLB-1725170449-1-kit-com-100-unidades-diodo-de-sinal-1n4148-_JM) |

## 4. Wiring — M-F Dupont jumpers + Ethernet bus wire

6 M-F jumpers per device (cut in half → 12 wires). Bus wire: old Ethernet
cable (8 solid-color AWG 24 wires, free).

| Platform | Item | Price | Product page |
|----------|------|-------|--------------|
| AliExpress | 40-pin M-M + M-F + F-F kit, 20 cm | ~R$ 1.94 | [aliexpress.com/item/1005012508436047](https://pt.aliexpress.com/item/1005012508436047.html) |
| Mercado Livre | 60 M-F jumpers, 20 cm | ~R$ 12 | [produto.mercadolivre.com.br/MLB-1020024323](https://produto.mercadolivre.com.br/MLB-1020024323-kit-cabo-jumper-wire-20cm-60-pecas-p-arduino-veja-anuncio-_JM) |
| Shopee | 40 M-F jumpers, 20 cm | ~R$ 12 | [shopee.com.br/list/Conector/Femea](https://shopee.com.br/list/Conector/Femea) |

## 5. Zip ties 2.5×100mm (×100)

| Platform | Price | Product page |
|----------|-------|--------------|
| Shopee | ~R$ 10 | [shopee.com.br/Abraçadeira-de-NYLON-2-5x100mm-100-unidades-BRANCA-i.696093200.19699179805](https://shopee.com.br/Abra%C3%A7adeira-de-NYLON-2-5x100mm-100-unidades-BRANCA-i.696093200.19699179805) |
| Mercado Livre | ~R$ 10 | [produto.mercadolivre.com.br/MLB-3790199176](https://produto.mercadolivre.com.br/MLB-3790199176-kit-abracadeira-de-nylon-100-pecas-25x100-mm-branco-_JM) |

## 6. USB-C data cable 1 m (×1)

| Platform | Price | Product page |
|----------|-------|--------------|
| Shopee | ~R$ 12 | [shopee.com.br/Cabo-USB-Tipo-C-Carregamento-Rápido-Turbo-Cabo-de-Dados-1-Metro-i.1219810528.23293487068](https://shopee.com.br/Cabo-USB-Tipo-C-Carregamento-R%C3%A1pido-Turbo-Cabo-de-Dados-1-Metro-i.1219810528.23293487068) |
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

### AliExpress (~20–30 days, free shipping)
| Item | Kit | Price | Per device | Covers |
|------|-----|-------|------------|--------|
| RP2040-Zero | 1 pc | R$ 7 | R$ 7.00 | 1 |
| Push buttons | 100 pcs | R$ 9 | R$ 3.24 | 2.7 |
| 1N4148 diodes | 100 pcs | R$ 1 | R$ 0.36 | 2.7 |
| M-F jumpers | 40-pin kit | R$ 2 | R$ 0.30 | 6.6 |
| Zip ties | 100 pcs | R$ 5 | R$ 0.50 | 10 |
| USB-C cable | 1 pc | R$ 8 | R$ 8.00 | 1 |
| **Total per device** | | | **~R$ 19** | |

### Shopee (~7–15 days)
| Item | Kit | Price | Per device | Covers |
|------|-----|-------|------------|--------|
| RP2040-Zero | 1 pc | R$ 25 | R$ 25.00 | 1 |
| Push buttons | 100 pcs | R$ 15 | R$ 5.40 | 2.7 |
| 1N4148 diodes | 100 pcs | R$ 14 | R$ 5.04 | 2.7 |
| M-F jumpers | 40 pcs | R$ 12 | R$ 1.80 | 6.6 |
| Zip ties | 100 pcs | R$ 10 | R$ 1.00 | 10 |
| USB-C cable | 1 pc | R$ 12 | R$ 12.00 | 1 |
| **Total per device** | | | **~R$ 50** | |

### Mercado Livre (~2–5 days)
| Item | Kit | Price | Per device | Covers |
|------|-----|-------|------------|--------|
| RP2040-Zero | 1 pc | R$ 42 | R$ 42.00 | 1 |
| Push buttons | 100 pcs | R$ 25 | R$ 9.00 | 2.7 |
| 1N4148 diodes | 100 pcs | R$ 28 | R$ 10.08 | 2.7 |
| M-F jumpers | 60 pcs | R$ 12 | R$ 1.20 | 10 |
| Zip ties | 100 pcs | R$ 10 | R$ 1.00 | 10 |
| USB-C cable | 1 pc | R$ 19 | R$ 19.00 | 1 |
| **Total per device** | | | **~R$ 82** | |

> Case: add ~R$ 5–15 for filament, or ~R$ 25–45 for a printing service.
> Tools (soldering iron, cutters) not included — builder's responsibility.
