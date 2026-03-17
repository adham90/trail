package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adham90/trail/internal/plan"
)

// --- Test helpers ---

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
  - [ ] 2.1. verify step one
  - [ ] 2.2. verify step two
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

// --- plan command ---

func TestPlanCreate(t *testing.T) {
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
	if !strings.Contains(content, "## Acceptance Criteria") {
		t.Error("plan should contain Acceptance Criteria section")
	}

	// Should be set as current
	current, _ := plan.GetCurrent()
	if current != "my-new-plan" {
		t.Errorf("current plan = %q, want 'my-new-plan'", current)
	}
}

func TestPlanCreateRequiresGoal(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	planNew = true
	planGoal = ""
	planNoBranch = true
	defer func() {
		planNew = false
		planGoal = ""
		planNoBranch = false
	}()

	err := planCmd.RunE(planCmd, []string{"no-goal-plan"})
	if err == nil {
		t.Fatal("expected error when --goal is missing")
	}
}

func TestPlanCreateWithBranch(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	planNew = true
	planGoal = "Test branching"
	planNoBranch = false
	defer func() {
		planNew = false
		planGoal = ""
		planNoBranch = false
	}()

	err := planCmd.RunE(planCmd, []string{"branch-test"})
	if err != nil {
		t.Fatalf("trail plan --new with branch failed: %v", err)
	}

	branch, _ := plan.CurrentBranch()
	if branch != "plan/branch-test" {
		t.Errorf("branch = %q, want 'plan/branch-test'", branch)
	}
}

func TestPlanCreateDuplicate(t *testing.T) {
	_, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	planNew = true
	planGoal = "Duplicate"
	planNoBranch = true
	defer func() {
		planNew = false
		planGoal = ""
		planNoBranch = false
	}()

	err := planCmd.RunE(planCmd, []string{"test-plan"})
	if err == nil {
		t.Fatal("expected error for duplicate plan")
	}
}

func TestPlanList(t *testing.T) {
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

func TestPlanSelectByName(t *testing.T) {
	_, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	plan.SetCurrent("")
	planNew = false
	err := planCmd.RunE(planCmd, []string{"test-plan"})
	if err != nil {
		t.Fatalf("trail plan <name> failed: %v", err)
	}

	current, _ := plan.GetCurrent()
	if current != "test-plan" {
		t.Errorf("current = %q, want 'test-plan'", current)
	}
}

func TestPlanSelectNotFound(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	planNew = false
	err := planCmd.RunE(planCmd, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent plan")
	}
}

// --- done command ---

func TestDone(t *testing.T) {
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
	// Task 3 should still be unchecked
	if !strings.Contains(content, "- [ ] **3.** Third task todo") {
		t.Error("task 3 should still be unchecked")
	}
}

func TestDoneCreatesBackup(t *testing.T) {
	dir, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	doneCmd.RunE(doneCmd, []string{"2"})

	backupPath := filepath.Join(dir, "plans", ".backup")
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("backup file should exist after done")
	}
}

func TestDoneOutOfRange(t *testing.T) {
	_, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	err := doneCmd.RunE(doneCmd, []string{"10"})
	if err == nil {
		t.Fatal("expected error for out-of-range task")
	}
}

func TestDoneInvalidArg(t *testing.T) {
	_, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	err := doneCmd.RunE(doneCmd, []string{"abc"})
	if err == nil {
		t.Fatal("expected error for non-numeric arg")
	}
}

func TestDoneAlreadyDone(t *testing.T) {
	_, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	// Task 1 is already [x] — should be a no-op (no error)
	err := doneCmd.RunE(doneCmd, []string{"1"})
	if err != nil {
		t.Fatalf("trail done on already-done task failed: %v", err)
	}
}

func TestDoneSubTask(t *testing.T) {
	dir, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	err := doneCmd.RunE(doneCmd, []string{"2.1"})
	if err != nil {
		t.Fatalf("trail done 2.1 failed: %v", err)
	}

	content := readPlanFile(t, dir)
	if !strings.Contains(content, "- [x] 2.1. verify step one") {
		t.Error("sub-task 2.1 should be marked done")
	}
	// Other sub-task unchanged
	if !strings.Contains(content, "- [ ] 2.2. verify step two") {
		t.Error("sub-task 2.2 should still be unchecked")
	}
	// Parent task unchanged
	if !strings.Contains(content, "- [ ] **2.** Second task pending") {
		t.Error("parent task 2 should still be unchecked")
	}
}

func TestDoneSubTaskNotFound(t *testing.T) {
	_, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	err := doneCmd.RunE(doneCmd, []string{"2.9"})
	if err == nil {
		t.Fatal("expected error for nonexistent sub-task")
	}
}

func TestDoneMarksSubTasksToo(t *testing.T) {
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
	if !strings.Contains(content, "- [x] 2.1. verify step one") {
		t.Error("sub-task 2.1 should be marked done")
	}
	if !strings.Contains(content, "- [x] 2.2. verify step two") {
		t.Error("sub-task 2.2 should be marked done")
	}
	// Task 3 should be unchanged
	if !strings.Contains(content, "- [ ] **3.** Third task todo") {
		t.Error("task 3 should still be unchecked")
	}
}

// --- block command ---

func TestBlock(t *testing.T) {
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

func TestBlockCreatesBackup(t *testing.T) {
	dir, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	blockCmd.RunE(blockCmd, []string{"2", "reason"})

	backupPath := filepath.Join(dir, "plans", ".backup")
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("backup file should exist after block")
	}
}

