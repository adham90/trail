package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adhameldeeb/trail/internal/plan"
)

// setupTestRepo creates a temp git repo for testing.
func setupTestRepo(t *testing.T) (string, func()) {
	t.Helper()
	dir := t.TempDir()

	run(t, dir, "git", "init")
	run(t, dir, "git", "config", "user.email", "test@test.com")
	run(t, dir, "git", "config", "user.name", "Test")

	touchFile := filepath.Join(dir, ".gitkeep")
	os.WriteFile(touchFile, []byte{}, 0o644)
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "init")

	// Create plans dir
	plansDir := filepath.Join(dir, "plans")
	os.MkdirAll(plansDir, 0o755)

	origDir, _ := os.Getwd()
	os.Chdir(dir)

	return dir, func() {
		os.Chdir(origDir)
	}
}

// setupTestRepoWithPlan creates a temp repo with a named plan file.
func setupTestRepoWithPlan(t *testing.T) (string, *plan.Plan, func()) {
	t.Helper()
	dir, cleanup := setupTestRepo(t)

	p := &plan.Plan{
		Name:         "test-plan",
		Goal:         "Test plan",
		Status:       "active",
		SessionCount: 1,
		Created:      "2026-03-14",
		Updated:      "2026-03-14",
		CurrentTask:  0,
		Tasks: []plan.Task{
			{Text: "First task", Status: "done"},
			{Text: "Second task", Status: "active"},
			{Text: "Third task", Status: "todo"},
			{Text: "Fourth task", Status: "todo"},
		},
		Context: plan.Context{
			CurrentFile: "main.go",
		},
		Decisions: []plan.Decision{},
	}

	planPath := filepath.Join(dir, "plans", "test-plan.md")
	if err := plan.WriteFile(planPath, p, ""); err != nil {
		t.Fatalf("setup: writing plan: %v", err)
	}

	// Set as current plan
	plan.SetCurrent("test-plan")

	return dir, p, cleanup
}

func run(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, out)
	}
	return string(out)
}

func readPlan(t *testing.T, dir string) *plan.Plan {
	t.Helper()
	planPath := filepath.Join(dir, "plans", "test-plan.md")
	p, _, err := plan.ReadFile(planPath)
	if err != nil {
		t.Fatalf("reading plan: %v", err)
	}
	return p
}

func TestNextCommand(t *testing.T) {
	dir, _, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	err := nextCmd.RunE(nextCmd, []string{})
	if err != nil {
		t.Fatalf("trail next failed: %v", err)
	}

	p := readPlan(t, dir)
	if p.Tasks[1].Status != "done" {
		t.Errorf("Tasks[1].Status = %q, want 'done'", p.Tasks[1].Status)
	}
	if p.Tasks[2].Status != "active" {
		t.Errorf("Tasks[2].Status = %q, want 'active'", p.Tasks[2].Status)
	}
}

func TestNextSkipCommand(t *testing.T) {
	dir, _, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	nextCmd.Flags().Set("skip", "true")
	defer nextCmd.Flags().Set("skip", "false")

	err := nextCmd.RunE(nextCmd, []string{})
	if err != nil {
		t.Fatalf("trail next --skip failed: %v", err)
	}

	p := readPlan(t, dir)
	if p.Tasks[1].Status != "todo" {
		t.Errorf("Tasks[1].Status = %q, want 'todo' (skipped)", p.Tasks[1].Status)
	}
	if p.Tasks[2].Status != "active" {
		t.Errorf("Tasks[2].Status = %q, want 'active'", p.Tasks[2].Status)
	}
}

func TestCheckpointCommand(t *testing.T) {
	dir, _, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	checkpointCmd.Flags().Set("file", "new_file.go")
	checkpointCmd.Flags().Set("error", "some error")
	checkpointCmd.Flags().Set("tests", "5 passing")
	checkpointCmd.Flags().Set("note", "check this")

	err := checkpointCmd.RunE(checkpointCmd, []string{})
	if err != nil {
		t.Fatalf("trail checkpoint failed: %v", err)
	}

	p := readPlan(t, dir)
	if string(p.Context.CurrentFile) != "new_file.go" {
		t.Errorf("Context.CurrentFile = %q, want 'new_file.go'", p.Context.CurrentFile)
	}
	if string(p.Context.LastError) != "some error" {
		t.Errorf("Context.LastError = %q, want 'some error'", p.Context.LastError)
	}
	if string(p.Context.TestState) != "5 passing" {
		t.Errorf("Context.TestState = %q, want '5 passing'", p.Context.TestState)
	}
}

