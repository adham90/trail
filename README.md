# trail

A CLI planning tool that keeps persistent plan files for AI coding agents.

Trail solves a fundamental problem with AI-assisted coding: **every new session starts blind.** Trail gives each project a persistent plan file that acts as both a task tracker and implementation spec — so the agent knows what's done, what's next, and exactly how to build it.

## Install

```bash
go install github.com/adham90/trail@latest
```

Or download a binary from [Releases](https://github.com/adham90/trail/releases).

## Quick Start

```bash
# Create a plan (auto-creates a git branch)
trail plan --new deploy-pipeline --goal "Add deployment pipeline"

# Add tasks with implementation specs
trail add "Define deploy endpoint" \
  --spec "REST endpoint that triggers DeployJob" \
  --verify "POST /deploys returns 201","creates Deploy record" \
  --files "app/controllers/deploys_controller.rb"

# Start working
trail next

# Track progress
trail checkpoint --file main.go --tests "12 passing"

# Log decisions
trail decide "Use Turbo Streams over ActionCable — simpler"

# Complete the plan
trail done
```

## How It Works

Trail stores plans as Markdown files with YAML frontmatter in a `plans/` directory at your git root. The YAML holds all structured state. Below it, trail renders a human-readable Markdown view that looks great in any editor.

```
project/
  plans/
    deploy-pipeline.md    # active plan
    auth-rewrite.md       # another active plan
    archive/
      onboarding.md       # completed
```

### Plan file format

```yaml
---
name: deploy-pipeline
goal: Add deployment pipeline with status broadcasting
branch: plan/deploy-pipeline
status: active
session_count: 2
constraints:
  - All tests must pass before moving to next task
  - Atomic file writes only
files:
  - path: app/controllers/deploys_controller.rb
    role: REST endpoint
tasks:
  - text: Define deploy endpoint
    status: active
    spec: |
      REST endpoint that triggers DeployJob.
      Accept deploy_id in params.
    verify:
      - "POST /deploys returns 201"
      - "creates Deploy record"
    files:
      - app/controllers/deploys_controller.rb
  - text: Broadcast status updates
    status: todo
context:
  current_file: app/controllers/deploys_controller.rb
  last_error: ~
  test_state: 8 passing
decisions:
  - time: 2026-03-14T10:00:00Z
    text: Turbo Streams over ActionCable
---
```

Below the YAML, trail auto-generates a readable Markdown view with task checkboxes, the active task expanded with its spec/verify/files, a context table, and decisions list.

## Commands

| Command | Description |
|---|---|
| `trail plan` | List all plans |
| `trail plan <name>` | Open a plan |
| `trail plan --new <name> --goal "..."` | Create a plan (with git branch) |
| `trail plan --new <name> --goal "..." --no-branch` | Create without a branch |
| `trail use <name>` | Set current plan, switch to its branch |
| `trail next` | Complete active task, activate next |
| `trail next --skip` | Skip to next without completing |
| `trail add "task"` | Add a task |
| `trail add "task" --spec "..." --verify "a,b" --files "x,y"` | Add with full spec |
| `trail edit N "new text"` | Reword a task |
| `trail block "reason"` | Block the active task |
| `trail block N "reason"` | Block task by index |
| `trail checkpoint --file x --tests "..." --error "..."` | Save context |
| `trail checkpoint --verify "step text"` | Mark a verify step as passed |
| `trail decide "reason"` | Log a decision |
| `trail done` | Complete and archive the plan |
| `trail status` | Show all plans with progress |
| `trail resume` | Print CLAUDE.md + plan for session handoff |
| `trail undo` | Revert the last write |

## Agent Workflow

### Starting a new session

```bash
trail resume deploy-pipeline
```

This prints your `CLAUDE.md` and the full plan file — paste it into Claude Code as the first message.

### Working through tasks

The agent reads the active task's `spec`, implements it, runs each `verify` step, and advances:

```bash
trail next                                    # activate first task
# ... agent implements ...
trail checkpoint --verify "POST returns 201"  # mark verify passed
trail checkpoint --verify "creates record"    # mark another
trail checkpoint --tests "10 passing"         # save test state
trail next                                    # complete, activate next
```

### Resuming after context loss

When a Claude Code session ends, the plan persists. Next session:

```bash
trail resume
# Paste output into new session
# Agent picks up exactly where you left off
```

## Design Principles

- **The plan file IS the state.** No database, no config files, no lock-in.
- **YAML is the source of truth.** The rendered Markdown below is regenerated on every write.
- **Atomic writes.** Every write goes to a temp file first, then `os.Rename()`. Plans never corrupt on crash.
- **Visible and committed.** Plans live in `plans/`, not hidden. They're project documentation — commit them.
- **Each task is an implementation spec.** Not just a title — `spec` tells the agent what to build, `verify` tells it how to check, `files` tells it where to look.
- **Constraints are global rules.** The agent reads them before every task.
- **Monochrome output.** Symbols `✓ ▶ ! ○`, ANSI bold, nothing else.

## Building from source

```bash
git clone https://github.com/adham90/trail.git
cd trail
go build ./...
go test ./...
```

## License

MIT
