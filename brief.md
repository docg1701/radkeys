# RadKeys — Brief Técnico

> **Data:** 2026-07-08
> **Repo:** https://github.com/docg1701/radkeys
> **Release atual:** v0.2.1

---

## 1. Estado atual (v0.2.1)

| Componente | Status |
|------------|--------|
| Parser TOML + validação | ✅ `internal/config/` com 6 testes |
| Navegação (deck) | ✅ `internal/deck/` com 8 testes |
| HID reader (Elgato + DIY) | ✅ `internal/hid/` com build tags cgo/!cgo |
| HID mock (dev sem hardware) | ✅ 4 testes |
| UI — aba Atalhos (preview + keypad) | ✅ Preview topo 50%, keypad 4×5 embaixo 50% |
| UI — aba Ajustes (Settings) | ✅ 3 seções (Config, Appearance, USB Device), grid 3 colunas |
| UI — aba About | ✅ Layout limpo, 1 link GitHub, autor docg1701 |
| i18n (7 idiomas) | ✅ Mapa Go único (`internal/i18n/i18n.go`), sem JSON |
| 13 temas de cores | ❌ Ver seção 2.3 — sistema quebrado, todas as cores erradas |
| Ícone do app | ✅ Seletor com preview + Browse (PNG customizado) |
| Firmware Arduino | ✅ `firmware/arduino/diy24.ino` (matriz 6×4) |
| Firmware RP2040 | ✅ `firmware/rp2040/diy24.ino` (24 GPIO) |
| CI | ✅ Linux-only, test + release com binário Linux |
| Release Windows | ✅ Cross-compile local (mingw), upload manual |
| Release Linux | ✅ `go build -tags flatpak` para file dialog nativo via portal |
| Release macOS | ❌ Cross-compile impossível (SDK Apple). Build nativo num Mac. |
| Always-on-top | ⏳ Pendente Fyne v2.8.0 estável |
| Botão Paste | ✅ Adicionado como 4º botão fixo (Copy, Paste, Back, Home) |
| AGENTS.md | ✅ Regras de versão, build em dist/, dev cycle |

## 2. Pendências

### 2.1 macOS

Cross-compile impossível com CGO. Build nativo num Mac.

### 2.2 Always-on-top

Pendente Fyne v2.8.0 estável (função `RequestAlwaysOnTop`).

### 2.3 🔴 GRAVÍSSIMO — Sistema de temas completamente quebrado

**O sistema de temas está quebrado e precisa ser refeito do zero.**
A implementação atual em `internal/theme/custom.go` deriva todas as 28 cores
do Fyne a partir de apenas 3 valores hex (Background, Button, Fixed) usando
funções aritméticas (`lighten`, `darken`, `blend`) com fatores arbitrários.
Isso NÃO funciona — as cores resultantes não respeitam os temas originais
nem garantem contraste adequado.

**Problemas específicos:**

1. **`ColorNamePrimary` usado como cor de texto da aba selecionada**
   (Fyne `container/tabs.go:698`). Foi definido como branco puro (#fff) no
   escuro e preto puro (#000) no claro para forçar legibilidade. Isso quebrou
   o indicador visual da aba (underline) que também usa `ColorNamePrimary`.

2. **`ColorNameForeground` rebaixado para cinza** (`#c8c8c8` escuro, `#333`
   claro) para que abas não selecionadas não se destaquem. Isso afeta TODO
   o texto da interface, deixando-o apagado.

3. **`ColorNameForegroundOnPrimary` definido como oposto de
   `ColorNamePrimary`** para o botão Save ficar legível, ignorando o acento
   do tema.

4. **`headerBg()` usa `lighten(bg, 0.20)` / `darken(bg, 0.12)`** — valores
   arbitrários, sem base em teoria de cor.

5. **`inputBg()` usa `lighten(bg, 0.03-0.06)`** — diferença insignificante,
   campos de input indistinguíveis do fundo.

6. **`hover()` usa `lighten(btn, 0.08)`** — arbitrário.

7. **`ColorNameHyperlink` = `t.fg`** — mesma cor do texto, links invisíveis.

8. **`ColorNameFocus` e `ColorNameSelection` usam `t.fix` com alpha fixo**
   (`0x5c`, `0x40`) — arbitrário.

9. **Múltiplos valores mágicos**: `0xCC` (overlay), `0.20` (input border),
   `0.38` (disabled), `0.42` (placeholder), `0.12` (pressed), `0x5c` (focus),
   `0x40` (selection), `0.45` (disabled button), `0.14` (scrollbar), `0.05`
   (scrollbar bg), `0.12` (separator), `0x10` (shadow).

10. **`dist/radkeys.config.toml` com `preset = "system"`** enquanto
    `radkeys.config.toml` tem `preset = "light_gray"`. O binário em `dist/`
    carrega o config errado. Manter UM único config ou sincronizar.

11. **`variantFor()` detecta light/dark pela luminância do fundo**, mas para
    `theme.DefaultTheme()` isso pode retornar o variant errado dependendo
    de como o tema do sistema é consultado.

**O que fazer:**

Revisar CADA tema visualmente, abrindo o app:
1. Dracula, 2. Solarized Dark, 3. Monokai, 4. Gruvbox Dark, 5. Nord,
6. One Dark, 7. Tokyo Night, 8. Catppuccin Mocha, 9. Solarized Light,
10. Gruvbox Light, 11. Light Gray, 12. Dark Gray, 13. Padrão do sistema.

Para cada tema verificar:
- Fundo geral, preview, keypad (botões inativos).
- Texto legível (contraste ≥ 4.5:1).
- Campos de input com bordas visíveis.
- Abas: selecionada destacada, não selecionadas recuadas.
- Links identificáveis.
- Botão Save sempre visível.
- Trocar de tema sem resíduos.

**Proibido:** testes automatizados. Verificação é visual.

## 3. Estrutura do repo

```
radkeys/
├── AGENTS.md / README.md / LICENSE / brief.md
├── main.go / go.mod / go.sum
├── radkeys.config.toml
├── dist/                        # Binários de release (gitignored)
├── internal/
│   ├── config/    config.go     # Parser TOML + tipos + validação
│   ├── deck/      deck.go       # Estado de navegação (Copy/Paste/Back/Home)
│   ├── hid/       hid.go        # Interface + mock + go-hid real
│   ├── ui/        ui.go         # Fyne UI (Atalhos + Ajustes + About)
│   ├── i18n/      i18n.go       # Mapa Go único com 7 idiomas
│   ├── theme/     custom.go     # Tema customizado (28 ThemeColorName)
│   │              presets.go    # 13 presets
│   └── assets/    assets.go     # Ícone default + ícones Obsidian embed
├── firmware/arduino/            # Arduino Pro Micro
├── firmware/rp2040/             # RP2040
└── research/                    # Notas técnicas
```

## 4. Regras de versão

- **Única fonte da verdade:** `[app] version` em `radkeys.config.toml`.
- Nunca hardcodar versão em código Go, templates ou testes.
- `ensureConfig` em `main.go` NÃO inclui `version` (vazio = fallback "dev").
- Test fixtures usam `"0.0.0-test"`.
