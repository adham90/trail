package plan

import (
	"os"
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

func TestSetGetCurrent(t *testing.T) {
	dir := t.TempDir()
	plansDir := filepath.Join(dir, "plans")
	os.MkdirAll(plansDir, 0o755)

	currentPath := filepath.Join(plansDir, ".current")
	os.WriteFile(currentPath, []byte("my-plan\n"), 0o644)
	data, err := os.ReadFile(currentPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(data) != "my-plan\n" {
		t.Errorf("current file = %q, want 'my-plan\\n'", string(data))
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
