package plan

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GitRoot walks up from cwd to find the directory containing .git/.
func GitRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting cwd: %w", err)
	}

	for {
		if info, err := os.Stat(filepath.Join(dir, ".git")); err == nil && info.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not inside a git repository (walked up to %s)", dir)
		}
		dir = parent
	}
}

// NameToFilename converts a plan name to a filename slug.
func NameToFilename(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "/", "-")
	return slug + ".md"
}

// PlansDir returns the plans/ directory path at the git root.
func PlansDir() (string, error) {
	root, err := GitRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "plans"), nil
}

// ResolvePlanPath returns the full path to a named plan file.
func ResolvePlanPath(name string) (string, error) {
	dir, err := PlansDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, NameToFilename(name)), nil
}

// CurrentFile returns the path to the .current file that tracks the active plan.
func CurrentFile() (string, error) {
	dir, err := PlansDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".current"), nil
}

// SetCurrent writes the current plan name to plans/.current.
func SetCurrent(name string) error {
	path, err := CurrentFile()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(name+"\n"), 0o644)
}

// GetCurrent reads the current plan name from plans/.current.
// Returns empty string if no current plan is set.
func GetCurrent() (string, error) {
	path, err := CurrentFile()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", nil // no current plan set
	}
	return strings.TrimSpace(string(data)), nil
}

// ResolveCurrentPlan figures out which plan to use:
// 1. If a name is given, use that
// 2. If plans/.current exists, use that
// 3. If only one plan exists, use that
// Returns the plan name or error.
func ResolveCurrentPlan(name string) (string, error) {
	if name != "" {
		return name, nil
	}

	// Check .current
	current, err := GetCurrent()
	if err != nil {
		return "", err
	}
	if current != "" {
		return current, nil
	}

	// Single plan fallback
	dir, err := PlansDir()
	if err != nil {
		return "", err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("no current plan set — use 'trail plan <name>' to create or select a plan")
	}

	var plans []string
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		plans = append(plans, strings.TrimSuffix(e.Name(), ".md"))
	}

	if len(plans) == 1 {
		return plans[0], nil
	}

	return "", fmt.Errorf("no current plan set — use 'trail plan <name>' to create or select a plan")
}

// ListPlans returns all plan file slugs in the plans/ directory.
func ListPlans() ([]string, error) {
	dir, err := PlansDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		names = append(names, strings.TrimSuffix(e.Name(), ".md"))
	}
	return names, nil
}