func TestDecideCommand(t *testing.T) {
	dir, _, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	err := decideCmd.RunE(decideCmd, []string{"Use", "Go", "over", "Ruby"})
	if err != nil {
		t.Fatalf("trail decide failed: %v", err)
	}

	p := readPlan(t, dir)
	if len(p.Decisions) != 1 {
		t.Fatalf("len(Decisions) = %d, want 1", len(p.Decisions))
	}
	if p.Decisions[0].Text != "Use Go over Ruby" {
		t.Errorf("Decision text = %q, want 'Use Go over Ruby'", p.Decisions[0].Text)
	}
}

func TestBlockActiveTask(t *testing.T) {
	dir, _, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	err := blockCmd.RunE(blockCmd, []string{"waiting on CI"})
	if err != nil {
		t.Fatalf("trail block failed: %v", err)
	}

	p := readPlan(t, dir)
	if p.Tasks[1].Status != "blocked" {
		t.Errorf("Tasks[1].Status = %q, want 'blocked'", p.Tasks[1].Status)
	}
	if p.Tasks[1].Reason != "waiting on CI" {
		t.Errorf("Tasks[1].Reason = %q, want 'waiting on CI'", p.Tasks[1].Reason)
	}
}

func TestBlockByIndex(t *testing.T) {
	dir, _, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	err := blockCmd.RunE(blockCmd, []string{"3", "not ready"})
	if err != nil {
		t.Fatalf("trail block 3 failed: %v", err)
	}

	p := readPlan(t, dir)
	if p.Tasks[3].Status != "blocked" {
		t.Errorf("Tasks[3].Status = %q, want 'blocked'", p.Tasks[3].Status)
	}
}

func TestAddCommand(t *testing.T) {
	dir, _, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	afterIdx = -1
	err := addCmd.RunE(addCmd, []string{"New", "task", "here"})
	if err != nil {
		t.Fatalf("trail add failed: %v", err)
	}

	p := readPlan(t, dir)
	if len(p.Tasks) != 5 {
		t.Fatalf("len(Tasks) = %d, want 5", len(p.Tasks))
	}
	if p.Tasks[4].Text != "New task here" {
		t.Errorf("Tasks[4].Text = %q, want 'New task here'", p.Tasks[4].Text)
	}
}

func TestAddAfterIndex(t *testing.T) {
	dir, _, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	afterIdx = 1
	defer func() { afterIdx = -1 }()

	err := addCmd.RunE(addCmd, []string{"Inserted task"})
	if err != nil {
		t.Fatalf("trail add --after 1 failed: %v", err)
	}

	p := readPlan(t, dir)
	if len(p.Tasks) != 5 {
		t.Fatalf("len(Tasks) = %d, want 5", len(p.Tasks))
	}
	if p.Tasks[2].Text != "Inserted task" {
		t.Errorf("Tasks[2].Text = %q, want 'Inserted task'", p.Tasks[2].Text)
	}
}

func TestEditCommand(t *testing.T) {
	dir, _, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	err := editCmd.RunE(editCmd, []string{"2", "Updated", "task", "text"})
	if err != nil {
		t.Fatalf("trail edit failed: %v", err)
	}

	p := readPlan(t, dir)
	if p.Tasks[2].Text != "Updated task text" {
		t.Errorf("Tasks[2].Text = %q, want 'Updated task text'", p.Tasks[2].Text)
	}
}

func TestDoneCommand(t *testing.T) {
	dir, _, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	err := doneCmd.RunE(doneCmd, []string{})
	if err != nil {
		t.Fatalf("trail done failed: %v", err)
	}

	// Original should be gone
	origPath := filepath.Join(dir, "plans", "test-plan.md")
	if _, err := os.Stat(origPath); !os.IsNotExist(err) {
		t.Error("original plan should be moved to archive")
	}

	// Should be in archive
	archivePath := filepath.Join(dir, "plans", "archive", "test-plan.md")
	p, _, err := plan.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("reading archived plan: %v", err)
	}
	if p.Status != "complete" {
		t.Errorf("archived Status = %q, want 'complete'", p.Status)
	}
}

