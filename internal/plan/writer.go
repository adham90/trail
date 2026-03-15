package plan

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Serialize converts a Plan and notes string into the file format:
// ---\n<yaml>\n---\n<rendered markdown>
func Serialize(p *Plan, notes string) ([]byte, error) {
	yamlData, err := yaml.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("marshaling plan YAML: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(yamlData)
	buf.WriteString("---\n")
	buf.WriteString("\n<!-- generated below — do not edit, use trail commands -->\n")

	// Render constraints (if any)
	if len(p.Constraints) > 0 {
		buf.WriteString("\n## constraints\n\n")
		for _, c := range p.Constraints {
			buf.WriteString(fmt.Sprintf("- %s\n", c))
		}
	}

	// Render plan-level files (if any)
	if len(p.PlanFiles) > 0 {
		buf.WriteString("\n## files\n\n")
		buf.WriteString("| path | role |\n")
		buf.WriteString("|---|---|\n")
		for _, f := range p.PlanFiles {
			buf.WriteString(fmt.Sprintf("| %s | %s |\n", f.Path, f.Role))
		}
	}

	// Render tasks
	buf.WriteString("\n## tasks\n\n")
	if len(p.Tasks) == 0 {
		buf.WriteString("No tasks yet.\n")
	}
	for i, t := range p.Tasks {
		buf.WriteString(fmt.Sprintf("- [%s] %02d · %s", taskCheckbox(t.Status), i, t.Text))
		if t.Status == "blocked" && t.Reason != "" {
			buf.WriteString(fmt.Sprintf(" — blocked: %s", t.Reason))
		}
		buf.WriteString("\n")

		// Expand active task with spec/verify/files
		if t.Status == "active" {
			if t.Spec != "" {
				buf.WriteString(fmt.Sprintf("\n  **spec:** %s\n", t.Spec))
			}
			if len(t.Verify) > 0 {
				buf.WriteString("\n  **verify:**\n")
				for _, v := range t.Verify {
					if isVerifyPassed(v) {
						buf.WriteString(fmt.Sprintf("  - [x] %s\n", v))
					} else {
						buf.WriteString(fmt.Sprintf("  - [ ] %s\n", v))
					}
				}
			}
			if len(t.Files) > 0 {
				buf.WriteString("\n  **files:** ")
				for j, f := range t.Files {
					if j > 0 {
						buf.WriteString(", ")
					}
					buf.WriteString(f)
				}
				buf.WriteString("\n")
			}
			buf.WriteString("\n")
		}
	}

	// Render context
	buf.WriteString("\n## context\n\n")
	buf.WriteString(fmt.Sprintf("| field | value |\n"))
	buf.WriteString(fmt.Sprintf("|---|---|\n"))
	buf.WriteString(fmt.Sprintf("| current_file | %s |\n", nullDisplay(string(p.Context.CurrentFile))))
	buf.WriteString(fmt.Sprintf("| last_error | %s |\n", nullDisplay(string(p.Context.LastError))))
	buf.WriteString(fmt.Sprintf("| test_state | %s |\n", nullDisplay(string(p.Context.TestState))))
	buf.WriteString(fmt.Sprintf("| open_questions | %s |\n", nullDisplay(string(p.Context.OpenQuestions))))
	buf.WriteString(fmt.Sprintf("| pending_refactor | %s |\n", nullDisplay(string(p.Context.PendingRefactor))))

	// Render decisions
	buf.WriteString("\n## decisions\n\n")
	if len(p.Decisions) == 0 {
		buf.WriteString("No decisions yet.\n")
	}
	for _, d := range p.Decisions {
		buf.WriteString(fmt.Sprintf("- %s · %s\n", d.Time.Format("2006-01-02"), d.Text))
	}

	// Notes section (user-editable)
	buf.WriteString("\n## notes\n")
	if notes != "" {
		buf.WriteString("\n")
		buf.WriteString(notes)
	}

	return buf.Bytes(), nil
}

func taskCheckbox(status string) string {
	switch status {
	case "done":
		return "x"
	case "active":
		return "▶"
	case "blocked":
		return "!"
	default:
		return " "
	}
}

func isVerifyPassed(v string) bool {
	return strings.HasPrefix(v, "✓ ")
}

func nullDisplay(s string) string {
	if s == "" {
		return "~"
	}
	return s
}

// WriteFile atomically writes a plan to disk.
// It creates a backup of the existing file (if any) at .plans/.backup before writing.
func WriteFile(path string, p *Plan, notes string) error {
	data, err := Serialize(p, notes)
	if err != nil {
		return err
	}

	// Create backup if file already exists
	if _, statErr := os.Stat(path); statErr == nil {
		if err := createBackup(path); err != nil {
			return fmt.Errorf("creating backup: %w", err)
		}
	}

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	// Atomic write: temp file + rename
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp) // clean up on failure
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}

// createBackup copies the current file to .plans/.backup in the same directory.
func createBackup(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	backupPath := filepath.Join(dir, ".backup")
	return os.WriteFile(backupPath, data, 0o644)
}
