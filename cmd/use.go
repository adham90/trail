package cmd

import (
	"fmt"
	"os"

	"github.com/adham90/trail/internal/plan"
	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Set the current plan and switch to its branch",
	Args:  cobra.ExactArgs(1),
	RunE:  runUse,
}

func init() {
	rootCmd.AddCommand(useCmd)
}

func runUse(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Verify plan exists
	planPath, err := plan.ResolvePlanPath(name)
	if err != nil {
		return err
	}
	if _, err := os.Stat(planPath); os.IsNotExist(err) {
		return fmt.Errorf("plan %q not found", name)
	}

	p, _, err := plan.ReadFile(planPath)
	if err != nil {
		return fmt.Errorf("reading plan: %w", err)
	}

	// Set as current
	if err := plan.SetCurrent(name); err != nil {
		return fmt.Errorf("setting current plan: %w", err)
	}

	// Switch to branch if plan has one
	if p.Branch != "" {
		currentBranch, _ := plan.CurrentBranch()
		if currentBranch != p.Branch {
			if err := plan.SwitchBranch(p.Branch); err != nil {
				return err
			}
			fmt.Printf("Switched to branch %s\n", p.Branch)
		}
	}

	fmt.Printf("Now using plan: %s\n", name)
	return nil
}
