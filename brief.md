# RadKeys — Brief Técnico

> **Data:** 2026-07-12
> **Repo:** https://github.com/docg1701/radkeys
> **Release atual:** v0.2.1 → **meta: v0.3.0 (refatoração arquitetural)**
> **Status:** 🚧 Em andamento

---

## Progresso da refatoração (v0.3.0)

| Tarefa | Status |
|--------|--------|
| config.go — novo modelo (Layers, ações, sem FixedButtons) | ✅ |
| config_test.go — testes atualizados | ✅ |
| hid.go — Event com (Row, Col) | ✅ |
| reader_cgo.go — protocolo DIY (row,col) 2 bytes | ✅ |
| reader_nocgo.go — sem alterações necessárias | ✅ |
| ui.go — reescrever sem deck, layerIndex, novas ações | ✅ |
| deck/ — deletar | ✅ |
| main.go — atualizar ensureConfig | ✅ |
| radkeys.config.toml — novo formato | ✅ |
| hid_test.go — atualizar com (Row, Col) | ✅ |
| theme/custom.go — refazer do zero | ✅ |
| theme/presets.go — campos com cores explícitas por preset | ✅ |

---

## 1. Decisões de arquitetura (já definidas)

Estas 3 decisões vêm do protótipo inicial (`novidades-radkeys.md`) e substituem
o design atual. São superiores e **devem ser implementadas como descritas**:

### 1.1 Protocolo firmware: `(row, col)`, NÃO bitmap fixo

O firmware envia **2 bytes** por evento: `[row, col]`. Nunca um bitmap de tamanho
fixo. Isso torna o grid configurável sem recompilar firmware.

```
Firmware → Host: [row: uint8, col: uint8] (2 bytes)
Exemplo: botão na linha 3, coluna 5 → [0x03, 0x05]
```

Vantagem: o usuário define `columns × rows` no config.toml e o app mapeia
`(row, col) → index` automaticamente. Mesmo firmware funciona com 4×4, 6×6, 3×2...

### 1.2 Navegação hierárquica com `navigate` + `target`

O modelo é **screens com IDs** conectadas por `navigate` com `target`.
`prev` (stack-based) volta para a tela anterior. `home` vai para a raiz.

```
[[layers]]
name = "RX Tórax"
[[layers.buttons]]
row = 5; col = 5; label = "Próx"; action = "next"
[[layers.buttons]]
row = 5; col = 0; label = "Voltar"; action = "prev"
```

Vantagem: reordenar categorias no config não quebra referências. O usuário
edita a ordem das camadas e pronto, sem caçar targets quebrados.

### 1.3 Botões de sistema são ações normais, NÃO índices fixos

Nada de `[app.fixed_buttons]` reservando índices 0-2 do hardware.
`copy`, `paste`, `prev`, `next`, `home` são ações como qualquer outra —
o usuário as coloca na posição que quiser do grid.

```toml
[[layers.buttons]]
row = 5; col = 3; label = "Copy"; action = "copy"
[[layers.buttons]]
row = 5; col = 4; label = "Paste"; action = "paste"
```

Vantagem: layout 100% livre. Se o radiologista quiser 36 botões de texto e
zero de copy, pode. Se quiser 4 botões de copy em posições diferentes, pode.

---

## 2. O que se mantém do projeto atual

| Componente | Status |
|------------|--------|
| TOML (config) | ✅ Manter — parser + validação |
| i18n (7 idiomas) | ✅ Manter — mapa Go único |
| Preview de texto | ✅ Manter |
| UI — aba Settings (grid 3 colunas) | ✅ Manter |
| UI — aba About | ✅ Manter |
| Ícone do app (seletor Browse) | ✅ Manter |
| CI + Release Linux/Windows | ✅ Manter |
| HID reader (go-hid) | ✅ Manter — adaptar protocolo |
| Always-on-top | ⏳ Pendente Fyne v2.8.0 |
| macOS | ❌ Não entregue — só instruções de build |

---

## 3. O que é JOGADO FORA e refeito

### 3.1 `internal/deck/` — state machine com stack + grafo

**Substituir por:** um inteiro `layerIndex` e funções `prevLayer()`/`nextLayer()`.
Zero dependência de IDs, zero grafos, zero stacks.

### 3.2 `internal/theme/custom.go` — sistema de temas quebrado

**Refazer do zero** conforme regras da seção 5. Manter `presets.go` (IDs de preset).

### 3.3 Modelo de navegação `navigate` + `target`

**Substituir por:** ações `next` e `prev` (sequencial) + `home` (vai pra camada 0).

### 3.4 `FixedButtons` e índices reservados

**Remover completamente.** Copy/paste/prev/next/home viram ações como `text`.

---

## 4. Novo modelo de dados (TOML)

### 4.1 Config completo de exemplo

