package cmd

import (
	"fmt"
	"strings"

	"github.com/adham90/trail/internal/plan"
	"github.com/spf13/cobra"
)

var (
	afterIdx  int
	addSpec   string
	addVerify []string
	addFiles  []string
)

var addCmd = &cobra.Command{
	Use:   "add [task description]",
	Short: "Add a new task to the plan",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runAdd,
}

func init() {
	addCmd.Flags().IntVar(&afterIdx, "after", -1, "Insert after task index N")
	addCmd.Flags().StringVar(&addSpec, "spec", "", "Task implementation spec")
	addCmd.Flags().StringSliceVar(&addVerify, "verify", nil, "Verification steps (comma-separated or repeat flag)")
	addCmd.Flags().StringSliceVar(&addFiles, "files", nil, "Related files (comma-separated or repeat flag)")
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	planPath, err := resolvePlanPathFromArgs(nil)
	if err != nil {
		return err
	}

	p, notes, err := plan.ReadFile(planPath)
	if err != nil {
		return fmt.Errorf("reading plan: %w", err)
	}

	task := plan.Task{
		Text:   strings.Join(args, " "),
		Status: "todo",
		Spec:   addSpec,
		Verify: addVerify,
		Files:  addFiles,
	}

	var idx int
	if afterIdx >= 0 {
		if afterIdx >= len(p.Tasks) {
			return fmt.Errorf("--after index %d out of range (0-%d)", afterIdx, len(p.Tasks)-1)
		}
		idx = afterIdx + 1
		p.Tasks = append(p.Tasks[:idx], append([]plan.Task{task}, p.Tasks[idx:]...)...)
	} else {
		idx = len(p.Tasks)
		p.Tasks = append(p.Tasks, task)
	}

	if err := plan.WriteFile(planPath, p, notes); err != nil {
		return fmt.Errorf("writing plan: %w", err)
	}

	fmt.Printf("Added task %02d · %s\n", idx, task.Text)
	return nil
}
