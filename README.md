# trail

A CLI planning tool that keeps persistent plan files for AI coding agent sessions.

Plans are pure Markdown — the agent writes and maintains them directly. Trail handles scaffolding, status parsing, and checkbox manipulation.

## Install

```bash
go install github.com/adham90/trail@latest
```

## Quick Start

```bash
trail plan --new "auth-rewrite" --goal "Replace JWT middleware with OAuth2" --open
# Plan opens in $EDITOR — define your tasks
trail done 1          # Mark task 1 complete
trail block 2 "waiting on API keys"
trail status          # See progress
trail resume          # Print plan for session handoff
```

## Commands

| Command | Description |
|---------|-------------|
| `trail plan --new "name" --goal "..."` | Create plan from template (`--open` to open in `$EDITOR`) |
| `trail plan` | List all plans |
| `trail status` | Show progress across plans |
| `trail use "name"` | Set active plan |
| `trail done N` | Mark task N as `[x]` (1-based) |
| `trail block N "reason"` | Mark task N as blocked |
| `trail prompt` | Output format guide |
| `trail resume` | Print plan for session handoff |
| `trail archive [name]` | Archive a completed plan |
| `trail undo` | Revert last write |

## Plan Format

```markdown
# Auth Rewrite

Replace JWT middleware with OAuth2.

## Acceptance Criteria

- [ ] All endpoints use OAuth2
- [ ] Existing sessions migrated

## Tasks

- [x] **1.** Set up OAuth2 provider
  - [x] 1.1. Configure client credentials
  - [x] 1.2. Test token exchange
  `auth/provider.go`

- [ ] **2.** Replace JWT middleware
  Swap out the JWT validation for OAuth2 token introspection.
  - [ ] 2.1. Update middleware chain
  - [ ] 2.2. Integration tests pass
  `middleware/auth.go`, `middleware/auth_test.go`

- [ ] **3.** Migrate sessions [blocked: waiting on DB migration]

## Decisions

- 2026-03-15: Use OAuth2 over SAML — simpler for our use case

## Notes

Check with DevOps on token rotation policy.
```

Trail parses **only top-level checkboxes** under `## Tasks`. Sub-items, descriptions, and other sections are maintained by the agent.

## CLAUDE.md Instructions

Add the following to your project's `CLAUDE.md` so the agent knows how to use trail. You can also generate this with `trail prompt`.

````markdown
## Planning & Task Management: trail

Use `trail` for planning and task management across sessions. Plans live in `plans/` as pure Markdown — edit directly to add tasks, specs, decisions, notes. Use `trail done N` / `trail block N "reason"` for checkbox changes only.

- `trail plan --new "name" --goal "..."` — create plan (auto-creates git branch)
- `trail plan` — list all plans
- `trail use "name"` — set active plan
- `trail done N` — mark task N done (1-based)
- `trail block N "reason"` — mark task N blocked
- `trail status` — progress overview
- `trail archive` — archive completed plan
- `trail resume` — print plan for session handoff
- `trail undo` — revert last change
- `trail prompt` — output this format guide

Before creating a plan, run `trail prompt` to review the expected format. Trail parses ONLY top-level `- [ ]` / `- [x]` under `## Tasks`. Keep that heading exact. Sub-items are ignored.
````

## How It Works

- Plans live in `plans/` at the git root — visible and committed
- `plans/.current` tracks the active plan
- `plans/.backup` holds previous state for undo
- Branch `plan/<name>` is auto-created with `--new` (use `--no-branch` to skip)
- Atomic writes via temp file + rename — never corrupts plan files

## Building from source

```bash
git clone https://github.com/adham90/trail.git
cd trail
go build ./...
go test ./...
```

## License

MIT
