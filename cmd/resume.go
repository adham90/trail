package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adham90/trail/internal/plan"
	"github.com/spf13/cobra"
)

var resumeCmd = &cobra.Command{
	Use:   "resume [plan-name]",
	Short: "Print plan for session handoff",
	RunE:  runResume,
}

func init() {
	rootCmd.AddCommand(resumeCmd)
}

func runResume(cmd *cobra.Command, args []string) error {
	planPath, err := resolvePlanPathFromArgs(args)
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

	data, err := os.ReadFile(planPath)
	if err != nil {
		return fmt.Errorf("plan not found: %w", err)
	}

	fmt.Print(string(data))
	return nil
}
