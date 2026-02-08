# aiops

AI-powered operations infrastructure for your codebase. Installs senior-engineer behavior as code — rules, workflows, coordination, and evolution — for any AI-powered IDE.

**Supported targets:** Windsurf (Cascade), Cursor, Continue (VS Code), GitHub Copilot. Auto-detected — generates for all IDEs found on your system.

## What It Does

`aiops` scans your repository, detects your tech stack, and generates a complete set of AI session management artifacts:

- **Global rules** — behavioral constitution loaded into every AI session
- **Orchestrator** — cross-session coordination with advisory locks and build failure protocol
- **Workflows** — default execution mode, evolution audits, multiagency specs
- **Multiagency module** — complete Go module with CLI, spec parser, LLM clients, agent executor, and pipeline orchestrator
- **Intent guardrails** — prevents agents from drifting outside task scope
- **Escalation budget** — prevents over-cautious behavior
- **Human override** (`@directive`) — escape hatch that overrides process, not safety

## Installation

### Option 1: Install script (recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/voltic-software/aiops/main/install.sh | sh
```

### Option 2: Go install

```bash
go install github.com/voltic-software/aiops/cmd/aiops@latest
```

### Option 3: Build from source

```bash
git clone https://github.com/voltic-software/aiops.git
cd aiops
make install
```

## Quick Start

```bash
# Initialize in your project
cd your-project
aiops init

# After adding/removing MCP servers or IDEs
aiops sync

# Check for drift
aiops status
```

## Commands

### `aiops init`

Scans your repo, asks 1-2 questions, generates all artifacts.

```
$ aiops init

aiops init — scanning /Users/you/myproject

Scanning repository...

Detected:
  Languages: go (backend), typescript (frontend)
  Frameworks: temporal, gin, mqtt, nextjs, tailwindcss
  Build: go build ./..., npx tsc --noEmit, npm run build
  Patterns: domain-driven-design, event-sourcing, code-generation, containerized
  MCP servers: vera (windsurf), github (cursor), postgres (cursor)

Is this correct? [Y/n] y

Project name [myproject]: myproject

IDE targets: Windsurf (Cascade), Cursor, GitHub Copilot
Project maturity: bootstrap

Generating artifacts...
  ✓ ~/.codeium/windsurf/memories/global_rules.md  (global policy)
  ✓ .windsurf/rules/aiops.md                      (repo rules)
  ✓ .windsurf/workflows/default-mode.md
  ✓ .windsurf/workflows/orchestrator.md
  ✓ .windsurf/workflows/multiagency.md
  ✓ .windsurf/orchestrator/session_state.yaml
  ✓ .cursor/rules/aiops.mdc                       (repo rules)
  ✓ .cursor/prompts/default-mode.md
  ✓ .cursor/prompts/orchestrator.md
  ✓ .cursor/prompts/multiagency.md
  ✓ .cursor/orchestrator/session_state.yaml
  ✓ .github/copilot-instructions.md                (repo rules)
  ✓ multiagency/go.mod
  ✓ multiagency/cmd/multiagency/main.go
  ✓ multiagency/internal/...
  ✓ multiagency/specs/design.yaml
  ✓ multiagency/specs/code_review.yaml
  ✓ multiagency/specs/manager.yaml
  ✓ multiagency/specs/evolution_audit.yaml
  ✓ multiagency/specs/risks.yaml
  ✓ decisions/0001-aiops-initialized.md

✅ aiops initialized! 29 files generated.

🚀 Bootstrap mode detected — recommended first actions:
  1. Open an AI session and run: /multiagency design.yaml
  2. Produce architecture.md, risks.md, assumptions.md
  3. After architecture is framed, start building (single-agent)
  4. Run `aiops sync` after the project matures
```

### `aiops scan`

Read-only scan — shows what's detected without writing files.

```
$ aiops scan

Detected:
  Languages: go (backend), typescript (frontend)
  Frameworks: temporal, gin, mqtt, nextjs, tailwindcss
  Build: go build ./..., npx tsc --noEmit, npm run build
  Patterns: domain-driven-design, event-sourcing, containerized
