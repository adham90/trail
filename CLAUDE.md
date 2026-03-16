# trail

A Go CLI planning tool that keeps persistent plan files for Claude Code sessions.

## Build & Test

```bash
go build ./...          # compile
go test ./...           # run all tests
go run . --version      # check version
go run . <command>      # run any command
```

## Architecture

- `cmd/` — Cobra commands (plan, done, block, archive, status, use, resume, undo, prompt)
- `internal/plan/` — Markdown parser, template generator, atomic file ops, git operations
- `internal/renderer/` — Terminal output (ANSI bold, symbols ✓ ○)

## Plans

- Plans live in `plans/` at the git root — visible, committed to git
- Each plan is a named `.md` file: `plans/deploy-pipeline.md`
- `plans/.current` tracks the active plan (gitignored)
- `plans/.backup` holds previous state for undo (gitignored)
- `plans/archive/` holds completed plans

## File Format

Plans are pure Markdown files. The coding agent writes and maintains them directly. Trail parses ONLY top-level checkboxes under `## Tasks` for status. Task numbering is 1-based.

```markdown
# Plan Name

Goal description.

## Acceptance Criteria

- [ ] criterion 1

## Tasks

- [ ] **1.** Task title
  Description/spec.
  - [ ] 1.1. verify step
  `file1.go`, `file2.go`

- [x] **2.** Completed task

## Decisions (optional)

- 2026-03-16: Decision text

## Notes (optional)

Freeform.
```

## Commands

| Command | Description |
|---------|-------------|
| `trail plan --new "name" --goal "..."` | Create plan from template |
| `trail plan` | List all plans |
| `trail status` | Show progress across plans |
| `trail use "name"` | Set active plan |
| `trail done N` | Mark task N as `[x]` (1-based) |
| `trail block N "reason"` | Mark task N as blocked |
| `trail archive [name]` | Archive a completed plan |
| `trail prompt` | Output format guide for CLAUDE.md |
| `trail resume` | Print plan for session handoff |
| `trail undo` | Revert last write |

## Key Constraints

- Atomic writes: temp file + `os.Rename()` — never write directly
- Backup before every write
- Plan name → filename: lowercase, spaces/slashes to dashes
- Plan name → branch: `plan/<name>`
- Git root: walk up from cwd to find `.git/` directory
- Monochrome output: `fmt` + ANSI bold, symbols `✓ ○`
