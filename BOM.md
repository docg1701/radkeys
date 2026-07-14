# RadKeys — Bill of Materials

> Prices in BRL (July 2026). Verified links for AliExpress, Mercado Livre, and Shopee (Brazil).

---

## 1. RP2040-Zero (×1)

| Platform | Price | Link |
|----------|-------|------|
| **AliExpress** Waveshare original | ~R$ 12.89 | [aliexpress.com/item/1005009682636404](https://pt.aliexpress.com/item/1005009682636404.html) |
| **AliExpress** Tenstar (generic) | ~R$ 5.99 | [aliexpress.com](https://pt.aliexpress.com/wholesale?SearchText=RP2040-Zero) — search "RP2040-Zero" |
| **Shopee** | ~R$ 13.40 | [shopee.com.br/list/Raspberry/Pi%20Zero](https://shopee.com.br/list/Raspberry/Pi%20Zero) |
| **Mercado Livre** (pre-soldered pins) | ~R$ 42 | [lista.mercadolivre.com.br/rp2040-zero](https://lista.mercadolivre.com.br/rp2040-zero) |

> Mercado Livre is more expensive (~R$ 42) but delivers in 2–3 days. AliExpress: free shipping, 15–30 days. Shopee: domestic shipping, ~7–15 days.

---

## 2. Push buttons 6×6mm 4-pin — 36 units

**Buy a 100-pack** (spares for future replacement).

| Platform | Price (100 pcs) | Link |
|----------|-----------------|------|
| **Shopee** | ~R$ 15.06 | [shopee.com.br](https://shopee.com.br/list/Bot%C3%A3o/Liga%20Desliga) — search "100 Pcs Tact Button Switch 6x6 4 pin" |
| **Mercado Livre** | ~R$ 20–30 | [lista.mercadolivre.com.br](https://lista.mercadolivre.com.br) — search "push button 6x6 100" |
| **AliExpress** | ~R$ 8–12 | [aliexpress.com](https://pt.aliexpress.com/wholesale?SearchText=push+button+6x6+4+pin+100pcs) |

> Pick 7mm or 9mm height (taller = more comfortable). 5mm is too low for a keypad.

---

## 3. 1N4148 through-hole diodes — 36 units

**Buy a 50- or 100-pack** (cost is nearly the same as 36 loose units).

| Platform | Price | Qty | Link |
|----------|-------|-----|------|
| **Mercado Livre** | R$ 27.80 | 100 pcs | [produto.mercadolivre.com.br/MLB-1725170449](https://produto.mercadolivre.com.br/MLB-1725170449-1-kit-com-100-unidades-diodo-de-sinal-1n4148-_JM) |
| **Shopee** | ~R$ 8–15 | 100 pcs | [shopee.com.br](https://shopee.com.br) — search "1N4148 diode 100" |
| **AliExpress** | ~R$ 5–8 | 100 pcs | [aliexpress.com](https://pt.aliexpress.com/wholesale?SearchText=1N4148+diode+100pcs) |

---

## 4. Colored wire kit AWG 24–26

| Platform | Price | Link |
|----------|-------|------|
| **AliExpress** | ~R$ 6.99 (120 pcs, 6 colors) | [aliexpress.com/w/wholesale-wiring-kit](https://pt.aliexpress.com/w/wholesale-wiring-kit.html) |
| **Mercado Livre** | ~R$ 15–25 | [lista.mercadolivre.com.br](https://lista.mercadolivre.com.br) — search "kit jumper wires 20cm" |
| **Shopee** | ~R$ 10–18 | [shopee.com.br](https://shopee.com.br) — search "colored wire kit awg 24" |

> Free alternative: an Ethernet cable (CAT5/6) has 8 colored AWG 24 wires. 2m of cable gives you enough wire.

---

## 5. Female Dupont connectors (×12)

| Platform | Price | Link |
|----------|-------|------|
| **Mercado Livre** 620-piece kit | ~R$ 35–45 | [lista.mercadolivre.com.br/dupont](https://lista.mercadolivre.com.br/dupont) |
| **Mercado Livre** 60-piece pre-made jumper kit | ~R$ 12 | [produto.mercadolivre.com.br/MLB-1020024323](https://produto.mercadolivre.com.br/MLB-1020024323-kit-cabo-jumper-wire-20cm-60-pecas-p-arduino-veja-anuncio-_JM) |
| **Shopee** | ~R$ 8.25 (10 connectors) | [shopee.com.br](https://shopee.com.br) — search "dupont female connector kit" |

> Pre-made jumper kits: cut wires in half and solder directly — no crimping needed.

---

## 6. Zip ties (small)

| Platform | Price | Link |
|----------|-------|------|
| **Shopee** 100 pcs 2.5×100mm | ~R$ 6–10 | [shopee.com.br](https://shopee.com.br) — search "zip ties 100 2.5x100" |
| **Mercado Livre** 100 pcs | ~R$ 10–15 | [lista.mercadolivre.com.br](https://lista.mercadolivre.com.br) — search "zip tie 100" |

---

## 7. USB-C cable (×1)

| Platform | Price | Link |
|----------|-------|------|
| **Shopee** | ~R$ 8–15 | [shopee.com.br](https://shopee.com.br) — search "usb-c data cable" |
| **Mercado Livre** | ~R$ 12–20 | [lista.mercadolivre.com.br](https://lista.mercadolivre.com.br) — search "usb-c cable" |

> The RP2040-Zero has a USB-C port on-board. Any USB-C data cable works (charge-only cables won't).

---

## 8. 3D-printed case (optional)

Print in PETG or ABS. One 1 kg spool makes multiple cases (exact count TBD —
case model is not finalized).

| Material | 1 kg spool | Notes |
|----------|-----------|-------|
| PETG | ~R$ 66–110 | Strong, heat-resistant, no enclosure needed — ideal for final build |
| ABS | ~R$ 55–123 | Requires heated enclosure, more warp-prone |

> Prices from Shopee and Mercado Livre (July 2026). Outlet/generic rolls at the low
> end; Creality/eSUN at the high end.

Alternatively, hire a printing service (~R$ 25–45 per case on Shopee/ML).

STL and FreeCAD files: [`firmware/rp2040-zero/case/`](firmware/rp2040-zero/case/)

---

## Summary — per-device cost

Each device needs: 1× RP2040-Zero, 36× push buttons, 36× diodes, 12× Dupont
connectors, ~2 m of wire, 1× USB-C cable, ~10 zip ties, 1× case. Kits cover
multiple devices; cost per device = kit price ÷ devices covered.

### 🛒 AliExpress (lowest price, free shipping, ~20–30 days)
| Item | Kit qty | Kit price | Per device | Devices/kit |
|------|---------|-----------|------------|-------------|
| RP2040-Zero | 1 | R$ 6 | R$ 6.00 | 1 |
| Push buttons 6×6 | 100 pcs | R$ 10 | R$ 3.60 | 2.7 |
| 1N4148 diodes | 100 pcs | R$ 6 | R$ 2.16 | 2.7 |
| Colored wire kit | 1 kit | R$ 7 | R$ 7.00 | 1 |
| Dupont connectors | 1 kit | R$ 10 | R$ 10.00 | 1 |
| Zip ties | 100 pcs | R$ 5 | R$ 0.50 | 10 |
| USB-C cable | 1 | R$ 8 | R$ 8.00 | 1 |
| **Total per device** | | | **~R$ 37** | |

### 📦 Shopee (mid-range, ~7–15 days)
| Item | Kit qty | Kit price | Per device | Devices/kit |
|------|---------|-----------|------------|-------------|
| RP2040-Zero | 1 | R$ 14 | R$ 14.00 | 1 |
| Push buttons 6×6 | 100 pcs | R$ 15 | R$ 5.40 | 2.7 |
| 1N4148 diodes | 100 pcs | R$ 12 | R$ 4.32 | 2.7 |
| Wires + Dupont kit | 1 kit | R$ 15 | R$ 15.00 | 1 |
| Zip ties | 100 pcs | R$ 6 | R$ 0.60 | 10 |
| USB-C cable | 1 | R$ 10 | R$ 10.00 | 1 |
| **Total per device** | | | **~R$ 49** | |

### 🏠 Mercado Livre (highest price, fast delivery ~2–5 days)
| Item | Kit qty | Kit price | Per device | Devices/kit |
|------|---------|-----------|------------|-------------|
| RP2040-Zero (pre-soldered) | 1 | R$ 42 | R$ 42.00 | 1 |
| Push buttons 6×6 | 100 pcs | R$ 25 | R$ 9.00 | 2.7 |
| 1N4148 diodes | 100 pcs | R$ 28 | R$ 10.08 | 2.7 |
| Dupont jumper kit | 60 pcs | R$ 12 | R$ 2.40 | 5 |
| Zip ties | 100 pcs | R$ 10 | R$ 1.00 | 10 |
| USB-C cable | 1 | R$ 12 | R$ 12.00 | 1 |
| **Total per device** | | | **~R$ 76** | |

> Case: PETG ~R$ 66–110/spool or ABS ~R$ 55–123/spool. One spool covers
> multiple devices (exact count TBD — model not finalized). Add ~R$ 5–15 per
> device depending on filament cost and case design.
>
> Tools not included — the builder is expected to own a soldering iron, cutting
> pliers, and basic bench tools.
