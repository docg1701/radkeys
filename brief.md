# RadKeys — Brief Técnico Completo

> **Versão:** 2.1
> **Data:** 2026-07-08
> **Autor:** Nonatinho (consultor)
> **Cliente:** Galvani (radiologista)
> **Repo:** https://github.com/docg1701/radkeys
>
> **v2.1 — registra 9 bugs na aba Ajustes para correção pelo próximo agente.**
> O agente atual (Nonatinho) não conseguiu corrigir a aba Ajustes a contento.
> Esta versão documenta os bugs para o próximo agente resolver.

---

## 1. Visão do produto

RadKeys é um aplicativo desktop portátil, open source, multiplataforma, para
médicos radiologistas que **digitam laudos**. Exibe uma interface com um
**preview de texto em cima** (metade da tela) e um **keypad virtual embaixo**
(metade da tela) que reflete o dispositivo USB de botões. Cada botão carrega
uma frase pronta de laudo; o usuário confirma a cópia, o texto vai para a
área de transferência e é colado no RIS/PACS.

A operação é feita por um **dispositivo USB HID custom**, lido diretamente
pelo RadKeys — o hardware **não envia teclas**, logo **não rouba o foco do RIS**.

### Modelo de botões

| Botões | Função | Configurável? |
|--------|--------|---------------|
| **3 fixos** (índices 0/1/2 por padrão) | `copy`, `level_up`, `go_home` | Fixa (global) |
| **N−3 configuráveis** | `navigate` (sub-tela) ou `text` (frase) | Por tela |

- **N** = botões do hardware. Alvo: 24 (DIY).
- **Hierarquia de telas**: `navigate` entra em sub-tela; `level_up`/`go_home` retornam.

---

## 2. Restrições e requisitos

| Item | Definição |
|------|-----------|
| Open source | MIT |
| Config | `radkeys.config.toml` (plaintext TOML) |
| Plataformas | Windows 10/11, macOS (Intel+AS), Linux |
| Distribuição | **1 executável + 1 config**. Tudo embed no binário (ícone, traduções, temas). |
| Hardware | HID custom: Stream Deck/clone (Elgato) ou DIY 24 (Arduino) |
| Botões | 3 fixos + (N−3) configuráveis; alvo 24 |
| i18n | 7 idiomas (en, pt-BR, pt-PT, es, fr, de, it) via go-i18n |
| Temas | 12 presets (10 de terminal + 2 cinza) |
| Ícone | Do tema Obsidian (preferences-desktop-keyboard-shortcuts), embarcado |

---

## 3. Stack tecnológico

| Camada | Tecnologia |
|--------|-----------|
| Linguagem | Go 1.22+ |
| GUI | Fyne v2.7.4 (estável) |
| HID | `github.com/sstallion/go-hid` v0.15.0 (hidapi, CGO) |
| i18n | `github.com/nicksnyder/go-i18n/v2` v2.6.1 |
| Config | `github.com/BurntSushi/toml` |
| Build | `fyne-cross` ou nativo por OS |

### Notas de API (validadas)