```toml
[app]
name = "RadKeys"
radiologist = "Dr. Galvani"
language = "pt"
version = "0.3.0"

[app.device]
vendor_id = 0x1234
product_id = 0xABCD
protocol = "radkeys-diy"

[app.layout]
columns = 6   # grid físico: 6 colunas
rows = 6      # grid físico: 6 linhas (até 36 botões)

[app.theme]
preset = "dark_blue"

# ── Camadas (ordenadas — a ordem define prev/next) ──

[[layers]]
name = "RX Tórax"

[[layers.buttons]]
row = 0; col = 3; label = "Normal"
action = "text"
content = """
Radiografia de tórax em incidências PA e perfil, realizadas em aparelho digital.
Arcada costal intacta, campos pulmonares livres, seios costofrênicos agudos.
Não há evidência de derrame pleural, pneumotórax ou consolidação.
"""

[[layers.buttons]]
row = 0; col = 4; label = "Derrame D"
action = "text"
content = """
Radiografia de tórax demonstrando opacidade basal direita com obliteração
do seio costofrênico, sugestiva de derrame pleural.
Acompanhamento/ultrassonografia de tórax recomendado.
"""

[[layers.buttons]]
row = 0; col = 5; label = "Pneumotórax"
action = "text"
content = """
Radiografia de tórax evidenciando pneumotórax à direita, com retração pulmonar.
Avaliação clínica conjunta recomendada.
"""

[[layers.buttons]]
row = 5; col = 0; label = "Voltar"; action = "prev"
[[layers.buttons]]
row = 5; col = 3; label = "Copy";  action = "copy"
[[layers.buttons]]
row = 5; col = 4; label = "Paste"; action = "paste"
[[layers.buttons]]
row = 5; col = 5; label = "Próx";  action = "next"

[[layers]]
name = "RX Abdome"

[[layers.buttons]]
row = 0; col = 3; label = "Normal"
action = "text"
content = """
Radiografia simples de abdome em decúbito e em pé.
Distribuição normal de conteúdo aéreo. Não se observam níveis hidroaéreos
nem calcificações patológicas.
"""

[[layers.buttons]]
row = 5; col = 0; label = "Voltar"; action = "prev"
[[layers.buttons]]
row = 5; col = 3; label = "Copy";  action = "copy"
[[layers.buttons]]
row = 5; col = 4; label = "Paste"; action = "paste"
[[layers.buttons]]
row = 5; col = 5; label = "Próx";  action = "next"

[[layers]]
name = "TC"

[[layers.buttons]]
row = 5; col = 0; label = "Voltar"; action = "prev"
[[layers.buttons]]
row = 5; col = 5; label = "Próx";  action = "next"

[[layers]]
name = "RM"

[[layers.buttons]]
row = 5; col = 0; label = "Voltar"; action = "prev"
[[layers.buttons]]
row = 5; col = 5; label = "Próx";  action = "next"
```

### 4.2 Tipos Go (novo `config.go`)

```go
type Config struct {
    App    App     `toml:"app"`
    Layers []Layer `toml:"layers"`
}

type App struct {
    Name        string  `toml:"name"`
    Radiologist string  `toml:"radiologist"`
    Language    string  `toml:"language"`
    Version     string  `toml:"version"`
    Device      Device  `toml:"device"`
    Layout      Layout  `toml:"layout"`
    Theme       Theme   `toml:"theme"`
}

type Device struct {
    VendorID  uint16 `toml:"vendor_id"`
    ProductID uint16 `toml:"product_id"`
    Protocol  string `toml:"protocol"` // "radkeys-diy" ou "elgato"
}

type Layout struct {
    Columns int `toml:"columns"` // grid físico (máx 6)
    Rows    int `toml:"rows"`    // grid físico (máx 6)
}

type Layer struct {
    Name    string   `toml:"name"`
    Buttons []Button `toml:"buttons"`
}

type Button struct {
    Row     int    `toml:"row"`            // 0-based
    Col     int    `toml:"col"`            // 0-based
    Label   string `toml:"label"`          // texto no botão da UI
    Action  string `toml:"action"`         // text | copy | paste | prev | next | home
    Content string `toml:"content,omitempty"` // só quando action = "text"
}
```

### 4.3 Ações suportadas

| Ação | Efeito |
|------|--------|
| `text` | Define `content` como texto atual e mostra no preview |
| `copy` | Copia o texto atual para o clipboard |
| `paste` | Cola o conteúdo do clipboard no preview (texto atual) |
| `next` | Vai para a próxima camada (circular: última → primeira) |
| `prev` | Vai para a camada anterior (circular: primeira → última) |
| `home` | Vai para a primeira camada |

### 4.4 Validação (mínimo)

- `columns` e `rows`: 0 < valor ≤ 6
- `row` e `col` de cada botão: 0 ≤ row < rows, 0 ≤ col < columns
- `action` deve ser um dos 6 valores da tabela acima
- `action = "text"` exige `content` não-vazio
- `action ≠ "text"` não deve ter `content`
- Pelo menos 1 camada
- Cada camada precisa de `name` não-vazio
- `language` deve ser um código suportado (pt, en, es, fr, de, it, ja)

---

## 5. Sistema de temas — refazer do zero

### 5.1 Apagar e começar limpo

`internal/theme/custom.go` **não tem conserto**. Deve ser apagado por inteiro.

### 5.2 Regras

