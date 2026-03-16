package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/adham90/trail/internal/plan"
	"github.com/adham90/trail/internal/renderer"
	"github.com/spf13/cobra"
)

var doneCmd = &cobra.Command{
	Use:   "done N",
	Short: "Mark task N as done (1-based)",
	Args:  cobra.ExactArgs(1),
	RunE:  runDone,
}

func init() {
	rootCmd.AddCommand(doneCmd)
}

func runDone(cmd *cobra.Command, args []string) error {
	taskNum, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid task number: %s", args[0])
	}

	planPath, err := resolvePlanPathFromArgs(nil)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(planPath)
	if err != nil {
		return fmt.Errorf("reading plan: %w", err)
	}

	// Get task text before marking done
	tasks := plan.ParseTasks(data)
	var taskText string
	for _, t := range tasks {
		if t.Index == taskNum {
			taskText = t.Text
			break
		}
	}

	if err := plan.CreateBackup(planPath); err != nil {
		return fmt.Errorf("creating backup: %w", err)
	}

	result, err := plan.SetTaskDone(data, taskNum)
	if err != nil {
		return err
	}

	if err := plan.AtomicWriteFile(planPath, result); err != nil {
		return fmt.Errorf("writing plan: %w", err)
	}

	fmt.Printf("%s %s\n", renderer.SymbolDone, taskText)
	return nil
}
