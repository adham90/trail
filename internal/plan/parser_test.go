package plan

import (
	"testing"
	"time"
)

func TestParsePlan(t *testing.T) {
	input := `---
name: build-trail
goal: Build trail
branch: main
status: active
session_count: 1
created: "2026-03-14"
updated: "2026-03-14"
current_task: 0
tasks:
  - text: Init Go module
    status: done
  - text: Add dependencies
    status: active
  - text: Wire root command
    status: todo
  - text: Deploy config
    status: blocked
    reason: waiting on DevOps
context:
  current_file: main.go
  last_error: "~"
  test_state: 2 passing
  open_questions: "~"
  pending_refactor: "~"
decisions:
  - time: 2026-03-14T10:00:00Z
    text: Go over Ruby
---

## notes

Some freeform notes here.
More notes.
`

	p, notes, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Check top-level fields
	if p.Goal != "Build trail" {
		t.Errorf("Goal = %q, want %q", p.Goal, "Build trail")
	}
	if p.Branch != "main" {
		t.Errorf("Branch = %q, want %q", p.Branch, "main")
	}
	if p.Status != "active" {
		t.Errorf("Status = %q, want %q", p.Status, "active")
	}
	if p.SessionCount != 1 {
		t.Errorf("SessionCount = %d, want 1", p.SessionCount)
	}
	if p.CurrentTask != 0 {
		t.Errorf("CurrentTask = %d, want 0", p.CurrentTask)
	}

	// Check tasks
	if len(p.Tasks) != 4 {
		t.Fatalf("len(Tasks) = %d, want 4", len(p.Tasks))
	}
	if p.Tasks[0].Status != "done" {
		t.Errorf("Tasks[0].Status = %q, want %q", p.Tasks[0].Status, "done")
	}
	if p.Tasks[1].Status != "active" {
		t.Errorf("Tasks[1].Status = %q, want %q", p.Tasks[1].Status, "active")
	}
	if p.Tasks[3].Status != "blocked" {
		t.Errorf("Tasks[3].Status = %q, want %q", p.Tasks[3].Status, "blocked")
	}
	if p.Tasks[3].Reason != "waiting on DevOps" {
		t.Errorf("Tasks[3].Reason = %q, want %q", p.Tasks[3].Reason, "waiting on DevOps")
	}

	// Check context
	if p.Context.CurrentFile != "main.go" {
		t.Errorf("Context.CurrentFile = %q, want %q", string(p.Context.CurrentFile), "main.go")
	}
	if p.Context.TestState != "2 passing" {
		t.Errorf("Context.TestState = %q, want %q", string(p.Context.TestState), "2 passing")
	}

	// Check decisions
	if len(p.Decisions) != 1 {
		t.Fatalf("len(Decisions) = %d, want 1", len(p.Decisions))
	}
	if p.Decisions[0].Text != "Go over Ruby" {
		t.Errorf("Decisions[0].Text = %q, want %q", p.Decisions[0].Text, "Go over Ruby")
	}
	expectedTime := time.Date(2026, 3, 14, 10, 0, 0, 0, time.UTC)
	if !p.Decisions[0].Time.Equal(expectedTime) {
		t.Errorf("Decisions[0].Time = %v, want %v", p.Decisions[0].Time, expectedTime)
	}

	// Check notes — parser extracts only content after "## notes"
	expectedNotes := "Some freeform notes here.\nMore notes."
	if notes != expectedNotes {
		t.Errorf("notes = %q, want %q", notes, expectedNotes)
	}
}

func TestParsePlanNoNotes(t *testing.T) {
	input := `---
name: minimal
goal: Minimal plan
branch: main
status: active
session_count: 1
created: "2026-03-14"
updated: "2026-03-14"
current_task: 0
tasks: []
context:
  current_file: "~"
  last_error: "~"
  test_state: "~"
  open_questions: "~"
  pending_refactor: "~"
decisions: []
---
`
	p, notes, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if p.Goal != "Minimal plan" {
		t.Errorf("Goal = %q, want %q", p.Goal, "Minimal plan")
	}
	if notes != "" {
		t.Errorf("notes = %q, want empty", notes)
	}
}

