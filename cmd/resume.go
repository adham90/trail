package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adhameldeeb/trail/internal/plan"
	"github.com/spf13/cobra"
)

var resumeCmd = &cobra.Command{
	Use:   "resume [plan-name]",
	Short: "Print CLAUDE.md + active plan for session handoff",
	RunE:  runResume,
}

func init() {
	rootCmd.AddCommand(resumeCmd)
}

func runResume(cmd *cobra.Command, args []string) error {
	name, err := resolvePlanName(args)
	if err != nil {
		return err
	}

	root, err := plan.GitRoot()
	if err != nil {
		return err
	}

	// Print CLAUDE.md if it exists
	claudePath := filepath.Join(root, "CLAUDE.md")
	if data, err := os.ReadFile(claudePath); err == nil {
		fmt.Println(string(data))
		fmt.Println("---")
		fmt.Println()
	}

	// Print plan file
	planPath, err := plan.ResolvePlanPath(name)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(planPath)
	if err != nil {
		return fmt.Errorf("plan %q not found: %w", name, err)
	}

	fmt.Print(string(data))
	return nil
}
