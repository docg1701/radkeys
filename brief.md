# RadKeys — Brief Técnico Completo

> **Versão:** 2.2  
> **Data:** 2026-07-08  
> **Repo:** https://github.com/docg1701/radkeys  
> **Release atual:** v0.2.0 (Linux + Windows + config)

---

## 1. Estado atual (v0.2.0)

| Componente | Status |
|------------|--------|
| Parser TOML + validação | ✅ `internal/config/` com 6 testes |
| Navegação (deck) | ✅ `internal/deck/` com 8 testes |
| HID reader (Elgato + DIY) | ✅ `internal/hid/` com build tags cgo/!cgo |
| HID mock (dev sem hardware) | ✅ 4 testes |
| UI — aba Atalhos (preview + keypad) | ✅ Preview topo 50%, keypad 4×5 embaixo 50% |
| UI — aba Ajustes | ❌ Layout quebrado (ver bugs abaixo) |
| i18n (7 idiomas) | ✅ go-i18n embed: en, pt-BR, pt-PT, es, fr, de, it |
| 12 temas de cores | ✅ `internal/theme/presets.go` |
| Ícone Obsidian | ✅ `internal/assets/` (preferences-desktop-keyboard-shortcuts) |
| Firmware Arduino | ✅ `firmware/arduino/diy24.ino` (matriz 6×4) |
| Firmware RP2040 | ✅ `firmware/rp2040/diy24.ino` (24 GPIO) |
| CI | ✅ Linux-only, test 40s + release com binário Linux |
| Release Windows | ✅ Cross-compile local (mingw), upload manual |
| Release macOS | ❌ Cross-compile impossível (SDK Apple). Build nativo num Mac. |
| Always-on-top | ⏳ Pendente Fyne v2.8.0 estável |
| AGENTS.md | ✅ Dev cycle + responsabilidades + checklist |
| README.md | ✅ Dependências, build, cross-compile |

## 2. Bugs — aba Ajustes (NÃO RESOLVIDOS)

Estes bugs foram documentados na v2.1 e **continuam pendentes**. O agente atual
(Nonatinho) tentou consertar com `widget.NewForm` mas o layout continua inaceitável.

| # | Bug | Status |
|---|-----|--------|
| 1 | Radiologista não atualiza título da janela | ✅ Corrigido (v0.2.0) |
| 2 | Campo "Nome do app" removido | ✅ Corrigido |
| 3 | Salvar aplica mudanças (título, idioma, tema, keypad) | ✅ Corrigido |
| 4 | Arquivo de config: agora tem seletor "Procurar..." | ✅ Adicionado, mas layout quebrado |
| 5 | Frase "Telas e botões..." removida | ✅ Corrigido |
| 6 | Cores individuais removidas (só seletor de tema) | ✅ Corrigido |
| 7 | Dispositivo USB: VID/PID/protocolo numa linha amontoada | ❌ Continua |
| 8 | Botão Save | ✅ Corrigido (usa `widget.NewForm.OnSubmit`) |
| 9 | Layout geral da aba Ajustes | ❌ Continua horrível |

## 3. Novos bugs específicos do layout da aba Ajustes (v2.2)

1. **Formulário estilo 1990**: labels alinhados à esquerda com inputs/selects
   gigantes até a margem direita. Uma linha embaixo da outra. Visual ultrapassado.
2. **Arquivo de config + botão Procurar**: desalinhados em relação ao resto do
   formulário (o `widget.NewForm` não foi feito para ter botões inline).
3. **Dispositivo USB**: VID, PID e dropdown de protocolo amontoados numa única
   linha, praticamente ilegível numa janela estreita.
4. **Layout não responsivo**: ao redimensionar a janela, os elementos não se
   reorganizam adequadamente.

## 4. Requisitos para o layout da aba Ajustes

- Layout moderno, limpo, profissional. Nada de formulário HTML dos anos 90.
- Grupos visuais claros: Radiologista, Idioma/Tema, Layout do Keypad,
  Arquivo de Configuração, Dispositivo USB.
