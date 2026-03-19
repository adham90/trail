---
name: trail
description: Persistent markdown planning with session recovery
hooks:
  - event: UserPromptSubmit
    script: skills/trail/scripts/session-recover.sh
  - event: Stop
    script: skills/trail/scripts/check-progress.sh
---

# Trail — Planning with persistent markdown files

You have access to the `trail` planning system. Plans are markdown files in `plans/` at the git root.

## How plans work

- Each plan is a `.md` file in `plans/` (e.g., `plans/auth-rewrite.md`)
- `plans/.current` contains the name of the active plan
- `plans/archive/` holds completed plans
- You read and edit plan files directly — trail only scaffolds and tracks progress

## Plan file format

```markdown
# Plan Name

## Tasks

- [ ] First task
  Description, specs, acceptance criteria.
  - [ ] Sub-step

- [x] Completed task

## Notes

Freeform context.
```

**Only top-level checkboxes** (no leading whitespace) under `## Tasks` count for progress. Indented checkboxes are sub-tasks for your own tracking.

## Commands (if `trail` binary is installed)

- `trail plan "name"` — create or select a plan
- `trail plan` — list all plans with progress
- `trail status` — show progress overview
- `trail archive [name]` — archive a completed plan

## Working with plans

1. **When a plan is active**, read it at the start of your work to understand context and remaining tasks.
2. **Check off tasks** by changing `- [ ]` to `- [x]` in the plan file as you complete them.
3. **Add notes** to the `## Notes` section as you discover important context.
4. **Don't create a new plan** if one already exists for the same topic — select it instead.

## Creating plans without the binary

If `trail` is not installed, create plans directly:

1. Create `plans/` directory at the git root if it doesn't exist
2. Write a `.md` file (e.g., `plans/my-plan.md`) using the format above
3. Write the plan name to `plans/.current` to make it active
