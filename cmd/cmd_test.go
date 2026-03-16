package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adham90/trail/internal/plan"
)

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

	plansDir := filepath.Join(dir, "plans")
	os.MkdirAll(plansDir, 0o755)

	origDir, _ := os.Getwd()
	os.Chdir(dir)

	return dir, func() {
		os.Chdir(origDir)
	}
}

func setupTestRepoWithPlan(t *testing.T) (string, func()) {
	t.Helper()
	dir, cleanup := setupTestRepo(t)

	planData := []byte(`# Test Plan

Test goal.

## Tasks

- [x] **1.** First task done
- [ ] **2.** Second task pending
- [ ] **3.** Third task todo
- [ ] **4.** Fourth task todo
`)
	planPath := filepath.Join(dir, "plans", "test-plan.md")
	if err := os.WriteFile(planPath, planData, 0o644); err != nil {
		t.Fatalf("setup: writing plan: %v", err)
	}

	plan.SetCurrent("test-plan")

	return dir, cleanup
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

func readPlanFile(t *testing.T, dir string) string {
	t.Helper()
	planPath := filepath.Join(dir, "plans", "test-plan.md")
	data, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("reading plan: %v", err)
	}
	return string(data)
}

func TestDoneCommand(t *testing.T) {
	dir, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	err := doneCmd.RunE(doneCmd, []string{"2"})
	if err != nil {
		t.Fatalf("trail done 2 failed: %v", err)
	}

	content := readPlanFile(t, dir)
	if !strings.Contains(content, "- [x] **2.** Second task pending") {
		t.Error("task 2 should be marked done")
	}
}

func TestDoneCommandOutOfRange(t *testing.T) {
	_, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	err := doneCmd.RunE(doneCmd, []string{"10"})
	if err == nil {
		t.Fatal("expected error for out-of-range task")
	}
}

func TestBlockCommand(t *testing.T) {
	dir, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	err := blockCmd.RunE(blockCmd, []string{"3", "waiting", "on", "API"})
	if err != nil {
		t.Fatalf("trail block 3 failed: %v", err)
	}

	content := readPlanFile(t, dir)
	if !strings.Contains(content, "[blocked: waiting on API]") {
		t.Error("task 3 should be blocked with reason")
	}
}

func TestStatusCommand(t *testing.T) {
	_, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	err := statusCmd.RunE(statusCmd, []string{})
	if err != nil {
		t.Fatalf("trail status failed: %v", err)
	}
}

func TestUseCommand(t *testing.T) {
	_, cleanup := setupTestRepoWithPlan(t)
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

func TestUndoCommand(t *testing.T) {
	dir, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	// Make a change
	err := doneCmd.RunE(doneCmd, []string{"2"})
	if err != nil {
		t.Fatalf("trail done 2 failed: %v", err)
	}

	content := readPlanFile(t, dir)
	if !strings.Contains(content, "- [x] **2.**") {
		t.Fatal("task 2 should be done before undo")
	}

	// Undo
	err = undoCmd.RunE(undoCmd, []string{})
	if err != nil {
		t.Fatalf("trail undo failed: %v", err)
	}

	content = readPlanFile(t, dir)
	if !strings.Contains(content, "- [ ] **2.**") {
		t.Error("task 2 should be unchecked after undo")
	}
}

func TestPlanCreateCommand(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	planNew = true
	planGoal = "Build something"
	planNoBranch = true
	defer func() {
		planNew = false
		planGoal = ""
		planNoBranch = false
	}()

	err := planCmd.RunE(planCmd, []string{"my-new-plan"})
	if err != nil {
		t.Fatalf("trail plan --new failed: %v", err)
	}

	planPath := filepath.Join(dir, "plans", "my-new-plan.md")
	data, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("plan file not created: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "# my-new-plan") {
		t.Error("plan should contain name heading")
	}
	if !strings.Contains(content, "Build something") {
		t.Error("plan should contain goal")
	}
	if !strings.Contains(content, "## Tasks") {
		t.Error("plan should contain Tasks section")
	}
}

func TestPlanListCommand(t *testing.T) {
	_, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	planNew = false
	err := planCmd.RunE(planCmd, []string{})
	if err != nil {
		t.Fatalf("trail plan (list) failed: %v", err)
	}
}

func TestPlanListEmpty(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	planNew = false
	err := planCmd.RunE(planCmd, []string{})
	if err != nil {
		t.Fatalf("trail plan (list empty) failed: %v", err)
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