- **Always-on-top**: `desktop.Window.RequestAlwaysOnTop()` **não existe em v2.7.4**
  (PR #6184 mergeado em `develop`/v2.8.0, ainda rc1). MVP sem always-on-top;
  re-adicionar ao subir para v2.8.0 estável. Ver `research/fyne-always-on-top.md`.
- **Clipboard**: `a.Clipboard().SetContent(texto)` (disponível desde 2.6).
- **go-hid**: `hid.Init()` → `hid.OpenFirst(vid,pid)` → `d.ReadWithTimeout(buf, 50ms)` → `d.Close()`.

---

## 4. Arquitetura

```
┌──────────────────────────────────────┐
│            RadKeys (Fyne UI)          │
│  ┌────────────┐  ┌─────────────────┐  │
│  │ Aba        │  │ Aba "Ajustes"   │  │
│  │ "Atalhos"  │  │ (settings)     │  │
│  │            │  │                 │  │
│  │ ┌────────┐ │  │                 │  │
│  │ │Preview │ │  │                 │  │
│  │ │(topo,  │ │  │                 │  │
│  │ │ 50%)   │ │  │                 │  │
│  │ ├────────┤ │  │                 │  │
│  │ │Keypad  │ │  │                 │  │
│  │ │(baixo, │ │  │                 │  │
│  │ │ 50%)   │ │  │                 │  │
│  │ │ 4×N    │ │  │                 │  │
│  │ └────────┘ │  │                 │  │
│  └────────────┘  └─────────────────┘  │
└──────────────────────────────────────┘
```

---

## 5. Formato do config

```toml
[app]
name = "RadKeys"
radiologist = "Dr. Galvani"
language = "pt-BR"
version = "2.0"

[app.device]
vendor_id  = 0x0fd9
product_id = 0x0063
protocol   = "elgato"

[app.fixed_buttons]
copy     = 0
level_up = 1
go_home  = 2

[app.layout]
columns = 4
rows    = 5

[app.theme]
preset = "Dracula"

[[screens]]
id = "root"
title = "Início"
buttons = [
  { index = 3, label = "RX Tórax", action = "navigate", target = "rx_torax" },
  { index = 4, label = "Normal", action = "text", content = "..." },
]
```

---

## 6. Interface do usuário

### 6.1 Aba "Atalhos"

- **Título da janela**: `"RadKeys — <nome do radiologista>"` (NÃO hardcoded; usa `cfg.App.Radiologist`).
- **Topo (50%)**: preview — `widget.Label` monospace com scroll, preserva quebras de linha.
- **Baixo (50%)**: keypad virtual — grid `columns × rows` (default 4×5). Botões e slots
  vazios têm o **mesmo tamanho** (o grid garante; não usar `NewGridWrap` com tamanhos diferentes).
- **Sem título** acima do preview (removido em v2.0).
- Tema escuro. Janela redimensionável (1280×800 default).

### 6.2 Navegação

- `copy`: copia preview para clipboard.
- `level_up`: sobe um nível.
- `go_home`: volta à raiz.
- `navigate`: entra em sub-tela.
- `text`: carrega texto no preview.

### 6.3 Aba "Ajustes"

**DEVE CONTER APENAS:**
- Nome do radiologista (Entry; **deve atualizar o título da janela ao salvar**).
- Idioma (Select com 7 opções: en, pt-BR, pt-PT, es, fr, de, it).
- Tema (Select com 12 presets: Dracula, Solarized Dark, Monokai, Gruvbox Dark,
  Nord, One Dark, Tokyo Night, Catppuccin Mocha, Solarized Light, Gruvbox Light,
  Light Gray, Dark Gray).
- Layout: colunas e linhas (Entry numérico cada).
- Dispositivo: VID (Entry), PID (Entry), protocolo (Select: elgato / radkeys-diy).
- Botão Salvar (largura **normal**, não full-width).

**NÃO DEVE CONTER:**
- ❌ Campo "Nome do app" (remover — imbecil).
- ❌ Campo "Arquivo de configuração" (remover — inútil, não é editável).
- ❌ Frase "Telas e botões são editados manualmente..." (remover da UI).
- ❌ Campos individuais de cor (background/button/fixed) — só o preset de tema.
- ❌ Dropdown "elgato" com label cortado ("elgado") — corrigir label.

**AO SALVAR:**
- As mudanças **devem aplicar-se imediatamente à interface** (re-renderizar
  keypad, atualizar título, trocar tema, trocar idioma). O bug atual é que
  salvar não muda nada na UI.

---

## 7. BUGS REGISTRADOS (v2.1) — para o próximo agente corrigir

| # | Bug | Severidade |
|---|-----|------------|
| 1 | Mudar o nome do radiologista não muda a interface. Deveria compor o título da janela (`"RadKeys — <radiologista>"`). | Alta |
| 2 | Campo "Nome do app" nos Ajustes é inútil e imbecil. Remover. | Média |
| 3 | Salvar ajustes não aplica nada na interface (keypad, tema, idioma não atualizam). O `save()` chama `resolveTheme` e `renderScreen` mas não reconstrói o layout se columns/rows mudaram, não atualiza o título com o radiologista, e o tema pode não estar sendo aplicado corretamente. | Crítica |
| 4 | Arquivo de config aparece na tela de Ajustes mas não é editável. Remover. | Média |
| 5 | Frase "Telas e botões são editados manualmente neste arquivo (TOML)" na UI. Remover. | Baixa |
| 6 | Campos individuais de cor (background/button/fixed) na aba Ajustes. Remover — só o seletor de tema. | Média |
| 7 | Seção "Dispositivo USB": VID/PID em caixinhas minúsculas + dropdown "elgado" (label cortado). Corrigir layout e label. | Média |
| 8 | Botão Salvar ocupa a largura toda da tela. Deve ter largura normal. | Baixa |
| 9 | Layout geral da aba Ajustes é caótico. Reorganizar: agrupar por seção com labels claros, espaçamento adequado, sem elementos inúteis. | Alta |

---

## 8. Hardware USB

### Opções

1. **Comprar pronto**: Stream Deck/clone (protocolo Elgato). 15/32 teclas.
2. **DIY (primário)**: Arduino Pro Micro (ATmega32U4) + chaves de teclado chinês
   + caixa 3D + cabo USB. Matriz 6×4. ~R$30-50. Firmware: `firmware/arduino/`.
3. **DIY (alt)**: Raspberry Pi Pico (RP2040). `firmware/rp2040/`.

### Protocolos

- **Elgato**: input report `0x01/0x00`, 1 byte/botão. Feature `0x08` = rows×cols.
- **DIY**: vendor-defined, report ID 1 + 24 bytes. `parseDIYReport` aceita 25 ou 24 bytes.

### Permissões Linux

`/dev/hidraw*`. Grupo `input` ou regra udev por VID/PID.

---

## 9. i18n

- `internal/i18n/` com go-i18n v2.6.1 e 7 arquivos JSON embed.
- Idiomas: en (default), pt-BR, pt-PT, es, fr, de, it.
- `i18n.T(key)` para todas as strings de UI.
- `i18n.SetLanguage(lang)` troca o idioma em runtime.
- Para adicionar um idioma: criar `locales/<code>.json` e adicionar a `i18n.Supported`.

---

## 10. Temas

- `internal/theme/presets.go` com 12 presets + "Custom".
- 10 inspirados em temas de terminal: Dracula, Solarized Dark/Light, Monokai,
  Gruvbox Dark/Light, Nord, One Dark, Tokyo Night, Catppuccin Mocha.
- 2 de cinza: Light Gray, Dark Gray.
- Seletor na aba Ajustes. Ao selecionar, aplica as cores do preset.
- **Não mostrar campos individuais de cor** — só o seletor de preset.

---

## 11. Release

- **1 executável + 1 config**. Tudo embed: ícone (Obsidian), traduções (7 JSON),
  temas (12 presets). Nenhum arquivo externo além do config.
- CI: `.github/workflows/build.yml` — testes em 3 OS + auto-release de tag
  (binários linux/windows/macos + radkeys.config.toml como assets).
- Tag **lightweight**: `git tag vX.Y.Z <sha>` (NÃO anotada).
- Changelog categorizado por conventional commits.

---

## 12. Estado atual de desenvolvimento

| Componente | Status | Notas |
|------------|--------|-------|
| Parser TOML | ✅ | `internal/config/` com testes |
| Navegação (deck) | ✅ | `internal/deck/` com testes |
| HID reader (Elgato+DIY) | ✅ | `internal/hid/` com build tags cgo/!cgo |
| HID mock | ✅ | `internal/hid/hid.go` |
| UI — aba Atalhos | ✅ | Preview + keypad 4×5 |
| UI — aba Ajustes | ❌ **9 bugs** | Ver seção 7 |
| i18n (7 idiomas) | ✅ | `internal/i18n/` |
| Temas (12 presets) | ✅ | `internal/theme/` |
| Ícone (Obsidian) | ✅ | `internal/assets/` |
| Firmware Arduino | ✅ | `firmware/arduino/` |
| Firmware RP2040 | ✅ | `firmware/rp2040/` |
| CI | ✅ | `.github/workflows/build.yml` |
| AGENTS.md | ✅ | 6 áreas + ciclo dev→CI→release |
| README.md | ✅ | Best practices 2026 |
| Always-on-top | ⏳ | Pendente Fyne v2.8.0 estável |
| Cross-compile Windows/macOS | ⏳ | CI testa mas não validado localmente |

---

## 13. Próximos passos (para o próximo agente)

1. **Corrigir os 9 bugs da aba Ajustes** (seção 7). Prioridade: bug #3 (crítico).
2. Ao salvar Ajustes: reconstruir o keypad se columns/rows mudaram,
   atualizar título com radiologista, aplicar tema, aplicar idioma.
3. Remover: campo "Nome do app", campo "Arquivo de config", frase "Telas e
   botões...", campos individuais de cor.
4. Reorganizar a aba Ajustes com layout limpo e agrupado.
5. Botão Salvar com largura normal (não full-width).
6. Dispositivo: VID/PID com labels e tamanho adequados, dropdown com label correto.
7. Testar com hardware real (Stream Deck ou DIY 24).
8. Validar cross-compile Windows/macOS.

---

## 14. Fontes-chave

- `sstallion/go-hid`: https://pkg.go.dev/github.com/sstallion/go-hid
- libusb/hidapi: https://github.com/libusb/hidapi
- Elgato HID: https://docs.elgato.com/streamdeck/hid/general
- Fyne v2.7.4: https://pkg.go.dev/fyne.io/fyne/v2
- PR #6184 (always-on-top): https://github.com/fyne-io/fyne/pull/6184
- go-i18n: https://pkg.go.dev/github.com/nicksnyder/go-i18n/v2
- DIY Stream Deck (ref): https://github.com/Mercawa/DIYStreamDeck-HIDKeyboard

---

## 15. Estrutura do repo

```
radkeys/
├── AGENTS.md                      # Regras para agentes (6 áreas + ciclo release)
├── brief.md                       # Este brief (v2.1)
├── main.go                        # Entrypoint
├── radkeys.config.toml            # Config de exemplo (comentado para humano/LLM)
├── README.md / LICENSE
├── go.mod / go.sum
├── .github/workflows/build.yml    # CI: test + auto-release
├── internal/
│   ├── config/                    # Parser TOML + validação + tipos
│   ├── deck/                      # Estado de navegação
│   ├── hid/                       # Interface Reader + Mock + go-hid
│   ├── ui/                        # Fyne UI (Atalhos + Ajustes)
│   ├── i18n/                       # go-i18n + 7 JSON embed
│   ├── theme/                     # 12 preset themes
│   └── assets/                    # Ícone Obsidian embarcado
├── firmware/
│   ├── arduino/                   # Arduino Pro Micro (matriz 6×4)
│   └── rp2040/                    # RP2040 (24 GPIO)
└── research/
    └── fyne-always-on-top.md      # Investigação PR #6184
```

### Testes

```
go test ./...  →  config: ok (6)  deck: ok (8)  hid: ok (4)
go vet ./...   →  clean
go build .     →  radkeys (32 MB, CGO + Fyne + go-hid + i18n + icon embed)
```