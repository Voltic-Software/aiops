---
description: Session coordination - sync state, claim locks, detect conflicts across parallel Cascade sessions
---

# Session Orchestrator

Coordinates parallel Cascade sessions through explicit shared state. Sessions don't know about each other — they synchronize through `.windsurf/orchestrator/session_state.yaml`.

## Commands

When the user runs `/orchestrator`, determine which subcommand they want:

### `/orchestrator sync` — Start of Session

Read the shared state and report what's happening:

1. Read `.windsurf/orchestrator/session_state.yaml`
2. Clean up expired locks (remove locks where `expires_at` is in the past)
3. Report:
   - **Active sessions**: What's in flight, who owns what
   - **Active locks**: Which areas are claimed, expected instability
   - **Conflicts**: If the user's intended work overlaps with locked areas
   - **Global notes**: Cross-cutting information
4. Ask the user:
   - "What is this session's focus?"
   - "Which areas will you touch?"
5. Create a session entry and optionally a lock

### `/orchestrator lock` — Claim an Area

Acquire an advisory lock:

1. Ask the user:
   - Which area? (e.g., `backend`, `frontend`)
   - Specific scope? (e.g., `internal/domain/energy`, `apps/customer`)
   - Expected instability? (will builds break temporarily?)
   - Allow external fixes? (can other sessions fix failures in your area?)
   - Duration? (default: 2 hours)
2. Check for conflicting locks
3. Add the lock to `session_state.yaml`

### `/orchestrator unlock` — Release a Lock

1. Show the user's active locks
2. Confirm which lock to release
3. Remove it from `session_state.yaml`

### `/orchestrator status` — Check Current State

Read and display the current state without modifications:

1. Active sessions with their status
2. Active locks with remaining time
3. Any conflicts or overlaps
4. Global notes

### `/orchestrator complete` — End a Session

1. Ask for a brief outcome summary
2. Move the session entry to `completed_sessions` (keep last 10)
3. Release all locks owned by this session
4. Update `session_state.yaml`

### `/orchestrator note` — Add a Global Note

Add a cross-cutting note that all sessions should see:

1. Ask for the note content
2. Append to `global_notes` in `session_state.yaml`

## Session State Schema

```yaml
active_sessions:
  - id: string              # Unique session identifier (kebab-case)
    focus: string            # What this session is working on
    status: string           # discovery | in_progress | blocked | completed
    touched_areas: string[]  # Codebase paths being modified
    risks: string[]          # Potential side effects
    notes: string[]          # Session-specific notes
    started_at: string       # ISO 8601 timestamp
    last_updated: string     # ISO 8601 timestamp

locks:
  - area: string             # High-level area: backend, frontend, infra
    owner: string            # Session ID that owns this lock
    scope: string[]          # Specific paths within the area
    reason: string           # Why this lock exists
    expected_instability: bool  # If true, build failures are expected
    allow_external_fixes: bool  # If true, other sessions may fix failures
    expires_at: string       # ISO 8601 timestamp — lock auto-expires

completed_sessions:
  - id: string
    focus: string
    completed_at: string
    outcome: string          # Brief summary of what was accomplished
    files_changed: string[]  # Key files that were modified

global_notes: string[]       # Cross-cutting information for all sessions

directive_log: []            # Tracks @directive overrides for evolution analysis
```

## Critical Rules

1. **Never modify another session's entry** — only update your own
2. **Always check locks before fixing build failures** — see Build Failure Protocol below
3. **Locks are advisory** — they signal intent, not hard blocks
4. **Expired locks should be cleaned up** — remove them during `sync`
5. **Keep completed_sessions to last 10** — prune older entries
6. **Never auto-merge decisions** — propose changes, let the user decide

## Build Failure Protocol

**This is the most important behavior change.**

When a build fails, follow this decision tree:

```
Build failed
    ↓
What area failed?
    ↓
Is there an active lock for that area?
    ├─ YES → Am I the lock owner?
    │     ├─ YES → Fix it (it's your responsibility)
    │     └─ NO → Does lock allow external fixes?
    │           ├─ YES → Fix it (but note the lock owner)
    │           └─ NO → REPORT ONLY, do not fix
    │                   "Build failed in [area]. This area is locked by
    │                    [session-id]: [reason]. Expected instability: [yes/no].
    │                    I will not attempt a fix."
    └─ NO → Safe to fix (no ownership conflict)
```

**Key principle**: Agents may observe anything, but may only modify what they explicitly own.

## Directive Logging

When a `@directive` override is used, log it to `session_state.yaml`:

1. Add an entry to `directive_log` with:
   - `session`: current session ID
   - `directive`: the override text
   - `reason`: why the override was needed
   - `timestamp`: ISO 8601
   - `rule_overridden`: which rule was bypassed (e.g., `escalation`, `tier_classification`, `intent_scope`)
2. The evolution audit reads this log to detect patterns
3. If the same rule is overridden 3+ times, the audit proposes changing the default rule

This closes the feedback loop: overrides → patterns → rule proposals → human approval → better defaults.

## Intent Guardrail

Every session has an implicit intent scope. The orchestrator reinforces this:

1. When `/orchestrator sync` registers a session, the `focus` and `touched_areas` define the intent boundary
2. During work, the agent should only modify files/areas within `touched_areas` or directly serving `focus`
3. If the agent needs to expand scope, it must ask the user first
4. Before completing a session (`/orchestrator complete`), verify all changes serve the original `focus`

This prevents the common failure mode: agent starts on Task A, notices a code smell, starts refactoring, and now owns a lock but has drifted from the original intent.

## Integration with Default Mode

At session start, the agent should:
1. Check if `.windsurf/orchestrator/session_state.yaml` exists
2. If it does, briefly scan for active locks relevant to the task
3. If a conflict is detected, warn the user before proceeding
4. This check should be quick and silent unless a conflict exists
