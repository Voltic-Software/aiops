# AIops

AIops is a **repo-aware agent orchestration system** that brings structure, safety,
and scalability to AI-assisted software development.

It installs _senior-engineer behavior as code_ into your repository — rules,
workflows, coordination, and evolution — so AI works the way real teams do.

## Core Ideas

- **Default single-agent mode** for fast, pragmatic work grounded in the codebase
- **Explicit escalation** into multi-agent workflows when complexity or risk demands it
- **Agent roles, skills, and governance** that evolve with the repository
- **Safe parallelism** with intent guardrails, advisory locks, and build-failure isolation
- **Human override** (`@directive`) that bypasses process without bypassing safety

AIops is not just about operations.  
It is about **how AI should work inside a real codebase**.

**Supported IDEs:** Windsurf (Cascade), Cursor, Continue (VS Code), GitHub Copilot  
Auto-detected — AIops generates configuration for all supported IDEs found on your system.

## What It Does

`aiops` scans your repository, detects your tech stack, and generates a complete,
repo-specific AI execution system:

- **Soul** — constitutional philosophy anchoring all AI behavior (`.aiops/soul.md`)
- **Repo rules** — behavioral constitution loaded into every AI session, scoped to the repository
- **Orchestrator** — cross-session coordination with advisory locks and conflict prevention
- **Workflows** — default execution mode, evolution audits, and multi-agent pipelines
- **Multiagency module** — CLI, spec parser, agent executor, and pipeline orchestrator
- **Intent guardrails** — prevent task drift and scope creep
- **Escalation budget** — prevents over-cautious, timid behavior
- **Human override** (`@directive`) — override process without breaking safety

AIops makes AI behave like a disciplined teammate, not an unpredictable intern.

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
  ✓ .aiops/soul.md                                 (constitution)
  ✓ .aiops/soul.local.md                           (local extension)
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

Shows installation status, detected features, and checks for drift.

```
$ aiops status

AIops installed ✔
Version:    0.3.1
Project:    myproject
Maturity:   active
Targets:    windsurf, copilot
MCP:        vera
Skills:     5
Workflows:  8

Artifacts:
  ✓ Soul (constitution)
  ✓ Default mode workflow
  ✓ Multiagency workflow
  ✓ Orchestrator workflow
  ✓ Session state
  ✓ Repo rules (Windsurf (Cascade))

6 installed, 0 missing

Drift check...
✓ No drift detected
```

Scriptable check: `aiops version` returns exit code 0 if installed, or use `aiops status` in CI.

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

### `aiops doctor`

Checks the integrity of the aiops installation — verifies all artifacts exist, soul.md is canonical, version matches, and no orphaned state.

```
$ aiops doctor

aiops doctor — checking installation integrity

  ✓ .aiops.yaml
  ✓ soul.md (canonical)
  ✓ soul.local.md (optional)
  ✓ kill switch (inactive)
  ✓ repo rules (Windsurf (Cascade))
  ✓ repo rules (GitHub Copilot)
  ✓ workflow/default-mode.md
  ✓ workflow/multiagency.md
  ✓ workflow/orchestrator.md
  ✓ session_state.yaml
  ✓ decisions/
  ✓ multiagency/go.mod
  ✓ version (0.2.0)

13 passed, 0 warnings, 0 failed

✅ Installation is healthy.
```

If soul.md has been manually modified, doctor will warn and suggest `aiops sync` to restore the canonical version.

### `aiops uninstall`

Removes all aiops-generated artifacts from the repository. Does **not** remove the global binary or editor settings.

```
$ aiops uninstall

This will remove AIops from this repository.

The following will be deleted:
  - .aiops.yaml
  - .aiops/ (soul.md, soul.local.md)
  - decisions/ (seed only)
  - multiagency/
  - .windsurf/rules/aiops.md
  - .windsurf/workflows/default-mode.md
  - .windsurf/workflows/multiagency.md
  - .windsurf/workflows/orchestrator.md
  - .windsurf/orchestrator/
  - .github/copilot-instructions.md

Global tools and binaries will NOT be removed.

Proceed? [Y/n] y

  ✓ Removed .aiops.yaml
  ✓ Removed multiagency/
  ...

✅ AIops uninstalled. 9 items removed.
```

