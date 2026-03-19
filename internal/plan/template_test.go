package plan

import (
	"strings"
	"testing"
)

func TestGenerateTemplate(t *testing.T) {
	data := GenerateTemplate("Deploy Pipeline")
	s := string(data)

	if !strings.Contains(s, "# Deploy Pipeline") {
		t.Error("template should contain plan name heading")
	}
	if !strings.Contains(s, "## Tasks") {
		t.Error("template should contain Tasks section")
	}
	if !strings.Contains(s, "- [ ] Define tasks") {
		t.Error("template should contain placeholder task")
	}

	// Should be valid for parsing
	done, total := ParseTaskCounts(data)
	if total != 1 {
		t.Errorf("template should have 1 placeholder task, got %d", total)
	}
	if done != 0 {
		t.Errorf("template should have 0 done tasks, got %d", done)
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
