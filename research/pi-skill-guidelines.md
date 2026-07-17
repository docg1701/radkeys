# Pi Skill Creation Guidelines (July 2026)

> Compiled from the official Pi docs, Agent Skills specification, and the pi monorepo.
> Last updated: 2026-07-17

---

## What is a Skill

A skill is a self-contained capability package that the agent loads on-demand. It provides
specialized workflows, setup instructions, helper scripts, and reference documentation for
specific tasks. Pi implements the [Agent Skills standard](https://agentskills.io/specification).

## Directory Structure

```
my-skill/
├── SKILL.md              # Required: frontmatter + instructions
├── scripts/              # Executable code (bash, python, js, etc.)
├── references/           # Detailed docs loaded on-demand
│   └── api-reference.md
└── assets/               # Templates, images, static data
```

Only `SKILL.md` is required. Everything else is freeform.

## SKILL.md Format

```markdown
---
name: my-skill
description: What this skill does and when to use it. Be specific.
---

# My Skill

## Setup

```bash
cd /path/to/skill && npm install
```

## Usage

```bash
./scripts/process.sh <input>
```
```

Use relative paths from the skill directory.

## Frontmatter Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | **Yes** | Max 64 chars. Lowercase a-z, 0-9, hyphens only. |
| `description` | **Yes** | Max 1024 chars. What it does and when to use it. |
| `license` | No | License name or reference to bundled file. |
| `compatibility` | No | Max 500 chars. Environment requirements. |
| `metadata` | No | Arbitrary key-value mapping. |
| `allowed-tools` | No | Space-delimited pre-approved tools (experimental). |
| `disable-model-invocation` | No | When `true`, hides from system prompt; only `/skill:name`. |

### Name Rules

- 1–64 characters
- **Only** lowercase letters (`a-z`), digits (`0-9`), and hyphens (`-`)
- No leading or trailing hyphens
- No consecutive hyphens (`--`)
- **Pi does NOT require** the name to match the parent directory (unlike the standard)

✅ Valid: `pdf-processing`, `data-analysis`, `code-review`
❌ Invalid: `PDF-Processing`, `-pdf`, `pdf--processing`

### Description Best Practices

The description is what makes the agent decide to load the skill. Be specific.

✅ Good:
```yaml
description: Extracts text and tables from PDF files, fills PDF forms, and merges multiple PDFs. Use when working with PDF documents.
```

❌ Poor:
```yaml
description: Helps with PDFs.
```

## Skill Locations

Pi loads skills from:

- **Global:**
  - `~/.pi/agent/skills/`
  - `~/.agents/skills/`
- **Project** (only after the project is trusted):
  - `.pi/skills/`
  - `.agents/skills/` in `cwd` and ancestor directories (up to git repo root)
- **npm packages:** `skills/` directory or `pi.skills` entry in `package.json`
- **Settings:** `skills` array in `settings.json`
- **CLI:** `--skill <path>` (repeatable, additive even with `--no-skills`)

### Discovery Rules

- In `~/.pi/agent/skills/` and `.pi/skills/`, root `.md` files are discovered as individual skills
- In **all** locations, directories containing `SKILL.md` are discovered recursively
- In `~/.agents/skills/` and project `.agents/skills/`, root `.md` files are **ignored**

Disable discovery with `--no-skills` (explicit `--skill` paths still load).

## How Skills Work (Progressive Disclosure)

1. At startup, pi scans skill locations, extracts `name` + `description`
2. The system prompt includes available skills in XML format
3. When a task matches, the agent uses `read` to load the full `SKILL.md`
4. The agent follows the instructions using relative paths

Only descriptions are always in context. Full instructions load on-demand.

## Skill Commands

Every skill registers as a `/skill:name` command:

```bash
/skill:brave-search           # Load and execute
/skill:pdf-tools extract      # Load with arguments
```

Arguments after the name are appended as `User: <args>`.

Enable in `settings.json`:
```json
{
  "enableSkillCommands": true
}
```

## Validation

Pi validates against the Agent Skills standard. Most issues warn but still load:

- Name >64 chars or invalid characters → warning
- Name starts/ends with hyphen or has `--` → warning
- Description >1024 chars → warning
- **Missing description** → **skill does NOT load** (only hard error)
- Name collision → warning, keeps first found
- Unknown frontmatter fields → silently ignored

## Dynamic Skills from Extensions

Extensions can provide skills programmatically:

```typescript
import type { ExtensionAPI, Skill } from "@mariozechner/pi-coding-agent";

export default function (pi: ExtensionAPI) {
  const skill: Skill = {
    name: "dynamic-skill",
    description: "Skill provided by extension",
    content: "# Dynamic Skill\n\nInstructions here.",
    source: "extension",
  };

  pi.on("resources_discover", async (event) => {
    return { skills: [skill] };
  });
}
```

## Using Skills from Other Harnesses

Add their directories to `settings.json`:

Global (`~/.pi/agent/settings.json`):
```json
{
  "skills": ["~/.claude/skills", "~/.codex/skills"]
}
```

Project (`.pi/settings.json`):
```json
{
  "skills": ["../.claude/skills"]
}
```

## Skill Repositories

- [Anthropic Skills](https://github.com/anthropics/skills) — docx, pdf, pptx, xlsx, web dev
- [Pi Skills](https://github.com/badlogic/pi-skills) — web search, browser automation, Google APIs, transcription

---

## References

- [Pi Skills Docs](https://pi.dev/docs/latest/skills)
- [Creating Agent Skills](https://badlogic-pi-mono.mintlify.app/guides/creating-skills)
- [Agent Skills Specification](https://agentskills.io/specification)
- [Pi monorepo — skills.md](https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/skills.md)
