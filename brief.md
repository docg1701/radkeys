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
| 13 temas de cores | ❌ `internal/theme/custom.go` — quebrados, cores inconsistentes (ver 2.3) |
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
| README.md | ✅ Dependências, build, cross-compile |

## 2. Pendências

### 2.1 macOS

Cross-compile impossível com CGO. Build nativo num Mac. Pendente.

### 2.2 Always-on-top

Pendente Fyne v2.8.0 estável (função `RequestAlwaysOnTop`).

### 2.3 🔴 GRAVÍSSIMO — Revisar e corrigir TODOS os temas

**O sistema de temas está quebrado.** Cada um dos 13 temas precisa ser revisado
**linha por linha**, abrindo o app com o tema selecionado e verificando
visualmente:

- Preview: fundo coerente com o tema, texto legível (contraste adequado).
- Keypad: botões inativos com cor coerente, não pretos em tema claro.
- Settings: campos de texto, selects, labels com cores corretas.
- About: texto, link, fundo tudo coerente.
- Tabs: **Em temas escuros, a aba selecionada tem fundo escuro que esconde o texto.
  As abas não selecionadas têm fundo claro que destaca — é o INVERSO do correto.**
  `ColorNameHeaderBackground` precisa ser mais claro que o fundo em tema escuro,
  e a aba selecionada deve ser a mais destacada.

**Problemas identificados e NÃO resolvidos:**

1. **`ColorNamePrimary` usado como cor de texto da aba selecionada** (Fyne `container/tabs.go:698`).
   Definido como branco puro no escuro / preto puro no claro para contraste máximo.
   Isso quebra o indicador visual da aba (underline) que também usa `ColorNamePrimary`
   e agora fica branco/preto em vez da cor de acento do tema.

2. **`ColorNameForeground` rebaixado para cinza** (`#c8c8c8` escuro, `#333` claro)
   para que abas não selecionadas não se destaquem. Mas isso afeta TODO o texto
   da interface, deixando-o apagado e com baixo contraste.

3. **`ColorNameForegroundOnPrimary` definido como oposto de `ColorNamePrimary`**
   (preto no escuro, branco no claro) para o botão Save ficar legível. Mas isso
   ignora completamente a cor de acento do tema.

4. **`headerBg()` usa `lighten(bg, 0.20)` / `darken(bg, 0.12)`** — valores
   arbitrários, não baseados em teoria de cor ou contraste.

5. **`inputBg()` usa `lighten(bg, 0.03-0.06)`** — diferença insignificante,
   campos de input visualmente indistinguíveis do fundo.

6. **`hover()` usa `lighten(btn, 0.08)` / `blend(btn, fg, 0.10)`** — valores
   arbitrários, sem garantir contraste mínimo.

7. **`ColorNameHyperlink` = `t.fg`** — links com mesma cor do texto, impossível
   identificar que são clicáveis.

8. **`ColorNameFocus` e `ColorNameSelection` usam `t.fix` com alpha fixo**
   (`0x5c`, `0x40`) — valores arbitrários.

9. **Tema "Padrão do sistema" delega para `theme.DefaultTheme()`** mas a detecção
   de variante (`variantFor`) usa luminância do fundo escuro como fallback,
   podendo retornar o variant errado em alguns sistemas.

10. **Múltiplos valores mágicos espalhados** em `custom.go`:
    - `0xCC` (overlay alpha), `0.20` (input border blend), `0.38` (disabled blend),
      `0.42` (placeholder blend), `0.12` (pressed darken), `0x5c` (focus alpha),
      `0x40` (selection alpha), `0.45` (disabled button blend), `0.14` (scrollbar blend),
      `0.05` (scrollbar bg blend), `0.12` (separator blend), `0x10` (shadow alpha),
      `0xff` (foreground on error/success), `0x1a` (foreground on warning).

**NÃO usar testes automatizados.** Verificação visual, tema por tema:

1. Dracula
2. Solarized Dark
3. Monokai
4. Gruvbox Dark
5. Nord
6. One Dark
7. Tokyo Night
8. Catppuccin Mocha
9. Solarized Light
10. Gruvbox Light
11. Light Gray
12. Dark Gray
13. Padrão do sistema

Para cada tema:
- Fundo geral corresponde ao nome do tema.
- Texto legível (contraste com fundo).
- Campos de input visíveis e com bordas distinguíveis.
- Abas: selecionada destacada, não selecionadas recuadas.
- Links identificáveis como clicáveis.
- Botão Save sempre visível com texto legível.
- Nenhuma cor preta absoluta em tema claro, nenhuma cor branca em tema escuro.
- Ao trocar de tema, zero resíduos do tema anterior.

**NÃO usar testes automatizados.** Verificação visual, tema por tema:

1. Dracula
2. Solarized Dark
3. Monokai
4. Gruvbox Dark
5. Nord
6. One Dark
7. Tokyo Night
8. Catppuccin Mocha
9. Solarized Light
10. Gruvbox Light
11. Light Gray
12. Dark Gray
13. Padrão do sistema

Para cada tema:
- Fundo geral corresponde ao nome do tema.
- Texto legível (contraste com fundo).
- Campos de input visíveis.
- Nenhuma cor preta absoluta em tema claro, nenhuma cor branca em tema escuro.
- Ao trocar de tema, zero resíduos do tema anterior.

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
