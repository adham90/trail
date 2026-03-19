# trail

A CLI planning tool that keeps persistent plan files for AI coding agent sessions.

Plans are pure Markdown — the agent reads and edits them directly. Trail handles scaffolding, tracking which plan is active, and parsing task progress.

## Install

### As a Claude Code plugin (recommended)

Trail works as a Claude Code plugin with automatic session recovery — when you start a new session, the agent automatically knows about your active plan.

```bash
# Via plugin marketplace
/plugin marketplace add adham90/trail
/plugin install trail@trail
```

Or manually:
```bash
mkdir -p .claude/plugins
git clone https://github.com/adham90/trail.git .claude/plugins/trail
```

### CLI binary (optional)

Install the Go binary for richer `trail status` output and atomic file operations:

```bash
go install github.com/adham90/trail@latest
```

The plugin works without the binary — hook scripts parse plan files directly as a fallback.

## Quick Start

```bash
trail init                   # Set up trail in your project
trail plan "auth-rewrite"    # Create a new plan (or select existing)
# Edit plans/auth-rewrite.md — add tasks, specs, notes
trail status                 # See progress across all plans
trail archive                # Archive when done
```

## Commands

| Command | Description |
|---------|-------------|
| `trail init` | Set up trail in current project (creates plans/, .gitignore, CLAUDE.md) |
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

## Session Recovery

When installed as a plugin, trail automatically recovers context across sessions:

1. **User sends a message** — the `UserPromptSubmit` hook checks for `plans/.current`
2. **Agent sees the active plan** — e.g., `[trail] Active plan: Auth Rewrite (3/7 tasks done)`
3. **Agent reads the plan file** and resumes work with full context
4. **Before stopping** — the `Stop` hook reminds to update checkboxes

This means `/clear` or starting a new session never loses plan context.

## How It Works

- Plans live in `plans/` at the git root — visible and committed
- `plans/.current` tracks the active plan (gitignored)
- Atomic writes via temp file + rename — never corrupts plan files
- The agent owns all plan content — trail only scaffolds and reads
- Plugin hooks provide automatic session recovery without agent configuration

## Building from source

```bash
git clone https://github.com/adham90/trail.git
cd trail
go build ./...
go test ./...
```

## License

MIT
