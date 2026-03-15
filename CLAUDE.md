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

- `cmd/` — Cobra commands (plan, next, checkpoint, decide, block, add, edit, done, resume, status, undo, use)
- `internal/plan/` — Plan model, YAML parser/writer, resolver, atomic file ops, git operations
- `internal/renderer/` — Terminal output (summary view, context block, ANSI styles)

## Plans

- Plans live in `plans/` at the git root — visible, committed to git
- Each plan is a named `.md` file: `plans/deploy-pipeline.md`
- `plans/.current` tracks the active plan (gitignored)
- `plans/.backup` holds previous state for undo (gitignored)
- `plans/archive/` holds completed plans

## File Format

All structured state lives in YAML frontmatter. Below `---`, trail generates a readable Markdown view (constraints, files, tasks with active task expanded, context, decisions, notes). YAML is the source of truth.

## Key Constraints

- Atomic writes: temp file + `os.Rename()` — never write directly
- Backup before every write
- Empty YAML fields use `~` (null), not `""`
- Plan name → filename: lowercase, spaces/slashes to dashes
- Plan name → branch: `plan/<name>`
- Git root: walk up from cwd to find `.git/` directory
- Monochrome output: `fmt` + ANSI bold, symbols `✓ ▶ ! ○`