```

### `aiops sync`

Re-scans MCP servers, IDE targets, and project maturity. Updates config and re-renders rules. No questions asked — designed to be fast and scriptable.

```
$ aiops sync

aiops sync — myproject

  + MCP added: postgres (cursor)
  ↑ Maturity changed: bootstrap → active

Re-rendering artifacts...

✅ Synced. 28 files updated.
```

Run this after adding/removing MCP servers, installing new IDEs, or when the project grows past bootstrap stage.

### `aiops status`

Shows what's installed and detects drift (new frameworks, languages, patterns, MCP servers added since last init).

```
$ aiops status

aiops status — myproject (v0.1.0)

Artifacts:
  ✓ Default mode workflow
  ✓ Orchestrator workflow
  ✓ Session state
  ✓ Global rules (memories)

4 installed, 0 missing

Re-scanning repository...
✓ Detected stack matches config — no drift detected
```

### `aiops update`

Regenerates artifacts from latest templates, shows what changed, applies with approval.

```
$ aiops update

aiops update — myproject

Computing diff against latest templates...

Update plan: 0 new, 2 modified, 7 unchanged

  ~ .windsurf/workflows/default-mode.md (modified)
  ~ .windsurf/workflows/orchestrator.md (modified)

Apply these changes? [y/n] y

✅ Updated 9 files.
```

### `aiops evolve`

Reads `@directive` override logs from the orchestrator and detects patterns that suggest rule changes.

```
$ aiops evolve

aiops evolve — analyzing directive logs for myproject

# Evolution Analysis Report

Total directives logged: 5
Patterns detected: 2

### Pattern 1: `escalation` overridden 3 times
Proposed: Raise escalation budget from 2 to 3 per session.

### Pattern 2: `intent_scope` overridden 2 times
Proposed: Relax intent guardrail for dependent-file changes.
```

### `aiops skills`

Auto-generates skill scaffolds based on detected frameworks. Skills are placed in each target's skills directory and auto-invoked based on task type.

```
$ aiops skills

aiops skills — generating skill scaffolds for myproject

Detected skills to generate:
  - @domain-changes — Guide for modifying domain entities
  - @mqtt-integration — Guide for MQTT message flows
  - @frontend-component — Guide for React/Next.js components
  - @code-review — Guide for code reviews

Generate these skill scaffolds? [Y/n] y

  ✓ .windsurf/skills/domain-changes/SKILL.md
  ✓ .windsurf/skills/mqtt-integration/SKILL.md
  ✓ .windsurf/skills/frontend-component/SKILL.md
  ✓ .windsurf/skills/code-review/SKILL.md