func TestBlockOutOfRange(t *testing.T) {
	_, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	err := blockCmd.RunE(blockCmd, []string{"10", "reason"})
	if err == nil {
		t.Fatal("expected error for out-of-range task")
	}
}

func TestBlockInvalidArg(t *testing.T) {
	_, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	err := blockCmd.RunE(blockCmd, []string{"abc", "reason"})
	if err == nil {
		t.Fatal("expected error for non-numeric arg")
	}
}

// --- status command ---

func TestStatus(t *testing.T) {
	_, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	err := statusCmd.RunE(statusCmd, []string{})
	if err != nil {
		t.Fatalf("trail status failed: %v", err)
	}
}

func TestStatusEmpty(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	err := statusCmd.RunE(statusCmd, []string{})
	if err != nil {
		t.Fatalf("trail status (empty) failed: %v", err)
	}
}

func TestStatusMultiplePlans(t *testing.T) {
	dir, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	// Add a second plan
	secondPlan := []byte("# Second Plan\n\nGoal.\n\n## Tasks\n\n- [x] **1.** Done\n- [ ] **2.** Todo\n")
	os.WriteFile(filepath.Join(dir, "plans", "second-plan.md"), secondPlan, 0o644)

	err := statusCmd.RunE(statusCmd, []string{})
	if err != nil {
		t.Fatalf("trail status (multiple) failed: %v", err)
	}
}

// --- use command ---

func TestUse(t *testing.T) {
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

func TestUseNotFound(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	err := useCmd.RunE(useCmd, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent plan")
	}
}

func TestUseSwitchesBranch(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create a plan with a branch
	planNew = true
	planGoal = "Test branch switch"
	planNoBranch = false
	defer func() {
		planNew = false
		planGoal = ""
		planNoBranch = false
	}()

	planCmd.RunE(planCmd, []string{"branched-plan"})

	// Switch back to main
	plan.SwitchBranch("main")

	// Use should switch branch
	err := useCmd.RunE(useCmd, []string{"branched-plan"})
	if err != nil {
		t.Fatalf("trail use with branch failed: %v", err)
	}

	branch, _ := plan.CurrentBranch()
	if branch != "plan/branched-plan" {
		t.Errorf("branch = %q, want 'plan/branched-plan'", branch)
	}
}

// --- resume command ---

func TestResume(t *testing.T) {
	_, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	err := resumeCmd.RunE(resumeCmd, []string{})
	if err != nil {
		t.Fatalf("trail resume failed: %v", err)
	}
}

func TestResumeByName(t *testing.T) {
	_, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	err := resumeCmd.RunE(resumeCmd, []string{"test-plan"})
	if err != nil {
		t.Fatalf("trail resume <name> failed: %v", err)
	}
}

func TestResumeNotFound(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	plan.SetCurrent("nonexistent")
	err := resumeCmd.RunE(resumeCmd, []string{})
	if err == nil {
		t.Fatal("expected error for nonexistent plan")
	}
}

// --- undo command ---

func TestUndo(t *testing.T) {
	dir, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	doneCmd.RunE(doneCmd, []string{"2"})

	content := readPlanFile(t, dir)
	if !strings.Contains(content, "- [x] **2.**") {
		t.Fatal("task 2 should be done before undo")
	}

	err := undoCmd.RunE(undoCmd, []string{})
	if err != nil {
		t.Fatalf("trail undo failed: %v", err)
	}

	content = readPlanFile(t, dir)
	if !strings.Contains(content, "- [ ] **2.**") {
		t.Error("task 2 should be unchecked after undo")
	}
}

func TestUndoNoBackup(t *testing.T) {
	_, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	err := undoCmd.RunE(undoCmd, []string{})
	if err == nil {
		t.Fatal("expected error when no backup exists")
	}
}

// --- archive command ---

func TestArchive(t *testing.T) {
	dir, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	err := archiveCmd.RunE(archiveCmd, []string{})
	if err != nil {
		t.Fatalf("trail archive failed: %v", err)
	}

	// Original should be gone
	origPath := filepath.Join(dir, "plans", "test-plan.md")
	if _, err := os.Stat(origPath); !os.IsNotExist(err) {
		t.Error("original plan should be moved to archive")
	}

	// Should be in archive
	archivePath := filepath.Join(dir, "plans", "archive", "test-plan.md")
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Error("plan should exist in archive/")
	}
}

func TestArchiveClearsCurrent(t *testing.T) {
	_, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	current, _ := plan.GetCurrent()
	if current != "test-plan" {
		t.Fatalf("current = %q before archive, want 'test-plan'", current)
	}

	archiveCmd.RunE(archiveCmd, []string{})

	current, _ = plan.GetCurrent()
	if current != "" {
		t.Errorf("current = %q after archive, want empty", current)
	}
}

func TestArchiveByName(t *testing.T) {
	dir, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	plan.SetCurrent("") // clear current so it resolves by name

	err := archiveCmd.RunE(archiveCmd, []string{"test-plan"})
	if err != nil {
		t.Fatalf("trail archive <name> failed: %v", err)
	}

	archivePath := filepath.Join(dir, "plans", "archive", "test-plan.md")
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Error("plan should exist in archive/")
	}
}

func TestArchiveNotFound(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	plan.SetCurrent("nonexistent")
	err := archiveCmd.RunE(archiveCmd, []string{})
	if err == nil {
		t.Fatal("expected error for nonexistent plan")
	}
}

// --- prompt command ---

func TestPrompt(t *testing.T) {
	// prompt uses Run not RunE, just verify it doesn't panic
	promptCmd.Run(promptCmd, []string{})
}

// --- helpers ---

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
