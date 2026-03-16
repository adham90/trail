package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adham90/trail/internal/plan"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show progress across all plans",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	plansDir, err := plan.PlansDir()
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(plansDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No plans found.")
			return nil
		}
		return fmt.Errorf("reading plans/: %w", err)
	}

	current, _ := plan.GetCurrent()

	found := false
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		path := filepath.Join(plansDir, e.Name())
		status, err := plan.ParsePlanStatus(path)
		if err != nil {
			continue
		}
		found = true
		marker := " "
		if status.Name != "" && plan.NameToFilename(current) == e.Name() {
			marker = "*"
		}
		fmt.Printf("%s %-25s %d/%d\n", marker, status.Name, status.DoneCount, status.Total)
	}

	if !found {
		fmt.Println("No plans found.")
	}
	return nil
}