✅ Generated 4 skill scaffolds.
```

## Supported IDE Targets

| Target       | Rules                                          | Workflows              | Orchestrator              | Auto-detected by                      |
| ------------ | ---------------------------------------------- | ---------------------- | ------------------------- | ------------------------------------- |
| **Windsurf** | `~/.codeium/windsurf/memories/global_rules.md` | `.windsurf/workflows/` | `.windsurf/orchestrator/` | `~/.codeium/windsurf/` exists         |
| **Cursor**   | `.cursor/rules/aiops.mdc`                      | `.cursor/prompts/`     | `.cursor/orchestrator/`   | `.cursor/` or `~/.cursor/` exists     |
| **Continue** | `.continue/rules/aiops.md`                     | `.continue/prompts/`   | `.continue/orchestrator/` | `.continue/` or `~/.continue/` exists |
| **Copilot**  | `.github/copilot-instructions.md`              | —                      | —                         | `.github/` or `~/.vscode/` exists     |

All targets get the same rules content, adapted to the correct file paths. Templates reference `{{.OrchestrDir}}` so each target's rules point to its own orchestrator location.

## What Gets Generated

### `aiops init` — Per-target artifacts (repeated for each detected IDE)

| File                               | Purpose                                         |
| ---------------------------------- | ----------------------------------------------- |
| Rules file (path varies by target) | Compact behavioral rules — always active        |
| Workflows directory (path varies)  | Default mode, orchestrator, multiagency prompts |
| Orchestrator state (path varies)   | Shared state for parallel sessions              |

### `aiops init` — Multiagency Go module

A complete, compilable Go module generated with import paths derived from your detected `go.mod`.

| File                                        | Purpose                                           |
| ------------------------------------------- | ------------------------------------------------- |
| `multiagency/go.mod`                        | Go module (auto-derived from project module path) |
| `multiagency/README.md`                     | Usage guide and architecture docs                 |
| `multiagency/cmd/multiagency/main.go`       | CLI — validate, show, list, init commands         |
| `multiagency/internal/spec/types.go`        | Workflow spec types and validation                |
| `multiagency/internal/spec/loader.go`       | YAML spec parsing                                 |
| `multiagency/internal/llm/client.go`        | LLM client interface                              |
| `multiagency/internal/llm/stub.go`          | Stub client for testing                           |
| `multiagency/internal/llm/anthropic.go`     | Anthropic Claude client                           |
| `multiagency/internal/agent/executor.go`    | Agent execution with retry and validation         |
| `multiagency/internal/agent/prompt.go`      | System/user prompt builder                        |
| `multiagency/internal/pipeline/context.go`  | Pipeline execution state                          |
| `multiagency/internal/pipeline/executor.go` | Pipeline orchestrator                             |
| `multiagency/specs/design.yaml`             | Architecture design workflow (4 agents)           |
| `multiagency/specs/code_review.yaml`        | Code review workflow (4 agents)                   |
| `multiagency/specs/manager.yaml`            | Task classification workflow (2 agents)           |
| `multiagency/specs/evolution_audit.yaml`    | Knowledge freshness audit (2 agents)              |

### `aiops skills` (framework-specific, skip if already exists)

| Skill                 | Detected When                  |
| --------------------- | ------------------------------ |
| `@domain-changes`     | `domain-driven-design` pattern |
| `@mqtt-integration`   | `mqtt` framework               |
| `@frontend-component` | `nextjs` framework             |
| `@code-review`        | Always generated               |

## Architecture

```
aiops/
├── cmd/aiops/main.go               # CLI (init, scan, sync, status, update, evolve, skills)
├── internal/
│   ├── config/config.go            # .aiops.yaml schema and I/O
│   ├── scanner/scanner.go          # Repo analysis, Go module detection, maturity detection
│   ├── target/target.go            # IDE target definitions + auto-detection
│   ├── renderer/
│   │   ├── renderer.go             # Multi-target template rendering engine
│   │   └── templates/
│   │       ├── memories/           # → Global policy rules (Windsurf only)
│   │       ├── repo_rules.md.tmpl  # → Repo implementation rules (all targets)
│   │       ├── decisions/          # → Decisions memory scaffold
│   │       ├── windsurf/           # → Workflows + orchestrator (rendered per target)
│   │       └── multiagency/        # → Complete Go module (rendered once)
│   ├── updater/updater.go          # Diff and apply template updates
│   ├── evolve/evolve.go            # Directive log analysis and rule proposals
│   └── skills/skills.go            # Framework-specific skill scaffold generation
└── README.md
```

**Key design decisions:**

- **Policy/implementation split** — global rules define policy (constitutional law); repo rules define implementation (statutes). Repo rules win on conflict.
- **Target abstraction** — each IDE is a `Target` with path mappings for global rules, repo rules, workflows, orchestrator, and skills
- **Auto-detection** — scans for IDE config directories (`~/.codeium/windsurf/`, `.cursor/`, etc.)
- **Render per target** — rules and workflows are rendered once per detected target with `{{.OrchestrDir}}` adapted
- **Shared artifacts** — multiagency module and decisions directory are rendered once (IDE-independent)
- **Kill switch** — `.aiops/disabled` disables all orchestration, escalation, and multi-agency
- **Decisions memory** — `decisions/` directory stores architectural decisions that agents must respect
- **`.go.tmpl` extension** — prevents compiler from treating template Go files as source code

## Phased Activation (Project Maturity)

aiops automatically detects project maturity and adapts AI behavior accordingly.

| Maturity      | Detected when                      | Multi-agency     | Escalation budget | Default mode |
| ------------- | ---------------------------------- | ---------------- | ----------------- | ------------ |
| **bootstrap** | < 10 source files, no CI, no tests | Auto-recommended | 4 per session     | Design-first |
| **active**    | Has source code, building          | Escalation-based | 2 per session     | Single-agent |
| **mature**    | Has CI + tests + packages          | Strict gating    | 1 per session     | Single-agent |

**Bootstrap mode** recommends multi-agency for architecture and risk discovery before any code is written. The generated rules include specific guidance:

- Run `/multiagency design.yaml` to lay out architecture
- Run `/multiagency risks.yaml` to surface unknowns
- Produce `architecture.md`, `risks.md`, `assumptions.md` as one-time snapshots

Maturity transitions automatically when you run `aiops sync` — as the project grows, rules adapt.

## Rules Architecture (Policy vs Implementation)

aiops separates rules into two layers:

| Layer                     | Contains                                                                                                     | Location                                   | Versioned? | Shared?  |
| ------------------------- | ------------------------------------------------------------------------------------------------------------ | ------------------------------------------ | ---------- | -------- |
| **Global (policy)**       | Core principles, escalation policy, intent guardrails, kill switch, repo auto-activation                     | `~/.codeium/windsurf/memories/`            | No         | Per-user |
| **Repo (implementation)** | Maturity behavior, tier routing, escalation mechanics, MCP awareness, session coordination, decisions memory | `.windsurf/rules/`, `.cursor/rules/`, etc. | Yes (git)  | Via git  |

**Global rules** are constitutional law — they apply everywhere and never reference repo-specific details. They include a **repo auto-activation** directive: if a repo has `.aiops/` config or IDE rules, adopt them immediately.

**Repo rules** are statutes — they define how this specific project operates. They include maturity-aware behavior, project-specific patterns, and escalation transparency.

**Kill switch:** Create `.aiops/disabled` in any repo to disable all orchestration, escalation, and multi-agency. The agent operates as a plain single-agent.

**Decisions memory:** The `decisions/` directory stores architectural decisions (ADRs). Agents read these at session start and must not contradict them without escalation.

## Design Principles

- **Scan, don't configure** — detect the stack, don't ask 20 questions
- **Policy vs implementation** — global rules define _when_ to escalate; repo rules define _how_
- **Templates, not copy-paste** — templates are parameterized by detected stack
- **Phased activation** — multi-agency is a thinking tool at start, gated at scale
- **Baseline vs project state** — aiops generates baseline artifacts; project-specific learning stays in separate files
- **Proposals, not mutation** — the evolution audit proposes changes, humans approve
- **Human approves everything** — aiops generates, it never auto-applies

## What Gets Detected

### Languages

Go, TypeScript/JavaScript, Python, Rust, Java, Ruby, Elixir

### Frameworks

Go: eventsrc, temporal, gin, chi, fiber, mqtt (paho)
TypeScript: Next.js, React, Vue, Svelte, Angular, Tailwind CSS
Python: Django, FastAPI, Flask

### Patterns

Domain-driven design, event-sourcing, code-generation, monorepo, containerized, GitHub Actions, GitLab CI, MCP server

### MCP Servers

Auto-detected from all known config locations:

| Location                              | Source label |
| ------------------------------------- | ------------ |
| `~/.codeium/windsurf/mcp_config.json` | windsurf     |
| `~/.cursor/mcp.json`                  | cursor       |
| `~/.continue/config.json`             | continue     |
| `.cursor/mcp.json` (project)          | cursor       |
| `.vscode/mcp.json` (project)          | vscode       |
| `.windsurf/mcp_config.json` (project) | windsurf     |
| `mcp.json` (project root)             | project      |

Detected servers are stored in `.aiops.yaml` and injected into generated rules so the AI knows which MCP tools are available and should be used proactively.

### Build Commands

Auto-detected based on language and framework. Includes build, test, and code generation commands.

## Future

- Plugin system for custom detectors
- `aiops update` across repos (pull from a shared template repo)
- Version pinning for templates
- Target-specific template overrides (e.g., Cursor-specific prompt format)
- `aiops watch` — file watcher for auto-sync on MCP config changes
