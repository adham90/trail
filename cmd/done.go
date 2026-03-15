package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/adham90/trail/internal/plan"
	"github.com/spf13/cobra"
)

var doneCmd = &cobra.Command{
	Use:   "done [plan-name]",
	Short: "Complete the plan and archive it",
	RunE:  runDone,
}

func init() {
	rootCmd.AddCommand(doneCmd)
}

func runDone(cmd *cobra.Command, args []string) error {
	name, err := resolvePlanName(args)
	if err != nil {
		return err
	}

	planPath, err := plan.ResolvePlanPath(name)
	if err != nil {
		return err
	}

	p, notes, err := plan.ReadFile(planPath)
	if err != nil {
		return fmt.Errorf("reading plan: %w", err)
	}

	// Warn about incomplete tasks
	todoCount := 0
	blockedCount := 0
	doneCount := 0
	for _, t := range p.Tasks {
		switch t.Status {
		case "done":
			doneCount++
		case "todo", "active":
			todoCount++
		case "blocked":
			blockedCount++
		}
	}
	if todoCount > 0 || blockedCount > 0 {
		fmt.Printf("Warning: %d todo, %d blocked tasks remaining\n", todoCount, blockedCount)
	}

	p.Status = "complete"
	p.Updated = time.Now().Format("2006-01-02")

	if err := plan.WriteFile(planPath, p, notes); err != nil {
		return fmt.Errorf("writing plan: %w", err)
	}

	// Archive
	archiveDir := filepath.Join(filepath.Dir(planPath), "archive")
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
		return fmt.Errorf("creating archive dir: %w", err)
	}

	archivePath := filepath.Join(archiveDir, filepath.Base(planPath))
	if err := os.Rename(planPath, archivePath); err != nil {
		return fmt.Errorf("archiving plan: %w", err)
	}

	// Clear current if this was it
	current, _ := plan.GetCurrent()
	if current == name {
		plan.SetCurrent("")
	}

	fmt.Printf("Plan complete\n")
	fmt.Printf("%d/%d tasks · %d decisions · %d sessions\n",
		doneCount, len(p.Tasks), len(p.Decisions), p.SessionCount)
	fmt.Printf("Archived to %s\n", archivePath)

	if p.Branch != "" {
		fmt.Printf("Branch %s ready for PR\n", p.Branch)
	}

	return nil
}