func TestUndoCommand(t *testing.T) {
	dir, _, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	err := decideCmd.RunE(decideCmd, []string{"some decision"})
	if err != nil {
		t.Fatalf("trail decide failed: %v", err)
	}

	p := readPlan(t, dir)
	if len(p.Decisions) != 1 {
		t.Fatalf("expected 1 decision, got %d", len(p.Decisions))
	}

	err = undoCmd.RunE(undoCmd, []string{})
	if err != nil {
		t.Fatalf("trail undo failed: %v", err)
	}

	p = readPlan(t, dir)
	if len(p.Decisions) != 0 {
		t.Errorf("expected 0 decisions after undo, got %d", len(p.Decisions))
	}
}

func TestStatusCommand(t *testing.T) {
	_, _, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	err := statusCmd.RunE(statusCmd, []string{})
	if err != nil {
		t.Fatalf("trail status failed: %v", err)
	}
}

func TestCheckpointVerifyFlag(t *testing.T) {
	dir, _, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	p := readPlan(t, dir)
	p.Tasks[1].Verify = []string{"unit tests pass", "integration tests pass"}
	planPath := filepath.Join(dir, "plans", "test-plan.md")
	plan.WriteFile(planPath, p, "")

	checkpointCmd.Flags().Set("verify", "unit tests")

	err := checkpointCmd.RunE(checkpointCmd, []string{})
	if err != nil {
		t.Fatalf("trail checkpoint --verify failed: %v", err)
	}

	p = readPlan(t, dir)
	if !strings.HasPrefix(p.Tasks[1].Verify[0], "✓") {
		t.Errorf("Verify[0] should start with ✓, got %q", p.Tasks[1].Verify[0])
	}
	if strings.HasPrefix(p.Tasks[1].Verify[1], "✓") {
		t.Errorf("Verify[1] should not be marked, got %q", p.Tasks[1].Verify[1])
	}
}

func TestAddCommandWithSpec(t *testing.T) {
	dir, _, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	afterIdx = -1
	addSpec = "detailed spec here"
	addVerify = []string{"step1", "step2"}
	addFiles = []string{"file1.go"}
	defer func() {
		addSpec = ""
		addVerify = nil
		addFiles = nil
	}()

	err := addCmd.RunE(addCmd, []string{"Rich", "task"})
	if err != nil {
		t.Fatalf("trail add failed: %v", err)
	}

	p := readPlan(t, dir)
	last := p.Tasks[len(p.Tasks)-1]
	if last.Spec != "detailed spec here" {
		t.Errorf("Spec = %q, want 'detailed spec here'", last.Spec)
	}
	if len(last.Verify) != 2 {
		t.Errorf("len(Verify) = %d, want 2", len(last.Verify))
	}
}

func TestUseCommand(t *testing.T) {
	_, _, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	err := useCmd.RunE(useCmd, []string{"test-plan"})
	if err != nil {
		t.Fatalf("trail use failed: %v", err)
	}

	current, _ := plan.GetCurrent()
	if current != "test-plan" {
		t.Errorf("current plan = %q, want 'test-plan'", current)
	}
}

func TestPlanListEmpty(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	planNew = false
	err := planCmd.RunE(planCmd, []string{})
	if err != nil {
		t.Fatalf("trail plan (list) failed: %v", err)
	}
}

func TestNameToFilename(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"deploy-pipeline", "deploy-pipeline.md"},
		{"Deploy Pipeline", "deploy-pipeline.md"},
		{"feat/auth", "feat-auth.md"},
	}
	for _, tt := range tests {
		got := plan.NameToFilename(tt.name)
		if got != tt.want {
			t.Errorf("NameToFilename(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestNameToBranch(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"deploy-pipeline", "plan/deploy-pipeline"},
		{"Auth Rewrite", "plan/auth-rewrite"},
	}
	for _, tt := range tests {
		got := plan.NameToBranch(tt.name)
		if got != tt.want {
			t.Errorf("NameToBranch(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}
