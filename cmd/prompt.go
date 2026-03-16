package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Output the plan format guide for CLAUDE.md",
	Run:   runPrompt,
}

func init() {
	rootCmd.AddCommand(promptCmd)
}

func runPrompt(cmd *cobra.Command, args []string) {
	fmt.Print(formatGuide)
}

var formatGuide = "# Trail Plan Format\n" +
	"\n" +
	"Trail plans are pure Markdown files. The coding agent writes and maintains them directly.\n" +
	"\n" +
	"## Structure\n" +
	"\n" +
	"```markdown\n" +
	"# Plan Name\n" +
	"\n" +
	"Goal description.\n" +
	"\n" +
	"## Acceptance Criteria\n" +
	"\n" +
	"- [ ] criterion 1\n" +
	"- [ ] criterion 2\n" +
	"\n" +
	"## Diagram (optional)\n" +
	"\n" +
	"```mermaid\n" +
	"graph TD\n" +
	"    A --> B\n" +
	"```\n" +
	"\n" +
	"## Constraints (optional)\n" +
	"\n" +
	"- rule 1\n" +
	"\n" +
	"## Tasks\n" +
	"\n" +
	"- [ ] **1.** Task title\n" +
	"  Description/spec.\n" +
	"  - [ ] 1.1. verify step\n" +
	"  - [ ] 1.2. verify step\n" +
	"  `file1.go`, `file2.go`\n" +
	"\n" +
	"- [x] **2.** Completed task\n" +
	"  - [x] 2.1. done step\n" +
	"\n" +
	"## Decisions (optional)\n" +
	"\n" +
	"- 2026-03-16: Decision text\n" +
	"\n" +
	"## Notes (optional)\n" +
	"\n" +
	"Freeform.\n" +
	"```\n" +
	"\n" +
	"## Rules\n" +
	"\n" +
	"- Trail parses ONLY top-level checkboxes under ## Tasks for status\n" +
	"- Task numbering is 1-based (in both Markdown and trail commands)\n" +
	"- Sub-items (indented checkboxes) are ignored by trail\n" +
	"- The agent edits the plan file directly for most changes\n" +
	"- Use `trail done N` to mark tasks complete\n" +
	"- Use `trail block N \"reason\"` to mark tasks blocked\n"
