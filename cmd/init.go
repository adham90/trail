package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adham90/trail/internal/plan"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Set up trail in the current project",
	Long:  "Creates plans/ directory, adds .current to .gitignore, and appends trail instructions to CLAUDE.md.",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	root, err := plan.GitRoot()
	if err != nil {
		return err
	}

	// 1. Create plans/ directory
	plansDir := filepath.Join(root, "plans")
	if err := os.MkdirAll(plansDir, 0o755); err != nil {
		return fmt.Errorf("creating plans/: %w", err)
	}
	fmt.Println("Created plans/")

	// 2. Add plans/.current to .gitignore
	gitignorePath := filepath.Join(root, ".gitignore")
	if ensureGitignore(gitignorePath) {
		fmt.Println("Added plans/.current to .gitignore")
	} else {
		fmt.Println("plans/.current already in .gitignore")
	}

	// 3. Append trail section to CLAUDE.md
	claudeMDPath := filepath.Join(root, "CLAUDE.md")
	if appendClaudeMD(claudeMDPath) {
		fmt.Println("Added trail instructions to CLAUDE.md")
	} else {
		fmt.Println("CLAUDE.md already has trail instructions")
	}

	return nil
}

// ensureGitignore adds plans/.current to .gitignore if not present.
// Returns true if the line was added.
func ensureGitignore(path string) bool {
	data, _ := os.ReadFile(path)
	content := string(data)

	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == "plans/.current" {
			return false
		}
	}

	// Ensure trailing newline before appending
	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += "plans/.current\n"
	os.WriteFile(path, []byte(content), 0o644)
	return true
}

const trailClaudeMD = `## Planning: trail

Use ` + "`trail`" + ` for planning across sessions. Plans live in ` + "`plans/`" + ` as Markdown — read and edit them directly for tasks, specs, notes.

- ` + "`trail plan \"name\"`" + ` — create or select a plan
- ` + "`trail plan`" + ` — list all plans
- ` + "`trail status`" + ` — progress overview
- ` + "`trail archive`" + ` — archive completed plan

Plan format: top-level ` + "`- [ ]`" + ` / ` + "`- [x]`" + ` under ` + "`## Tasks`" + ` are counted for progress. Sub-tasks (indented) are for your own tracking.
`

// appendClaudeMD appends trail instructions to CLAUDE.md if not already present.
// Returns true if instructions were added.
func appendClaudeMD(path string) bool {
	data, _ := os.ReadFile(path)
	content := string(data)

	// Check if trail section already exists
	if strings.Contains(content, "## Planning: trail") {
		return false
	}

	if len(content) > 0 && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	if len(content) > 0 {
		content += "\n"
	}
	content += trailClaudeMD
	os.WriteFile(path, []byte(content), 0o644)
	return true
}
