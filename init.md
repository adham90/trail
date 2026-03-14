---
goal: Build trail — a Go CLI planning tool for Claude Code
branch: main
status: active
session_count: 1
created: 2026-03-14
updated: 2026-03-14
current_task: 0
---

## tasks

- [ ] 00 · Init Go module and repo structure
- [ ] 01 · Add cobra + yaml dependencies
- [ ] 02 · Wire root command with version flag
- [ ] 03 · Write CLAUDE.md for this repo
- [ ] 04 · Define Plan struct with all fields (Option A — all state in YAML frontmatter)
- [ ] 05 · Write YAML frontmatter parser (read)
- [ ] 06 · Write plan serializer (write — YAML frontmatter + raw notes)
- [ ] 07 · Resolve .plans/ dir from git root
- [ ] 08 · trail plan — open or create plan, auto-init .gitignore, print session summary
- [ ] 09 · trail next — mark active done, activate next task (--skip flag to advance without completing)
- [ ] 10 · trail checkpoint — write context state to plan
- [ ] 11 · trail decide — append timestamped decision
- [ ] 12 · trail block — block current active task (optional N index override)
- [ ] 13 · trail add — insert new task at end or after index
- [ ] 14 · trail edit N "text" — reword a task
- [ ] 15 · trail done — complete plan, archive to .plans/archive/
- [ ] 16 · trail resume — print CLAUDE.md + active plan for session handoff
- [ ] 17 · trail status — list all active plans across branches
- [ ] 18 · trail undo — revert last write via .plans/.backup
- [ ] 19 · Build session summary renderer (trail plan output)
- [ ] 20 · Build context block renderer (trail next output)
- [ ] 21 · Style output — fmt + ANSI bold, status symbols, monochrome
- [ ] 22 · Auto-increment session_count on first trail plan call per day
- [ ] 23 · GoReleaser config for binary builds

## context

current_file: ~
last_error: ~
test_state: ~
open_questions: ~
pending_refactor: ~

## decisions

- 2026-03-14 · Go over Ruby — single binary, zero dep conflicts, works before project is set up
- 2026-03-14 · Markdown + YAML frontmatter — machine-parseable state + freeform notes in one file
- 2026-03-14 · .plans/{branch}.md at git root — auto-associated with branch, gitignored by default
- 2026-03-14 · Cobra for CLI — standard Go pattern, same mental model as Thor
- 2026-03-14 · No database, no config file — the .md file IS the state, nothing else
- 2026-03-14 · Atomic file writes via temp file + rename — never corrupt a plan on crash
- 2026-03-14 · Option A file format — all structured data in YAML frontmatter, Markdown body is freeform notes only. Renderer handles pretty output, not the file.
- 2026-03-14 · Drop lipgloss — fmt.Fprintf + ANSI bold is enough for monochrome symbols. Revisit post-v1.
- 2026-03-14 · Fold trail init into trail plan — auto-create .plans/ and append .gitignore on first use. No separate command.
- 2026-03-14 · Defer Homebrew tap — go install + goreleaser GitHub releases are enough for v1.
- 2026-03-14 · trail block defaults to current active task — index override is optional, not required.
- 2026-03-14 · Backup before every write — .plans/.backup holds previous state for trail undo.

## notes

### plan file format (Option A)

All structured state lives in YAML frontmatter. The Markdown body below `---` is freeform notes only. The pretty terminal output is the renderer's job, not the file format's job.

Example .plans/feat-deploy-pipeline.md:
```
---
goal: Add deployment pipeline
branch: feat-deploy-pipeline
status: active
session_count: 2
created: 2026-03-14
updated: 2026-03-14
current_task: 2
tasks:
  - text: Add DeployJob to Solid Queue
    status: done
  - text: Define deploy endpoint
    status: blocked
    reason: waiting on DevOps
  - text: Broadcast status via Turbo Stream
    status: active
  - text: Connect sidebar button
    status: todo
context:
  current_file: app/jobs/deploy_job.rb
  last_error: ~
  test_state: 4 passing
decisions:
  - time: 2026-03-14T10:00:00Z
    text: Turbo Streams over ActionCable
---

## notes

SSH key injection needs .env.test with KEY_PATH set.
Check with DevOps before task 04.
```

### file structure

```
trail/
  cmd/
    root.go
    plan.go
    next.go
    checkpoint.go
    decide.go
    block.go
    add.go
    edit.go
    done.go
    resume.go
    status.go
    undo.go
  internal/
    plan/
      model.go      — Plan struct (all fields map to YAML frontmatter)
      parser.go     — read: split frontmatter + notes, yaml.Unmarshal
      writer.go     — write: yaml.Marshal + notes, temp file + rename, backup
      resolve.go    — walk up from cwd to find .git/, derive plan path from branch
      backup.go     — copy current file to .plans/.backup before each write
    renderer/
      summary.go    — trail plan output (full session state)
      context.go    — trail next output (compact context block)
      styles.go     — ANSI bold, status symbols (✓ ▶ ! ○), spacing
  main.go
  go.mod
  CLAUDE.md
  .goreleaser.yml
```

