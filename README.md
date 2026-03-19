# trail

A CLI planning tool that keeps persistent plan files for AI coding agent sessions.

Plans are pure Markdown — the agent reads and edits them directly. Trail handles scaffolding, tracking which plan is active, and parsing task progress.

## Install in Claude Code

### From a plugin marketplace (recommended)

```bash
# 1. Add the marketplace (one-time)
/plugin marketplace add adham90/trail

# 2. Install the plugin
/plugin install trail@trail
```

That's it. Trail is now active in your project with automatic session recovery.

To manage the plugin later:

```bash
/plugin disable trail@trail      # Disable without uninstalling
/plugin enable trail@trail       # Re-enable
/plugin uninstall trail@trail    # Remove completely
```

### Manual installation

Clone the plugin into your project's `.claude/plugins/` directory:

```bash
mkdir -p .claude/plugins
git clone https://github.com/adham90/trail.git .claude/plugins/trail
```

### CLI binary (optional)

The plugin works without the Go binary — hook scripts parse plan files directly. Install the binary for richer `trail status` output and atomic file operations:

```bash
go install github.com/adham90/trail@latest
```

## Usage

### Getting started

Once installed, initialize trail in your project:

```
trail init
```

This creates:
- `plans/` directory for your plan files
- Adds `plans/.current` to `.gitignore`
- Appends trail instructions to your `CLAUDE.md`

### Creating and working with plans

```bash
trail plan "auth-rewrite"    # Create a new plan (or select existing)
# Edit plans/auth-rewrite.md — add tasks, specs, notes
trail status                 # See progress across all plans
trail archive                # Archive when done
```

### What happens automatically

When installed as a plugin, trail hooks into Claude Code's lifecycle:

1. **Every message you send** — trail detects your active plan and shows progress:
   ```
   [trail] Active plan: Auth Rewrite (3/7 tasks done)
   [trail] Read plans/auth-rewrite.md to resume work.
   ```
2. **Before the agent stops** — trail reminds it to update checkboxes:
   ```
   [trail] Progress: 5/7 tasks done.
   [trail] Update checkboxes in plans/auth-rewrite.md before stopping.
   ```

This means starting a new session or running `/clear` never loses plan context.

### Working without the binary

If you don't have the `trail` binary installed, you can work with plans directly:

1. Create `plans/` at your git root
2. Write a Markdown file (e.g., `plans/my-plan.md`) using the format below
3. Write the plan name to `plans/.current` to make it active

The plugin hooks will still detect and display progress from the raw Markdown.

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

## Adding to CLAUDE.md

Running `trail init` automatically adds instructions to your project's `CLAUDE.md`. If you prefer to add them manually:

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
