# trail

A CLI planning tool that keeps persistent plan files for AI coding agent sessions.

Plans are pure Markdown — the agent reads and edits them directly. Trail handles scaffolding, tracking which plan is active, and parsing task progress.

## Install

```bash
go install github.com/adham90/trail@latest
```

## Quick Start

```bash
trail plan "auth-rewrite"   # Create a new plan (or select existing)
# Edit plans/auth-rewrite.md — add tasks, specs, notes
trail status                 # See progress across all plans
trail archive                # Archive when done
```

## Commands

| Command | Description |
|---------|-------------|
| `trail plan "name"` | Create plan (or select if it exists) |
| `trail plan` | List all plans with progress |
| `trail status` | Show progress across all plans |
| `trail archive [name]` | Archive a completed plan |

## Plan Format

```markdown
# Auth Rewrite

## Tasks

- [x] Set up OAuth2 provider
  - [x] Configure client credentials
  - [x] Test token exchange

- [ ] Replace JWT middleware
  Swap out JWT validation for OAuth2 token introspection.
  - [ ] Update middleware chain
  - [ ] Integration tests pass

- [ ] Migrate sessions

## Notes

Check with DevOps on token rotation policy.
```

Trail counts **only top-level checkboxes** (no leading whitespace) under `## Tasks`. Indented checkboxes (sub-tasks), descriptions, and other sections are the agent's responsibility.

## CLAUDE.md Instructions

Add to your project's `CLAUDE.md` so the agent knows how to use trail:

````markdown
## Planning: trail

Use `trail` for planning across sessions. Plans live in `plans/` as Markdown — read and edit them directly for tasks, specs, notes.

- `trail plan "name"` — create or select a plan
- `trail plan` — list all plans
- `trail status` — progress overview
- `trail archive` — archive completed plan

Plan format: top-level `- [ ]` / `- [x]` under `## Tasks` are counted for progress. Sub-tasks (indented) are for your own tracking.
````

## How It Works

- Plans live in `plans/` at the git root — visible and committed
- `plans/.current` tracks the active plan (gitignored)
- Atomic writes via temp file + rename — never corrupts plan files
- The agent owns all plan content — trail only scaffolds and reads

## Building from source

```bash
git clone https://github.com/adham90/trail.git
cd trail
go build ./...
go test ./...
```

## License

MIT