### dependencies

```
github.com/spf13/cobra
gopkg.in/yaml.v3
```

### command specs

**trail plan**
- Look up current git branch
- Derive plan path: .plans/{branch-name}.md (replace / with -)
- If .plans/ doesn't exist: create it, append to .gitignore (idempotent)
- If file missing: prompt for goal, create empty plan, print created message
- If file exists: parse and print full session summary via renderer
- Auto-increment session_count if updated date differs from today
- Blocked tasks surface at the top of the task list

**trail next [--skip]**
- Default: find active task, mark done. Find next todo, mark active.
- --skip: find active task, mark todo (not done). Find next todo, mark active.
- Write updated plan file
- Print compact context block (current task + context{} fields)

**trail checkpoint --file --error --tests --note**
- All flags optional
- Updates context{} fields in frontmatter with provided values
- Stamps updated timestamp
- Prints confirmation line

**trail decide "reason string"**
- Appends to decisions[] with RFC3339 timestamp
- Prints confirmation line

**trail block ["reason"]**
- Defaults to current active task
- Optional: trail block N "reason" to block by index
- Sets task status=blocked, stores reason
- Prints updated task line

**trail add "task description" [--after N]**
- Appends new task with status=todo
- --after N: insert after task index N instead of appending
- Prints confirmation with task index

**trail edit N "new text"**
- Replaces text of task at index N
- Prints updated task line

**trail done**
- Sets status=complete in frontmatter
- Stamps updated timestamp
- Moves file to .plans/archive/{branch}.md
- Prints summary: N/N tasks, N decisions, N sessions

**trail resume**
- Prints contents of CLAUDE.md (if exists) followed by active plan file
- Designed to be copy-pasted into a new Claude Code session
- If no active plan, prints error

**trail status**
- Lists all .plans/*.md files (excluding archive/)
- For each: goal, branch, status, task progress (done/total)
- Sorted by updated timestamp descending

**trail undo**
- Copies .plans/.backup over the current plan file
- Prints what was reverted (diff summary or just "reverted last change")
- Only one level of undo (single backup)

### renderer output spec

trail plan output:
```
goal:     Build trail — a Go CLI planning tool for Claude Code
branch:   feat/deploy-pipeline
session:  3

! 04 · SSH key injection          blocked: .env.test missing key path

✓ 00 · Add DeployJob to Solid Queue
✓ 01 · Wire up Solid Queue adapter
▶ 02 · Broadcast status via Turbo Stream
○ 03 · Connect Forge sidebar button
○ 04 · SSH key injection

context:
  current_file:  app/models/deploy.rb:34
  last_error:    NoMethodError on broadcast_update
  test_state:    3 failing · deploy_job_spec.rb

decisions: 3 logged
```

trail next output:
```
✓ Add DeployJob to Solid Queue
▶ Define deploy endpoint in DeploysController

context:
  current_file:  ~
  last_error:    ~
  test_state:    ~
```

trail status output:
```
feat-deploy-pipeline   active     5/7    session 2
fix-login-redirect     active     2/4    session 1
main                   complete   20/20  archived
```

Status symbols: ✓ done · ▶ active · ! blocked · ○ todo
Keep output monochrome — no rainbow colors

### key constraints

- Atomic writes: always write to a temp file then os.Rename(), never write directly
- Backup before every write: copy current file to .plans/.backup first
- Branch name → filename: replace all / with - (feat/deploy → feat-deploy.md)
- All empty YAML fields use ~ not "" or null
- Git root resolution: walk up from os.Getwd() until directory containing .git/ is found
- Go 1.22+ minimum
- File format: all structured data in YAML frontmatter, body is freeform notes only

### opening prompt for claude code (session 1)

```
Read this plan file top to bottom, then start from task 00.
Work through each task in order. After each task is complete,
run: trail checkpoint --file <file> and confirm before moving on.

Key constraints to keep in mind:
- All state in YAML frontmatter, notes section is freeform Markdown
- Atomic file writes (temp + rename)
- Backup before every write (.plans/.backup)
- ~ for all empty YAML fields
- Monochrome renderer output (fmt + ANSI bold, no lipgloss)
- Walk up from cwd to find git root

Start with: go mod init github.com/yourname/trail
```

### resuming a session (subsequent sessions)

Run:
```
trail resume
```

Then paste the output into Claude Code, followed by:
```
Continue from the current active task. Run trail checkpoint
after each completed task.
```