func TestParsePlanInvalidYAML(t *testing.T) {
	input := `---
goal: [broken
---
`
	_, _, err := Parse([]byte(input))
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestParsePlanNoFrontmatter(t *testing.T) {
	input := `just some text without frontmatter`
	_, _, err := Parse([]byte(input))
	if err == nil {
		t.Fatal("expected error for missing frontmatter, got nil")
	}
}

func TestParsePlanRichFields(t *testing.T) {
	input := `---
name: rich-plan
goal: Rich plan
branch: main
status: active
session_count: 1
created: "2026-03-14"
updated: "2026-03-14"
current_task: 1
constraints:
  - No external API calls
  - All writes must be atomic
files:
  - path: internal/plan/model.go
    role: core model
  - path: cmd/checkpoint.go
    role: command handler
tasks:
  - text: Setup
    status: done
  - text: Implement feature
    status: active
    spec: Add rich plan fields to model and writer
    verify:
      - unit tests pass
      - round-trip serialize works
    files:
      - internal/plan/model.go
      - internal/plan/writer.go
  - text: Deploy
    status: todo
context:
  current_file: model.go
  last_error: "~"
  test_state: "~"
  open_questions: "~"
  pending_refactor: "~"
decisions: []
---
`
	p, _, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Plan-level constraints
	if len(p.Constraints) != 2 {
		t.Fatalf("len(Constraints) = %d, want 2", len(p.Constraints))
	}
	if p.Constraints[0] != "No external API calls" {
		t.Errorf("Constraints[0] = %q", p.Constraints[0])
	}

	// Plan-level files
	if len(p.PlanFiles) != 2 {
		t.Fatalf("len(PlanFiles) = %d, want 2", len(p.PlanFiles))
	}
	if p.PlanFiles[0].Path != "internal/plan/model.go" {
		t.Errorf("PlanFiles[0].Path = %q", p.PlanFiles[0].Path)
	}
	if p.PlanFiles[0].Role != "core model" {
		t.Errorf("PlanFiles[0].Role = %q", p.PlanFiles[0].Role)
	}

	// Task-level spec
	if p.Tasks[1].Spec != "Add rich plan fields to model and writer" {
		t.Errorf("Tasks[1].Spec = %q", p.Tasks[1].Spec)
	}

	// Task-level verify
	if len(p.Tasks[1].Verify) != 2 {
		t.Fatalf("len(Tasks[1].Verify) = %d, want 2", len(p.Tasks[1].Verify))
	}
	if p.Tasks[1].Verify[0] != "unit tests pass" {
		t.Errorf("Tasks[1].Verify[0] = %q", p.Tasks[1].Verify[0])
	}

	// Task-level files
	if len(p.Tasks[1].Files) != 2 {
		t.Fatalf("len(Tasks[1].Files) = %d, want 2", len(p.Tasks[1].Files))
	}

	// Task without rich fields — zero values
	if p.Tasks[0].Spec != "" {
		t.Errorf("Tasks[0].Spec should be empty, got %q", p.Tasks[0].Spec)
	}
	if p.Tasks[0].Verify != nil {
		t.Errorf("Tasks[0].Verify should be nil, got %v", p.Tasks[0].Verify)
	}
	if p.Tasks[0].Files != nil {
		t.Errorf("Tasks[0].Files should be nil, got %v", p.Tasks[0].Files)
	}
}

func TestParsePlanBackwardCompatible(t *testing.T) {
	// Plan without new fields should still parse; new fields are zero-valued
	input := `---
name: old-plan
goal: Old plan
branch: main
status: active
session_count: 1
created: "2026-03-14"
updated: "2026-03-14"
current_task: 0
tasks:
  - text: A task
    status: todo
context:
  current_file: "~"
  last_error: "~"
  test_state: "~"
  open_questions: "~"
  pending_refactor: "~"
decisions: []
---
`
	p, _, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if p.Constraints != nil {
		t.Errorf("Constraints should be nil, got %v", p.Constraints)
	}
	if p.PlanFiles != nil {
		t.Errorf("PlanFiles should be nil, got %v", p.PlanFiles)
	}
	if p.Tasks[0].Spec != "" {
		t.Errorf("Tasks[0].Spec should be empty, got %q", p.Tasks[0].Spec)
	}
}

func TestParseTildeAsEmpty(t *testing.T) {
	input := `---
name: test-tildes
goal: Test tildes
branch: main
status: active
session_count: 1
created: "2026-03-14"
updated: "2026-03-14"
current_task: 0
tasks: []
context:
  current_file: ~
  last_error: ~
  test_state: ~
  open_questions: ~
  pending_refactor: ~
decisions: []
---
`
	p, _, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	// YAML ~ maps to empty NullableString
	if p.Context.CurrentFile != "" {
		t.Errorf("Context.CurrentFile = %q, want empty (from ~)", string(p.Context.CurrentFile))
	}
}
