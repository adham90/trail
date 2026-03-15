package cmd

import (
	"fmt"

	"github.com/adham90/trail/internal/plan"
	"github.com/adham90/trail/internal/renderer"
	"github.com/spf13/cobra"
)

var nextCmd = &cobra.Command{
	Use:   "next [plan-name]",
	Short: "Complete the active task and activate the next one",
	RunE:  runNext,
}

func init() {
	nextCmd.Flags().Bool("skip", false, "Skip current task (mark as todo, not done)")
	rootCmd.AddCommand(nextCmd)
}

func runNext(cmd *cobra.Command, args []string) error {
	planPath, err := resolvePlanPathFromArgs(args)
	if err != nil {
		return err
	}

	p, notes, err := plan.ReadFile(planPath)
	if err != nil {
		return fmt.Errorf("reading plan: %w", err)
	}

	skip, _ := cmd.Flags().GetBool("skip")

	// Find and handle active task
	var completed *plan.Task
	activeIdx := -1
	for i := range p.Tasks {
		if p.Tasks[i].Status == "active" {
			activeIdx = i
			break
		}
	}

	if activeIdx >= 0 {
		if skip {
			p.Tasks[activeIdx].Status = "todo"
		} else {
			p.Tasks[activeIdx].Status = "done"
			completed = &p.Tasks[activeIdx]
		}
	}

	// Find next todo task (start after the skipped/completed task)
	startIdx := 0
	if activeIdx >= 0 {
		startIdx = activeIdx + 1
	}
	var active *plan.Task
	for offset := 0; offset < len(p.Tasks); offset++ {
		i := (startIdx + offset) % len(p.Tasks)
		if p.Tasks[i].Status == "todo" {
			p.Tasks[i].Status = "active"
			p.CurrentTask = i
			active = &p.Tasks[i]
			break
		}
	}

	if active == nil && completed == nil {
		return fmt.Errorf("no tasks to advance")
	}

	if err := plan.WriteFile(planPath, p, notes); err != nil {
		return fmt.Errorf("writing plan: %w", err)
	}

	if active == nil {
		fmt.Println("All tasks complete!")
		return nil
	}

	fmt.Print(renderer.ContextBlock(completed, active, p.Context))
	return nil
}
