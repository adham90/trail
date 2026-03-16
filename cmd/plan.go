package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/adham90/trail/internal/plan"
	"github.com/spf13/cobra"
)

var (
	planGoal     string
	planNew      bool
	planNoBranch bool
	planOpen     bool
)

var planCmd = &cobra.Command{
	Use:   "plan [name]",
	Short: "List plans, or create a new plan",
	RunE:  runPlan,
}

func init() {
	planCmd.Flags().StringVar(&planGoal, "goal", "", "Plan goal (used with --new)")
	planCmd.Flags().BoolVar(&planNew, "new", false, "Create a new plan")
	planCmd.Flags().BoolVar(&planNoBranch, "no-branch", false, "Don't create a git branch")
	planCmd.Flags().BoolVar(&planOpen, "open", false, "Open plan in $EDITOR after creation")
	rootCmd.AddCommand(planCmd)
}

func runPlan(cmd *cobra.Command, args []string) error {
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

	// Set as current plan
	planPath, err := plan.ResolvePlanPath(name)
	if err != nil {
		return err
	}
	if _, err := os.Stat(planPath); os.IsNotExist(err) {
		return fmt.Errorf("plan %q not found", name)
	}
	plan.SetCurrent(name)
	fmt.Printf("Now using plan: %s\n", name)
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
		status, err := plan.ParsePlanStatus(path)
		if err != nil {
			continue
		}
		found = true
		fmt.Printf("  %-25s %d/%d\n", status.Name, status.DoneCount, status.Total)
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

	if _, statErr := os.Stat(planPath); statErr == nil {
		return fmt.Errorf("plan %q already exists at %s", name, planPath)
	}

	if planGoal == "" {
		return fmt.Errorf("--goal is required when creating a plan: trail plan --new %s --goal \"...\"", name)
	}

	dir := filepath.Dir(planPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating plans/: %w", err)
	}

	data := plan.GenerateTemplate(name, planGoal)

	// Create branch unless --no-branch
	if !planNoBranch {
		branchName := plan.NameToBranch(name)
		if plan.BranchExists(branchName) {
			return fmt.Errorf("branch %s already exists", branchName)
		}
		if err := plan.CreateBranch(branchName); err != nil {
			return err
		}
		fmt.Printf("Created branch %s\n", branchName)
	}

	if err := plan.AtomicWriteFile(planPath, data); err != nil {
		return fmt.Errorf("writing plan: %w", err)
	}

	plan.SetCurrent(name)
	fmt.Printf("Created %s\n", planPath)

	if planOpen {
		return openInEditor(planPath)
	}
	return nil
}

func openInEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return fmt.Errorf("$EDITOR not set — use 'export EDITOR=zed' or pass the editor name")
	}
	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
