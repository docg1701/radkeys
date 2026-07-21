# ponytail-audit — RadKeys

Auditoria de over-engineering do repositório inteiro. Apenas complexidade —
bugs de correção, segurança e performance estão fora de escopo.
Ordenado do maior corte para o menor. Nada foi aplicado.

## Achados

### 1. `delete` — código morto em `applySettings`
`u.navMap = nil` e, duas linhas depois, `if u.navMap != nil { u.navMap.SetTheme(...) }`.
O bloco é inalcançável. Cortar o `if` inteiro.
[internal/ui/ui.go, `applySettings`]

### 2. `shrink` — rebuild duplo do editor
`updateButtonsTab()` chama `buildButtonsTab()`, que já reconstrói grid,
inspector, layer bar e problems. Portanto `refresh()`, `setButtonAction`,
`setButtonContent`, `setButtonTarget` e `resizeGrid` podem descartar as
chamadas individuais `refreshGrid/refreshInspector/refreshLayerBar/
refreshProblems` antes de `updateButtonsTab()` — hoje tudo é construído
duas vezes por mutação.
[internal/editor/editor.go]

### 3. `delete` — doc comment duplicado em `press()`
As 6 primeiras linhas do bloco de comentário aparecem duplicadas
verbatim. Remover a primeira cópia.
[internal/ui/ui.go, `press`]

### 4. `shrink` — `themeOptions()` duplicado
Mesma função em ui e editor (itera `Presets` → ids + nomes localizados).
Criar `theme.Options()` junto de `Presets` e chamar dos dois lados.
[internal/ui/ui.go, internal/editor/appsettings.go]

### 5. `shrink` — `FirmwareOutdated`
Três `if`s aninhados viram um retorno booleano:
`return !known || major < MinFirmwareMajor || (major == MinFirmwareMajor && minor < MinFirmwareMinor)`
[internal/hid/hid.go]

### 6. `shrink` — `clamp` manual
Go 1.21+ tem builtins: `return min(max(v, lo), hi)`.
[internal/ui/map.go:167]

### 7. `shrink` — `satAdd` / `satSub`
One-liners com builtins:
`satAdd` → `uint8(min(uint16(a)+uint16(b), 255))`
`satSub` → `a - min(a, b)`
[internal/theme/theme.go]

### 8. `yagni` — wrapper de uma linha
`NewCustomTheme` só delega ao privado `newTheme`, seu único chamador.
Fundir numa única função exportada.
[internal/theme/theme.go]

### 9. `shrink` — `validate()` com loop que retorna na 1ª iteração
`for _, issue := range c.Issues() { return ... }` →
`if iss := c.Issues(); len(iss) > 0 { return iss[0].Error(...) }`
[internal/config/config.go]

## Pulados por design

- Mapa i18n de 803 linhas — single map é mandatório (AGENTS.md).
- Tabela `issueFormatters` — mesmo tamanho do switch que substituiu.
- Thunks `arg` em `deviceCommands` — modificador de OS é resolvido em
  tempo de chamada de propósito.

## Resultado

net: ~-45 linhas, -0 deps possíveis.
