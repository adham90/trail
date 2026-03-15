package renderer

import (
	"strings"
	"testing"
	"time"

	"github.com/adham90/trail/internal/plan"
)

func TestStatusSymbol(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{"done", "✓"},
		{"active", "▶"},
		{"blocked", "!"},
		{"todo", "○"},
		{"unknown", "?"},
	}
	for _, tt := range tests {
		got := StatusSymbol(tt.status)
		if got != tt.want {
			t.Errorf("StatusSymbol(%q) = %q, want %q", tt.status, got, tt.want)
		}
	}
}

func TestSummary(t *testing.T) {
	p := &plan.Plan{
		Name:         "test",
		Goal:         "Build trail",
		Branch:       "main",
		Status:       "active",
		SessionCount: 2,
		Tasks: []plan.Task{
			{Text: "Init module", Status: "done"},
			{Text: "Add deps", Status: "done"},
			{Text: "Wire CLI", Status: "active"},
			{Text: "Deploy config", Status: "blocked", Reason: "waiting on DevOps"},
			{Text: "Write tests", Status: "todo"},
		},
		Context: plan.Context{
			CurrentFile: "main.go",
			LastError:   "",
			TestState:   "3 passing",
		},
		Decisions: []plan.Decision{
			{Time: time.Now(), Text: "Go over Ruby"},
			{Time: time.Now(), Text: "YAML frontmatter"},
		},
	}

	output := Summary(p)

	// Check header
	if !strings.Contains(output, "Build trail") {
		t.Error("summary missing goal")
	}
	if !strings.Contains(output, "main") {
		t.Error("summary missing branch")
	}

	// Blocked tasks should appear at top
	lines := strings.Split(output, "\n")
	blockedLineIdx := -1
	firstTaskLineIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "!") && blockedLineIdx < 0 {
			blockedLineIdx = i
		}
		if (strings.HasPrefix(line, "✓") || strings.HasPrefix(line, "▶") || strings.HasPrefix(line, "○")) && firstTaskLineIdx < 0 {
			firstTaskLineIdx = i
		}
	}
	if blockedLineIdx < 0 {
		t.Error("no blocked task line found")
	}
	if firstTaskLineIdx < 0 {
		t.Error("no task list found")
	}
	if blockedLineIdx > firstTaskLineIdx {
		t.Error("blocked tasks should appear before the main task list")
	}

	// Context
	if !strings.Contains(output, "current_file:  main.go") {
		t.Error("summary missing current_file")
	}
	if !strings.Contains(output, "last_error:    ~") {
		t.Error("empty last_error should display as ~")
	}

	// Decisions count
	if !strings.Contains(output, "2 logged") {
		t.Error("summary missing decisions count")
	}
}

func TestSummaryNoBlockedTasks(t *testing.T) {
	p := &plan.Plan{
		Name:         "test",
		Goal:         "Simple plan",
		Branch:       "main",
		SessionCount: 1,
		Tasks: []plan.Task{
			{Text: "First", Status: "active"},
			{Text: "Second", Status: "todo"},
		},
		Context:   plan.Context{},
		Decisions: []plan.Decision{},
	}

	output := Summary(p)
	if strings.Contains(output, "blocked:") {
		t.Error("should not contain blocked section when no tasks are blocked")
	}
}

func TestContextBlock(t *testing.T) {
	completed := &plan.Task{Text: "Add DeployJob", Status: "done"}
	active := &plan.Task{Text: "Define endpoint", Status: "active"}
	ctx := plan.Context{
		CurrentFile: "app/controllers/deploys.rb",
		TestState:   "2 passing",
	}

	output := ContextBlock(completed, active, ctx)

	if !strings.Contains(output, "✓ Add DeployJob") {
		t.Error("missing completed task")
	}
	if !strings.Contains(output, "▶ Define endpoint") {
		t.Error("missing active task")
	}
	if !strings.Contains(output, "current_file:  app/controllers/deploys.rb") {
		t.Error("missing current_file")
	}
	if !strings.Contains(output, "last_error:    ~") {
		t.Error("empty last_error should show ~")
	}
}

func TestContextBlockNoCompleted(t *testing.T) {
	active := &plan.Task{Text: "First task", Status: "active"}
	ctx := plan.Context{}

	output := ContextBlock(nil, active, ctx)

	if strings.Contains(output, "✓") {
		t.Error("should not show completed symbol when no completed task")
	}
	if !strings.Contains(output, "▶ First task") {
		t.Error("missing active task")
	}
}

func TestSummaryWithConstraintsAndFiles(t *testing.T) {
	p := &plan.Plan{
		Name:         "test",
		Goal:         "Rich plan",
		Branch:       "main",
		SessionCount: 1,
		Constraints:  []string{"no API", "atomic writes"},
		PlanFiles:    []plan.FileRef{{Path: "a.go", Role: "core"}},
		Tasks:        []plan.Task{{Text: "Task", Status: "active"}},
		Context:      plan.Context{},
		Decisions:    []plan.Decision{},
	}

	output := Summary(p)
	if !strings.Contains(output, "2 defined") {
		t.Error("missing constraints count")
	}
	if !strings.Contains(output, "1 tracked") {
		t.Error("missing files count")
	}
}

func TestSummaryNoConstraints(t *testing.T) {
	p := &plan.Plan{
		Name:         "test",
		Goal:         "Simple",
		Branch:       "main",
		SessionCount: 1,
		Tasks:        []plan.Task{{Text: "Task", Status: "todo"}},
		Context:      plan.Context{},
		Decisions:    []plan.Decision{},
	}

	output := Summary(p)
	if strings.Contains(output, "constraints:") {
		t.Error("should not show constraints line when none defined")
	}
	if strings.Contains(output, "files:") {
		t.Error("should not show files line when none tracked")
	}
}

func TestContextBlockWithSpecAndVerify(t *testing.T) {
	active := &plan.Task{
		Text:   "Define endpoint",
		Status: "active",
		Spec:   "REST endpoint for deploys",
		Verify: []string{"returns 200", "creates record"},
	}
	ctx := plan.Context{CurrentFile: "app.go"}

	output := ContextBlock(nil, active, ctx)

	if !strings.Contains(output, "spec: REST endpoint for deploys") {
		t.Error("missing spec in context block")
	}
	if !strings.Contains(output, "returns 200") {
		t.Error("missing verify step in context block")
	}
	if !strings.Contains(output, "creates record") {
		t.Error("missing second verify step")
	}
}

func TestContextBlockNoSpec(t *testing.T) {
	active := &plan.Task{Text: "Simple task", Status: "active"}
	ctx := plan.Context{}

	output := ContextBlock(nil, active, ctx)
	if strings.Contains(output, "spec:") {
		t.Error("should not show spec when empty")
	}
	if strings.Contains(output, "verify:") {
		t.Error("should not show verify when empty")
	}
}
