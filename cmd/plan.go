package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/adham90/trail/internal/plan"
	"github.com/adham90/trail/internal/renderer"
	"github.com/spf13/cobra"
)

var (
	planGoal     string
	planNew      bool
	planNoBranch bool
)

var planCmd = &cobra.Command{
	Use:   "plan [name]",
	Short: "List plans, or open/create a specific plan",
	RunE:  runPlan,
}

func init() {
	planCmd.Flags().StringVar(&planGoal, "goal", "", "Plan goal (used with --new)")
	planCmd.Flags().BoolVar(&planNew, "new", false, "Create a new plan")
	planCmd.Flags().BoolVar(&planNoBranch, "no-branch", false, "Don't create a git branch")
	rootCmd.AddCommand(planCmd)
}

func runPlan(cmd *cobra.Command, args []string) error {
	// No args — list all plans
	if len(args) == 0 && !planNew {
		return listPlans()
	}

	if len(args) == 0 && planNew {
		return fmt.Errorf("plan name required: trail plan --new <name> --goal \"...\"")
	}

	name := args[0]

	if planNew {
		return createNewPlan(name)
	}

	// Open existing plan
	return openPlan(name)
}

func listPlans() error {
	dir, err := plan.PlansDir()
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No plans found. Create one with: trail plan --new <name> --goal \"...\"")
			return nil
		}
		return err
	}

	found := false
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		path := filepath.Join(dir, e.Name())
		p, _, err := plan.ReadFile(path)
		if err != nil {
			continue
		}
		found = true
		doneCount := 0
		for _, t := range p.Tasks {
			if t.Status == "done" {
				doneCount++
			}
		}
		branchInfo := ""
		if p.Branch != "" {
			branchInfo = fmt.Sprintf("  (%s)", p.Branch)
		}
		fmt.Printf("%-25s %-10s %d/%d    session %d%s\n",
			p.Name, p.Status, doneCount, len(p.Tasks), p.SessionCount, branchInfo)
	}

	if !found {
		fmt.Println("No plans found. Create one with: trail plan --new <name> --goal \"...\"")
	}
	return nil
}

func createNewPlan(name string) error {
	planPath, err := plan.ResolvePlanPath(name)
	if err != nil {
		return err
	}

	// Check if already exists
	if _, statErr := os.Stat(planPath); statErr == nil {
		return fmt.Errorf("plan %q already exists at %s", name, planPath)
	}

	if planGoal == "" {
		return fmt.Errorf("--goal is required when creating a plan: trail plan --new %s --goal \"...\"", name)
	}

	// Ensure plans/ directory exists
	dir := filepath.Dir(planPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating plans/: %w", err)
	}

	today := time.Now().Format("2006-01-02")
	p := &plan.Plan{
		Name:         name,
		Goal:         planGoal,
		Status:       "active",
		SessionCount: 1,
		Created:      today,
		Updated:      today,
		CurrentTask:  0,
		Tasks:        []plan.Task{},
		Context:      plan.Context{},
		Decisions:    []plan.Decision{},
	}

	// Create branch unless --no-branch
	if !planNoBranch {
		branchName := plan.NameToBranch(name)
		if plan.BranchExists(branchName) {
			return fmt.Errorf("branch %s already exists", branchName)
		}
		if err := plan.CreateBranch(branchName); err != nil {
			return err
		}
		p.Branch = branchName
		fmt.Printf("Created branch %s\n", branchName)
	}

	if err := plan.WriteFile(planPath, p, ""); err != nil {
		return fmt.Errorf("writing plan: %w", err)
	}

	// Set as current plan
	plan.SetCurrent(name)

	fmt.Printf("Created %s\n", planPath)
	return nil
}

func openPlan(name string) error {
	planPath, err := plan.ResolvePlanPath(name)
	if err != nil {
		return err
	}

	p, _, err := plan.ReadFile(planPath)
	if err != nil {
		return fmt.Errorf("plan %q not found: %w", name, err)
	}

	// Auto-increment session_count if updated date differs from today
	today := time.Now().Format("2006-01-02")
	if p.Updated != today {
		p.SessionCount++
		p.Updated = today
		notes := getNotesFromFile(planPath)
		if err := plan.WriteFile(planPath, p, notes); err != nil {
			return fmt.Errorf("updating session count: %w", err)
		}
	}

	// Set as current
	plan.SetCurrent(name)

	fmt.Print(renderer.Summary(p))
	return nil
}

func getNotesFromFile(path string) string {
	_, notes, err := plan.ReadFile(path)
	if err != nil {
		return ""
	}
	return notes
}
