package cmd

import (
	"fmt"
	"os"

	"github.com/adham90/trail/internal/plan"
	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Set the current plan",
	Args:  cobra.ExactArgs(1),
	RunE:  runUse,
}

func init() {
	rootCmd.AddCommand(useCmd)
}

func runUse(cmd *cobra.Command, args []string) error {
	name := args[0]

	planPath, err := plan.ResolvePlanPath(name)
	if err != nil {
		return err
	}
	if _, err := os.Stat(planPath); os.IsNotExist(err) {
		return fmt.Errorf("plan %q not found", name)
	}

	if err := plan.SetCurrent(name); err != nil {
		return fmt.Errorf("setting current plan: %w", err)
	}

	// Switch to branch if plan/<name> exists
	branchName := plan.NameToBranch(name)
	if plan.BranchExists(branchName) {
		currentBranch, _ := plan.CurrentBranch()
		if currentBranch != branchName {
			if err := plan.SwitchBranch(branchName); err != nil {
				return err
			}
			fmt.Printf("Switched to branch %s\n", branchName)
		}
	}

	fmt.Printf("Now using plan: %s\n", name)
	return nil
}