1. Cada `ThemeColorName` com valor EXPLÍCITO por preset. Se um preset não define
   `ColorNameHeaderBackground`, usar `theme.DefaultTheme().Color(name, variant)`.
   **NUNCA** `lighten(bg, X)`.
2. Usar o parâmetro `variant` do método `Color()`. O Fyne já detecta light/dark
   do SO. Não reinventar detecção com luminância.
3. Para consultas fora do ciclo de renderização, usar
   `fyne.CurrentApp().Settings().ThemeVariant()`.
4. Presets precisam de no mínimo: bg, fg, primary, button, input bg, header bg,
   hover, selection. O resto cai no fallback do DefaultTheme.
5. Temas claros → fallback `DefaultTheme.Color(name, VariantLight)`.
   Temas escuros → fallback `DefaultTheme.Color(name, VariantDark)`.
6. **Proibido:** `lighten`, `darken`, `blend`, `setAlpha` com fatores mágicos.
   Se uma cor não está no preset, delega ao DefaultTheme.
7. Manter IDs de preset (i18n) e hash de cores por preset.

Verificação: abrir o app com cada um dos 13 temas e validar visualmente.
**Proibido testes automatizados.**

---

## 6. Limpeza de dívida técnica (aproveitar do código atual)

### 6.1 `internal/ui/ui.go`

1. Dead code em `buildAbout()`: variáveis `author` e `license` criadas com
   Wrapping mas nunca usadas — `repoLine` usa `i18n.T()` direto.
2. `renderScreen()` recria todos os botões (20+ closures novas) a cada chamada.
3. Save recria `buildSettings()` e `buildAbout()` inteiros — devia usar Refresh.
4. `resolveFullTheme`: fallback silencioso para Presets[0] se FindPreset falha.
5. `appIconData`: suprime erros de leitura, sem aviso ao usuário.
6. `configLbl.Wrapping = fyne.TextTruncate` — path truncado.

### 6.2 `internal/config/config.go` — ao reescrever, evitar

1. Campos mortos (Background, Button, Fixed — hoje não usados).
2. Campo `Preview` reservado — ou implementa ou remove.
3. `validate()` deve validar VID/PID, preset, language, rows, cols, row/col bounds.

---

## 7. Firmware — RP2040-Zero

### 7.1 Protocolo

```
Envia 2 bytes por evento: [row: uint8, col: uint8]
Report HID vendor-defined via TinyUSB.
```

### 7.2 Suporte a até 36 botões (6×6)

O firmware varre a matriz e envia `(row, col)` quando detecta pressionamento.
**Zero hardcode de tamanho de grid** — o firmware só varre os pinos que existem.
O app recebe `(row, col)` e mapeia para o grid definido no config.

### 7.3 Pinagem (6×6 = 12 GPIOs)

```
Linhas (output, LOW ativa): GP0, GP1, GP2, GP3, GP4, GP5
Colunas (input, pull-up):   GP6, GP7, GP8, GP9, GP10, GP11
```

GPIOs GP12–GP22 livres para expansão.

### 7.4 VID/PID

Placeholders `0x1234`/`0xABCD` — o usuário define os seus no config.toml e
no firmware. O app lê do config.

### 7.5 Permissão Linux (udev)

```
KERNEL=="hidraw*", SUBSYSTEM=="hidraw", ATTRS{idVendor}=="1234", ATTRS{idProduct}=="abcd", MODE="0660", GROUP="input"
```

### 7.6 Cuidado com clones

Usar vendedor confiável (2000+ vendas, nota ≥4.8). Testar ao receber:
conectar no PC, gravar firmware, reiniciar o PC 3× e confirmar que o
dispositivo não some do barramento USB (bug de clones com chip B0 stepping).

---

## 8. Estrutura de diretórios (pós-refatoração)

```
radkeys/
├── main.go
├── go.mod / go.sum
├── radkeys.config.toml      # config de exemplo (versionado)
├── dist/                    # gitignored
├── internal/
│   ├── config/              # TOML parser + validação + tipos
│   ├── hid/                 # HID reader (go-hid + mock)
│   ├── ui/                  # Fyne UI: preview + grid + settings + about
│   ├── i18n/                # mapa Go único (7 idiomas)
│   ├── theme/               # presets.go + custom.go (novo, do zero)
│   └── assets/              # ícone embed
├── firmware/
│   └── rp2040-zero/         # RP2040-Zero: diy.ino (TinyUSB, protocolo row,col)
└── BUILD.md                 # guia de montagem do hardware
```

> `internal/deck/` foi **removido**. Navegação é `layerIndex int` no ui.go.

---

## 9. Plataformas

| Plataforma | Binário na release | Responsabilidade |
|------------|---------------------|------------------|
| Linux | ✅ Build e entrega | Prioridade — testado |
| Windows | ✅ Cross-compile com mingw | Fornecido, mas NÃO testado pelo autor |
| macOS | ❌ Não entregue | Instruções de build no README |

---

## 10. Regras de versão

- Única fonte: `[app] version` em `radkeys.config.toml`.
- Nunca hardcodar versão em Go.
- Test fixtures usam `"0.0.0-test"`.
