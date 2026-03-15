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
	name, err := resolvePlanName(nil)
	if err != nil {
		return err
	}

	planPath, err := plan.ResolvePlanPath(name)
	if err != nil {
		return err
	}

	plansDir := filepath.Dir(planPath)
	backupPath := filepath.Join(plansDir, ".backup")

	backupData, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("no backup found: %w", err)
	}

	tmp := planPath + ".tmp"
	if err := os.WriteFile(tmp, backupData, 0o644); err != nil {
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := os.Rename(tmp, planPath); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("restoring backup: %w", err)
	}

	os.Remove(backupPath)

	fmt.Println("Reverted to previous state.")
	return nil
}
