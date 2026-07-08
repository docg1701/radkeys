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
| 13 temas de cores | ✅ `internal/theme/custom.go` — 28 ThemeColorName explícitos |
| Ícone do app | ✅ Seletor com preview + Browse (PNG customizado) |
| Firmware Arduino | ✅ `firmware/arduino/diy24.ino` (matriz 6×4) |
| Firmware RP2040 | ✅ `firmware/rp2040/diy24.ino` (24 GPIO) |
| CI | ✅ Linux-only, test + release com binário Linux |
| Release Windows | ✅ Cross-compile local (mingw), upload manual |
| Release macOS | ❌ Cross-compile impossível (SDK Apple). Build nativo num Mac. |
| Always-on-top | ⏳ Pendente Fyne v2.8.0 estável |
| Botão Paste | ✅ Adicionado como 4º botão fixo (Copy, Paste, Back, Home) |
| AGENTS.md | ✅ Regras de versão, build em dist/, dev cycle |
| README.md | ✅ Dependências, build, cross-compile |

## 2. Pendências

### 2.1 File dialog usa locale do SO

O `dialog.NewFileOpen` do Fyne usa o idioma do sistema operacional, não o
idioma selecionado no app. Sem API no Fyne para sobrescrever. Workaround:
executar com `LANG=en_US.UTF-8`.

### 2.2 macOS

Cross-compile impossível com CGO. Build nativo num Mac. Pendente.

### 2.3 Always-on-top

Pendente Fyne v2.8.0 estável (função `RequestAlwaysOnTop`).

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
