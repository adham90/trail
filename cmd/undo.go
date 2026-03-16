package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adham90/trail/internal/plan"
	"github.com/spf13/cobra"
)

var undoCmd = &cobra.Command{
	Use:   "undo",
	Short: "Revert the last plan write",
	RunE:  runUndo,
}

func init() {
	rootCmd.AddCommand(undoCmd)
}

func runUndo(cmd *cobra.Command, args []string) error {
	planPath, err := resolvePlanPathFromArgs(nil)
	if err != nil {
		return err
	}

	plansDir := filepath.Dir(planPath)
	backupPath := filepath.Join(plansDir, ".backup")

	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("no backup found: %w", err)
	}

	if err := plan.AtomicWriteFile(planPath, backupData); err != nil {
		return fmt.Errorf("restoring backup: %w", err)
	}

	os.Remove(backupPath)

	fmt.Println("Reverted to previous state.")
	return nil
}
