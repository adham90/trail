package plan

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestNameToFilename(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"deploy-pipeline", "deploy-pipeline.md"},
		{"Deploy Pipeline", "deploy-pipeline.md"},
		{"feat/auth", "feat-auth.md"},
		{"simple", "simple.md"},
	}
	for _, tt := range tests {
		got := NameToFilename(tt.name)
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
		got := NameToBranch(tt.name)
		if got != tt.want {
			t.Errorf("NameToBranch(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestGitRoot(t *testing.T) {
	root, err := GitRoot()
	if err != nil {
		t.Fatalf("GitRoot() failed: %v", err)
	}
	if root == "" {
		t.Fatal("GitRoot() returned empty string")
	}
}

func TestCurrentBranch(t *testing.T) {
	branch, err := CurrentBranch()
	if err != nil {
		t.Fatalf("CurrentBranch() failed: %v", err)
	}
	if branch == "" {
		t.Fatal("CurrentBranch() returned empty string")
	}
}

func TestResolvePlanPath(t *testing.T) {
	path, err := ResolvePlanPath("my-plan")
	if err != nil {
		t.Fatalf("ResolvePlanPath failed: %v", err)
	}
	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got %q", path)
	}
	if filepath.Base(path) != "my-plan.md" {
		t.Errorf("expected my-plan.md, got %q", filepath.Base(path))
	}
}

// setupGitRepo creates a temp git repo and chdirs into it.
func setupGitRepo(t *testing.T) (string, func()) {
	t.Helper()
	dir := t.TempDir()

	gitRun(t, dir, "git", "init")
	gitRun(t, dir, "git", "config", "user.email", "test@test.com")
	gitRun(t, dir, "git", "config", "user.name", "Test")
	os.WriteFile(filepath.Join(dir, ".gitkeep"), []byte{}, 0o644)
	gitRun(t, dir, "git", "add", ".")
	gitRun(t, dir, "git", "commit", "-m", "init")

	os.MkdirAll(filepath.Join(dir, "plans"), 0o755)

	origDir, _ := os.Getwd()
	os.Chdir(dir)
	return dir, func() { os.Chdir(origDir) }
}

func gitRun(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}

func TestSetGetCurrentRoundTrip(t *testing.T) {
	_, cleanup := setupGitRepo(t)
	defer cleanup()

	if err := SetCurrent("my-plan"); err != nil {
		t.Fatalf("SetCurrent: %v", err)
	}

	got, err := GetCurrent()
	if err != nil {
		t.Fatalf("GetCurrent: %v", err)
	}
	if got != "my-plan" {
		t.Errorf("GetCurrent = %q, want 'my-plan'", got)
	}
}

func TestGetCurrentEmpty(t *testing.T) {
	_, cleanup := setupGitRepo(t)
	defer cleanup()

	got, err := GetCurrent()
	if err != nil {
		t.Fatalf("GetCurrent: %v", err)
	}
	if got != "" {
		t.Errorf("GetCurrent = %q, want empty", got)
	}
}

func TestListPlans(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	os.WriteFile(filepath.Join(dir, "plans", "alpha.md"), []byte("# Alpha\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "plans", "beta.md"), []byte("# Beta\n"), 0o644)

	names, err := ListPlans()
	if err != nil {
		t.Fatalf("ListPlans: %v", err)
	}
	if len(names) != 2 {
		t.Fatalf("len(names) = %d, want 2", len(names))
	}
}

func TestListPlansEmpty(t *testing.T) {
	_, cleanup := setupGitRepo(t)
	defer cleanup()

	names, err := ListPlans()
	if err != nil {
		t.Fatalf("ListPlans: %v", err)
	}
	if len(names) != 0 {
		t.Errorf("len(names) = %d, want 0", len(names))
	}
}

func TestListPlansIgnoresDirectories(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	os.WriteFile(filepath.Join(dir, "plans", "real.md"), []byte("# Real\n"), 0o644)
	os.MkdirAll(filepath.Join(dir, "plans", "archive"), 0o755)

	names, err := ListPlans()
	if err != nil {
		t.Fatalf("ListPlans: %v", err)
	}
	if len(names) != 1 {
		t.Errorf("len(names) = %d, want 1", len(names))
	}
}

func TestResolveCurrentPlanExplicitName(t *testing.T) {
	_, cleanup := setupGitRepo(t)
	defer cleanup()

	got, err := ResolveCurrentPlan("explicit")
	if err != nil {
		t.Fatalf("ResolveCurrentPlan: %v", err)
	}
	if got != "explicit" {
		t.Errorf("got %q, want 'explicit'", got)
	}
}

func TestResolveCurrentPlanFromDotCurrent(t *testing.T) {
	_, cleanup := setupGitRepo(t)
	defer cleanup()

	SetCurrent("from-current")
	got, err := ResolveCurrentPlan("")
	if err != nil {
		t.Fatalf("ResolveCurrentPlan: %v", err)
	}
	if got != "from-current" {
		t.Errorf("got %q, want 'from-current'", got)
	}
}

func TestResolveCurrentPlanSingleFallback(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	os.WriteFile(filepath.Join(dir, "plans", "only-plan.md"), []byte("# Only\n"), 0o644)

	got, err := ResolveCurrentPlan("")
	if err != nil {
		t.Fatalf("ResolveCurrentPlan: %v", err)
	}
	if got != "only-plan" {
		t.Errorf("got %q, want 'only-plan'", got)
	}
}

func TestResolveCurrentPlanNoPlans(t *testing.T) {
	_, cleanup := setupGitRepo(t)
	defer cleanup()

	_, err := ResolveCurrentPlan("")
	if err == nil {
		t.Fatal("expected error when no plans exist")
	}
}

func TestResolveCurrentPlanMultipleNoDefault(t *testing.T) {
	dir, cleanup := setupGitRepo(t)
	defer cleanup()

	os.WriteFile(filepath.Join(dir, "plans", "a.md"), []byte("# A\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "plans", "b.md"), []byte("# B\n"), 0o644)

	_, err := ResolveCurrentPlan("")
	if err == nil {
		t.Fatal("expected error when multiple plans and no .current")
	}
}
