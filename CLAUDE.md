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

- `cmd/` — Cobra commands (plan, status, archive)
- `internal/plan/` — Markdown parser, template generator, atomic file ops
- `internal/renderer/` — Terminal output (ANSI bold, symbols ✓ ○)

## Plans

- Plans live in `plans/` at the git root — visible, committed to git
- Each plan is a named `.md` file: `plans/deploy-pipeline.md`
- `plans/.current` tracks the active plan (gitignored)
- `plans/archive/` holds completed plans

## File Format

Plans are pure Markdown files. The coding agent writes and maintains them directly. Trail parses ONLY top-level checkboxes (no leading whitespace) under `## Tasks` for progress counting.

```markdown
# Plan Name

## Tasks

- [ ] First task
  Description/spec.
  - [ ] verify step

- [x] Completed task

## Notes

Freeform.
```

## Commands

| Command | Description |
|---------|-------------|
| `trail init` | Set up trail in current project (.gitignore, CLAUDE.md) |
| `trail plan "name"` | Create plan (or select if exists) |
| `trail plan` | List all plans with progress |
| `trail status` | Show progress across all plans |
| `trail archive [name]` | Archive a completed plan |

## Plugin Structure

Trail is also a Claude Code plugin with automatic session recovery:

- `.claude-plugin/plugin.json` — Plugin metadata
- `skills/trail/SKILL.md` — Skill definition with UserPromptSubmit and Stop hooks
- `skills/trail/scripts/session-recover.sh` — Detects active plan on every user message
- `skills/trail/scripts/check-progress.sh` — Shows progress before agent stops

Hook scripts are shell-only (no binary dependency). They fall back to parsing plan files directly if the `trail` binary isn't in PATH.

## Key Constraints

- Atomic writes: temp file + `os.Rename()` — never write directly
- Plan name → filename: lowercase, spaces/slashes to dashes
- Git root: walk up from cwd to find `.git/` directory
- Monochrome output: `fmt` + ANSI bold, symbols `✓ ○`
