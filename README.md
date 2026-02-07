# aiops

AI-powered operations infrastructure for your codebase. Installs senior-engineer behavior as code — rules, workflows, coordination, and evolution — for Windsurf/Cascade AI sessions.

## What It Does

`aiops` scans your repository, detects your tech stack, and generates a complete set of AI session management artifacts:

- **Global rules** — behavioral constitution loaded into every Cascade session
- **Orchestrator** — cross-session coordination with advisory locks and build failure protocol
- **Workflows** — default execution mode, evolution audits, multiagency specs
- **Multiagency module** — complete Go module with CLI, spec parser, LLM clients, agent executor, and pipeline orchestrator
- **Intent guardrails** — prevents agents from drifting outside task scope
- **Escalation budget** — prevents over-cautious behavior
- **Human override** (`@directive`) — escape hatch that overrides process, not safety

## Quick Start

```bash
# Build the CLI
cd aiops && go build -o aiops ./cmd/aiops

# Initialize in your project
./aiops init --dir /path/to/your/project

# Check what's installed
./aiops status --dir /path/to/your/project
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

Is this correct? [Y/n] y

Project name [myproject]: myproject

✓ Created .aiops.yaml

Generating Cascade artifacts...
  ✓ global_rules.md
  ✓ .windsurf/workflows/default-mode.md
  ✓ .windsurf/workflows/orchestrator.md
  ✓ .windsurf/workflows/multiagency.md
  ✓ .windsurf/orchestrator/session_state.yaml
  ✓ multiagency/go.mod
  ✓ multiagency/README.md
  ✓ multiagency/cmd/multiagency/main.go
  ✓ multiagency/internal/spec/types.go
  ✓ multiagency/internal/spec/loader.go
  ✓ multiagency/internal/llm/client.go
  ✓ multiagency/internal/llm/stub.go
  ✓ multiagency/internal/llm/anthropic.go
  ✓ multiagency/internal/agent/executor.go
  ✓ multiagency/internal/agent/prompt.go
  ✓ multiagency/internal/pipeline/context.go
  ✓ multiagency/internal/pipeline/executor.go
  ✓ multiagency/specs/design.yaml
  ✓ multiagency/specs/code_review.yaml
  ✓ multiagency/specs/manager.yaml
  ✓ multiagency/specs/evolution_audit.yaml

✅ aiops initialized! 21 files generated.
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

### `aiops status`

Shows what's installed and detects drift (new frameworks, languages, patterns added since last init).

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

Auto-generates skill scaffolds based on detected frameworks. Skills are directories in `.windsurf/skills/` that Cascade auto-invokes based on task type.

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

## What Gets Generated

### `aiops init` — Cascade artifacts

| File                                        | Purpose                                  |
| ------------------------------------------- | ---------------------------------------- |
| `.aiops.yaml`                               | Project config (detected stack, paths)   |
| `global_rules.md` (Windsurf memories)       | Compact behavioral rules — always active |
| `.windsurf/workflows/default-mode.md`       | Detailed reference manual for Cascade    |
| `.windsurf/workflows/orchestrator.md`       | Session coordination commands            |
| `.windsurf/workflows/multiagency.md`        | Multi-agent workflow executor            |
| `.windsurf/orchestrator/session_state.yaml` | Shared state for parallel sessions       |

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
├── cmd/aiops/main.go               # CLI (init, scan, status, update, evolve, skills)
├── internal/
│   ├── config/config.go            # .aiops.yaml schema and I/O
│   ├── scanner/scanner.go          # Repo analysis + Go module detection
│   ├── renderer/
│   │   ├── renderer.go             # Template rendering engine (embed.FS)
│   │   └── templates/              # Embedded Go templates
│   │       ├── memories/           # → Windsurf memories directory
│   │       ├── windsurf/           # → .windsurf/ directory
│   │       └── multiagency/        # → Complete Go module
│   │           ├── go.mod.tmpl
│   │           ├── README.md.tmpl
│   │           ├── cmd/multiagency/main.go.tmpl
│   │           ├── internal/spec/  # types.go.tmpl, loader.go.tmpl
│   │           ├── internal/llm/   # client.go.tmpl, stub.go.tmpl, anthropic.go.tmpl
│   │           ├── internal/agent/ # executor.go.tmpl, prompt.go.tmpl
│   │           ├── internal/pipeline/ # context.go.tmpl, executor.go.tmpl
│   │           └── specs/          # design.yaml, code_review.yaml, ...
│   ├── updater/updater.go          # Diff and apply template updates
│   ├── evolve/evolve.go            # Directive log analysis and rule proposals
│   └── skills/skills.go            # Framework-specific skill scaffold generation
└── README.md
```

Go source templates use `.go.tmpl` extension to prevent the compiler from treating them as source code. Import paths are templatized with `{{.MultiagencyMod}}`, derived from the detected Go module path.

## Design Principles

- **Scan, don't configure** — detect the stack, don't ask 20 questions
- **Templates, not copy-paste** — templates are parameterized by detected stack
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

### Build Commands

Auto-detected based on language and framework. Includes build, test, and code generation commands.

## Future

- Plugin system for custom detectors
- `aiops update` across repos (pull from a shared template repo)
- Version pinning for templates
- Extract to standalone repo for distribution