To skip confirmation (for CI/scripts):

```bash
aiops uninstall --yes
```

**Safety rules:**

- User code is never removed
- `decisions/` is only removed if it contains only the aiops seed file
- Editor settings and global binaries are untouched
- Skills directories are preserved (user-customized content)

## Supported IDE Targets

| Target       | Rules                             | Workflows              | Orchestrator              | Auto-detected by                      |
| ------------ | --------------------------------- | ---------------------- | ------------------------- | ------------------------------------- |
| **Windsurf** | `.windsurf/rules/aiops.md`        | `.windsurf/workflows/` | `.windsurf/orchestrator/` | `~/.codeium/windsurf/` exists         |
| **Cursor**   | `.cursor/rules/aiops.mdc`         | `.cursor/prompts/`     | `.cursor/orchestrator/`   | `.cursor/` or `~/.cursor/` exists     |
| **Continue** | `.continue/rules/aiops.md`        | `.continue/prompts/`   | `.continue/orchestrator/` | `.continue/` or `~/.continue/` exists |
| **Copilot**  | `.github/copilot-instructions.md` | —                      | —                         | `.github/` or `~/.vscode/` exists     |

All targets get the same rules content, adapted to the correct file paths. Templates reference `{{.OrchestrDir}}` so each target's rules point to its own orchestrator location.

## What Gets Generated

### `aiops init` — Soul (constitutional layer)

| File                   | Purpose                                              | Owned by |
| ---------------------- | ---------------------------------------------------- | -------- |
| `.aiops/soul.md`       | Core constitution — identical across all repos       | AIops    |
| `.aiops/soul.local.md` | Optional repo-specific extension — never overwritten | User     |

`soul.md` is always overwritten by `aiops sync` / `aiops update` to keep the constitution canonical. `soul.local.md` is created once and never touched again — it belongs to the repository.

Agents do **not** load `soul.md` directly. Instead, the repo rules embed a distilled reference: _"All behavior must remain consistent with the principles defined in `.aiops/soul.md`."_ The soul influences runtime behavior through the policy layer, not by being injected into every session.

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
│   │       ├── soul/               # → Constitution (soul.md + soul.local.md)
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

- **Constitutional anchor** — `soul.md` defines invariant philosophy; `soul.local.md` allows repo-specific extension
- **Repo-scoped rules** — all rules live inside the repository, versioned with git, shared via git
- **Target abstraction** — each IDE is a `Target` with path mappings for rules, workflows, orchestrator, and skills
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

## Rules Architecture

AIops uses a layered architecture: **constitution → policy → execution**.

| Layer               | File                                 | Contains                                                                            | Owned by | Overwritten by sync? |
| ------------------- | ------------------------------------ | ----------------------------------------------------------------------------------- | -------- | -------------------- |
| **Constitution**    | `.aiops/soul.md`                     | Mission, optimization targets, escalation philosophy, non-negotiables               | AIops    | Yes (always)         |
| **Local extension** | `.aiops/soul.local.md`               | Repo-specific principles (optional)                                                 | User     | No (never)           |
| **Policy (rules)**  | `.windsurf/rules/`, `.cursor/rules/` | Kill switch, core principles, tier routing, escalation, MCP awareness, coordination | AIops    | Yes                  |
| **Execution**       | Workflows, orchestrator, skills      | Default mode, multiagency, session state, skill scaffolds                           | AIops    | Yes                  |

The constitution informs the policy layer. It is not re-read in every session — agents inherit distilled constraints through the repo rules.

**Kill switch:** Create `.aiops/disabled` in any repo to disable all orchestration, escalation, and multi-agency. The agent operates as a plain single-agent.

**Decisions memory:** The `decisions/` directory stores architectural decisions (ADRs). Agents read these at session start and must not contradict them without escalation.

## Design Principles

- **Scan, don't configure** — detect the stack, don't ask 20 questions
- **Repo-scoped** — all rules live in the repo, versioned and shared via git
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