- Campos com tamanhos adequados (não gigantescos).
- Responsivo: adapta-se ao redimensionamento da janela.
- Botão Salvar discreto, não full-width.
- Arquivo de config: label do caminho + botão "Procurar..." na mesma linha,
  alinhados corretamente.
- Dispositivo USB: VID e PID em campos pequenos lado a lado, protocolo abaixo
  ou ao lado com label claro.

## 5. Release

- **1 executável + 1 arquivo de configuração** para funcionar.
- CI: Linux-only (test + build binário Linux + release).
- Windows: cross-compilado localmente com mingw e uploaded manualmente.
- macOS: build nativo num Mac (cross-compile impossível com CGO).

## 6. Estrutura do repo

```
radkeys/
├── AGENTS.md / README.md / LICENSE / brief.md
├── main.go / go.mod / go.sum
├── radkeys.config.toml
├── dist/                        # Binários de release (gitignored)
├── internal/
│   ├── config/    config.go     # Parser TOML + tipos + validação
│   ├── deck/      deck.go       # Estado de navegação
│   ├── hid/       hid.go        # Interface + mock + go-hid real
│   ├── ui/        ui.go         # Fyne UI (Atalhos + Ajustes)
│   ├── i18n/      i18n.go       # go-i18n + 7 JSON embed
│   ├── theme/     presets.go    # 12 temas
│   └── assets/    assets.go     # Ícone Obsidian
├── firmware/arduino/            # Arduino Pro Micro
├── firmware/rp2040/             # RP2040
└── research/                    # Notas técnicas
```

## 7. Pendências para o próximo agente

### 7.1 Análise e reformulação completa da aba Settings

A aba "Ajustes" atual usa `widget.NewForm` mas o resultado é um formulário
estilo anos 90: labels à esquerda, inputs gigantes até a margem direita, uma
linha embaixo da outra. O layout precisa ser completamente repensado.

Requisitos:
- Layout moderno, limpo, profissional.
- Grupos visuais com cards ou seções colapsáveis.
- Campos com tamanhos adequados (não gigantescos).
- Responsivo: adapta-se ao redimensionamento.
- Dispositivo USB: VID e PID lado a lado em campos pequenos, protocolo
  com label claro, tudo bem diagramado.
- Arquivo de config: label do caminho + botão "Procurar..." alinhados.
- Salvar: botão discreto, integrado ao layout.

### 7.2 Aba "About"

Adicionar uma terceira aba "About" (ou "Sobre") com os dados fundamentais
do projeto, como todo bom aplicativo open source:
- Nome e versão do app.
- Breve descrição (1-2 linhas).
- Licença (MIT) com link.
- Repositório (github.com/docg1701/radkeys).
- Créditos: autor (Nonatinho/Galvani), ícone (Obsidian icon theme).
- Stack: Go, Fyne, go-hid, go-i18n, BurntSushi/toml.
- i18n disponível em 7 idiomas.

### 7.3 Reformulação completa dos temas

Os 12 temas atuais são aplicados apenas ao fundo do preview e aos botões,
mas o tema global do Fyne (`DarkTheme`) sobrescreve o restante da interface.
Resultado: mesmo os temas "Light" e "Solarized Light" ficam escuros.

Requisitos:
- Implementar um `fyne.Theme` customizado que aplique o preset a TODA a
  interface (fundo da janela, texto, botões, tabs, scrollbars).
- Temas claros devem usar o tema claro do Fyne como base; temas escuros
  usam o tema escuro.
- Respeitar o design original de cada tema (Solarized, Gruvbox, Nord,
  Dracula, Monokai, One Dark, Tokyo Night, Catppuccin).
- Incluir 1 tema claro padrão (ex.: Light Gray) e 1 tema escuro padrão
  (ex.: Dark Gray) que funcionem corretamente.