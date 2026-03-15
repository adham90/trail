package plan

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	frontmatterSep = []byte("---\n")
)

// Parse splits a plan file into YAML frontmatter and freeform notes,
// then unmarshals the frontmatter into a Plan struct.
func Parse(data []byte) (*Plan, string, error) {
	// Must start with ---
	if !bytes.HasPrefix(data, frontmatterSep) {
		return nil, "", fmt.Errorf("plan file missing frontmatter (must start with ---)")
	}

	// Find the closing ---
	rest := data[len(frontmatterSep):]
	idx := bytes.Index(rest, frontmatterSep)
	if idx < 0 {
		return nil, "", fmt.Errorf("plan file missing closing frontmatter separator (---)")
	}

	yamlData := rest[:idx]
	body := string(rest[idx+len(frontmatterSep):])

	// Extract only the ## notes section from the body.
	// Everything else (tasks, context, decisions) is generated and will be regenerated on write.
	notes := extractNotes(body)

	var p Plan
	if err := yaml.Unmarshal(yamlData, &p); err != nil {
		return nil, "", fmt.Errorf("parsing frontmatter YAML: %w", err)
	}

	return &p, notes, nil
}

// ReadFile reads a plan file from disk and parses it.
func ReadFile(path string) (*Plan, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	return Parse(data)
}

// extractNotes pulls out the content after "## notes" heading.
// Returns empty string if no notes section or if it's empty.
func extractNotes(body string) string {
	idx := strings.LastIndex(body, "## notes")
	if idx < 0 {
		return ""
	}

	// Get everything after "## notes\n"
	after := body[idx+len("## notes"):]
	after = strings.TrimPrefix(after, "\n")

	// Trim leading/trailing whitespace
	after = strings.TrimSpace(after)
	if after == "" {
		return ""
	}
	return after
}

func isBlank(s string) bool {
	for _, c := range s {
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			return false
		}
	}
	return true
}
