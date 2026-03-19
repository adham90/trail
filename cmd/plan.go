package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adham90/trail/internal/plan"
	"github.com/spf13/cobra"
)

var planCmd = &cobra.Command{
	Use:   "plan [name]",
	Short: "Create or select a plan (no args: list all)",
	RunE:  runPlan,
}

func init() {
	rootCmd.AddCommand(planCmd)
}

func runPlan(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return listPlans()
	}

	name := args[0]

	planPath, err := plan.ResolvePlanPath(name)
	if err != nil {
		return err
	}

	// If plan exists, select it
	if _, err := os.Stat(planPath); err == nil {
		if err := plan.SetCurrent(name); err != nil {
			return fmt.Errorf("setting current plan: %w", err)
		}
		fmt.Printf("Now using plan: %s\n", name)
		return nil
	}

	// Create new plan
	dir := filepath.Dir(planPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating plans/: %w", err)
	}

	data := plan.GenerateTemplate(name)

	if err := plan.AtomicWriteFile(planPath, data); err != nil {
		return fmt.Errorf("writing plan: %w", err)
	}

	if err := plan.SetCurrent(name); err != nil {
		return fmt.Errorf("setting current plan: %w", err)
	}
	fmt.Printf("Created %s\n", planPath)
	return nil
}

func listPlans() error {
	dir, err := plan.PlansDir()
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No plans found. Create one with: trail plan <name>")
			return nil
		}
		return err
	}

	current, _ := plan.GetCurrent()

	found := false
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		path := filepath.Join(dir, e.Name())
		status, err := plan.ParsePlanStatus(path)
		if err != nil {
			continue
		}
		found = true
		marker := " "
		if plan.NameToFilename(current) == e.Name() {
			marker = "*"
		}
		fmt.Printf("%s %-25s %d/%d\n", marker, status.Name, status.DoneCount, status.Total)
	}

	if !found {
		fmt.Println("No plans found. Create one with: trail plan <name>")
	}
	return nil
}
