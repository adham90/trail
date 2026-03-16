package plan

import (
	"strings"
	"testing"
)

func TestGenerateTemplate(t *testing.T) {
	data := GenerateTemplate("Deploy Pipeline", "Build and deploy the CI/CD pipeline")

	s := string(data)

	if !strings.Contains(s, "# Deploy Pipeline") {
		t.Error("template should contain plan name heading")
	}
	if !strings.Contains(s, "Build and deploy the CI/CD pipeline") {
		t.Error("template should contain goal")
	}
	if !strings.Contains(s, "## Tasks") {
		t.Error("template should contain Tasks section")
	}
	if !strings.Contains(s, "## Acceptance Criteria") {
		t.Error("template should contain Acceptance Criteria section")
	}
	if !strings.Contains(s, "- [ ] **1.**") {
		t.Error("template should contain numbered task placeholders")
	}

	// Should be valid for parsing
	tasks := ParseTasks(data)
	if len(tasks) != 3 {
		t.Errorf("template should have 3 placeholder tasks, got %d", len(tasks))
	}
}

func TestSlugToTitle(t *testing.T) {
	tests := []struct {
		slug string
		want string
	}{
		{"deploy-pipeline", "Deploy Pipeline"},
		{"auth", "Auth"},
		{"api-gateway-rewrite", "Api Gateway Rewrite"},
	}
	for _, tt := range tests {
		got := SlugToTitle(tt.slug)
		if got != tt.want {
			t.Errorf("SlugToTitle(%q) = %q, want %q", tt.slug, got, tt.want)
		}
	}
}
