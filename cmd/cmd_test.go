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

- [x] First task done
- [ ] Second task pending
  - [ ] verify step one
  - [ ] verify step two
- [ ] Third task todo
- [ ] Fourth task todo
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

// --- plan command ---

func TestPlanCreate(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	err := planCmd.RunE(planCmd, []string{"my-new-plan"})
	if err != nil {
		t.Fatalf("trail plan <name> (create) failed: %v", err)
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
	if !strings.Contains(content, "## Tasks") {
		t.Error("plan should contain Tasks section")
	}

	// Should be set as current
	current, _ := plan.GetCurrent()
	if current != "my-new-plan" {
		t.Errorf("current plan = %q, want 'my-new-plan'", current)
	}
}

func TestPlanCreateSetsAsCurrent(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	planCmd.RunE(planCmd, []string{"first-plan"})
	planCmd.RunE(planCmd, []string{"second-plan"})

	current, _ := plan.GetCurrent()
	if current != "second-plan" {
		t.Errorf("current = %q, want 'second-plan'", current)
	}
}

func TestPlanSelectExisting(t *testing.T) {
	_, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	plan.SetCurrent("")

	err := planCmd.RunE(planCmd, []string{"test-plan"})
	if err != nil {
		t.Fatalf("trail plan <name> (select) failed: %v", err)
	}

	current, _ := plan.GetCurrent()
	if current != "test-plan" {
		t.Errorf("current = %q, want 'test-plan'", current)
	}
}

func TestPlanList(t *testing.T) {
	_, cleanup := setupTestRepoWithPlan(t)
	defer cleanup()

	err := planCmd.RunE(planCmd, []string{})
	if err != nil {
		t.Fatalf("trail plan (list) failed: %v", err)
	}
}

func TestPlanListEmpty(t *testing.T) {
	_, cleanup := setupTestRepo(t)
	defer cleanup()

	err := planCmd.RunE(planCmd, []string{})
	if err != nil {
		t.Fatalf("trail plan (list empty) failed: %v", err)
	}
}

func TestPlanListShowsCurrentMarker(t *testing.T) {
	dir, cleanup := setupTestRepo(t)
	defer cleanup()

	// Create two plans
	os.WriteFile(filepath.Join(dir, "plans", "alpha.md"),
		[]byte("# Alpha\n\n## Tasks\n\n- [ ] task\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "plans", "beta.md"),
		[]byte("# Beta\n\n## Tasks\n\n- [x] task\n"), 0o644)
	plan.SetCurrent("alpha")

	// Just verify it runs without error (output goes to stdout)
	err := planCmd.RunE(planCmd, []string{})
	if err != nil {
		t.Fatalf("trail plan (list with current) failed: %v", err)
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

	secondPlan := []byte("# Second Plan\n\nGoal.\n\n## Tasks\n\n- [x] Done\n- [ ] Todo\n")
	os.WriteFile(filepath.Join(dir, "plans", "second-plan.md"), secondPlan, 0o644)

	err := statusCmd.RunE(statusCmd, []string{})
	if err != nil {
		t.Fatalf("trail status (multiple) failed: %v", err)
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
