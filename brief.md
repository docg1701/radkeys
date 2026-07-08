# RadKeys — Brief Técnico

> **Data:** 2026-07-08
> **Repo:** https://github.com/docg1701/radkeys
> **Release atual:** v0.2.1

---

## 1. Estado atual (v0.2.1)

| Componente | Status |
|------------|--------|
| Parser TOML + validação | ✅ |
| Navegação (deck) | ✅ |
| HID reader | ✅ |
| UI — aba Atalhos | ✅ |
| UI — aba Settings | ✅ Grid 3 colunas, 3 seções |
| UI — aba About | ✅ |
| i18n (7 idiomas) | ✅ Mapa Go único |
| 13 temas | ❌ Ver 2.3 — jogar fora e refazer |
| Ícone do app | ✅ Seletor Browse |
| Paste button | ✅ |
| CI + Release Linux/Windows | ✅ |
| macOS | ❌ Build nativo necessário |
| Always-on-top | ⏳ Fyne v2.8.0 |

## 2. Pendências

### 2.1 macOS
Build nativo num Mac.

### 2.2 Always-on-top
Pendente Fyne v2.8.0.

### 2.3 🔴 JOGAR FORA E REFAZER — Sistema de temas

**`internal/theme/custom.go` não tem conserto. Deve ser apagado e refeito do zero.**

A premissa está errada: derivar 28 cores de 3 hexes com `lighten`/`darken`/`blend`
e fatores mágicos (0.20, 0.38, 0.42, 0xCC, etc.) não funciona.

**Regras para o novo sistema:**

1. Apagar `custom.go`. Começar limpo.
2. Estudar temas Fyne reais (Catppuccin para Fyne, FyneLabs/notes).
3. Cada `ThemeColorName` com valor EXPLÍCITO por preset. Se um preset não define
   `ColorNameHeaderBackground`, usar `theme.DefaultTheme().Color(name, variant)`.
   **NUNCA** `lighten(bg, X)`.
4. Usar o parâmetro `variant` do método `Color()`. O Fyne já detecta light/dark
   do SO. Não reinventar detecção com luminância.
5. Para consultas fora do ciclo de renderização, usar
   `fyne.CurrentApp().Settings().ThemeVariant()`.
6. Presets precisam de mais de 3 cores. Mínimo: bg, fg, primary, button,
   input bg, header bg, hover, selection. O resto cai no fallback do DefaultTheme.
7. Temas claros → fallback `DefaultTheme.Color(name, VariantLight)`.
   Temas escuros → fallback `DefaultTheme.Color(name, VariantDark)`.
8. **Proibido:** `lighten`, `darken`, `blend`, `setAlpha` com fatores mágicos.
   Se uma cor não está no preset, delega ao DefaultTheme.
9. Manter IDs de preset (i18n) e hash de cores por preset.
10. **`dist/radkeys.config.toml`** deve ser sincronizado com `radkeys.config.toml`
    a cada build, ou manter apenas UM arquivo de config.

Verificação: abrir o app com cada um dos 13 temas e validar visualmente.
**Proibido testes automatizados.**

## 3. Estrutura

```
radkeys/
├── main.go / go.mod / go.sum
├── radkeys.config.toml
├── dist/                        # gitignored
├── internal/
│   ├── config/    config.go
│   ├── deck/      deck.go
│   ├── hid/       hid.go
│   ├── ui/        ui.go
│   ├── i18n/      i18n.go       # mapa Go, sem JSON
│   ├── theme/     custom.go     # ❌ JOGAR FORA
│   │              presets.go    # IDs de preset
│   └── assets/    assets.go
├── firmware/
└── research/
```

## 4. Regras de versão

- Única fonte: `[app] version` em `radkeys.config.toml`.
- Nunca hardcodar versão em Go.
- `ensureConfig` não inclui `version`.
- Test fixtures usam `"0.0.0-test"`.
