package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adham90/trail/internal/plan"
	"github.com/spf13/cobra"
)

var archiveCmd = &cobra.Command{
	Use:   "archive [plan-name]",
	Short: "Archive a completed plan",
	RunE:  runArchive,
}

func init() {
	rootCmd.AddCommand(archiveCmd)
}

func runArchive(cmd *cobra.Command, args []string) error {
	var name string
	if len(args) > 0 {
		name = args[0]
	}
	resolved, err := plan.ResolveCurrentPlan(name)
	if err != nil {
		return err
	}

	planPath, err := plan.ResolvePlanPath(resolved)
	if err != nil {
		return err
	}

	if _, err := os.Stat(planPath); os.IsNotExist(err) {
		return fmt.Errorf("plan not found: %s", planPath)
	}

	status, err := plan.ParsePlanStatus(planPath)
	if err != nil {
		return fmt.Errorf("reading plan: %w", err)
	}

	archiveDir := filepath.Join(filepath.Dir(planPath), "archive")
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
		return fmt.Errorf("creating archive dir: %w", err)
	}

	archivePath := filepath.Join(archiveDir, filepath.Base(planPath))
	if err := os.Rename(planPath, archivePath); err != nil {
		return fmt.Errorf("archiving plan: %w", err)
	}

	// Clear .current if this was the active plan
	current, _ := plan.GetCurrent()
	slug := filepath.Base(planPath)
	if plan.NameToFilename(current) == slug {
		plan.SetCurrent("")
	}

	fmt.Printf("Archived %s (%d/%d tasks done)\n", status.Name, status.DoneCount, status.Total)
	fmt.Printf("Moved to %s\n", archivePath)
	return nil
}
