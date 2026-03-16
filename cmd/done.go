package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/adham90/trail/internal/plan"
	"github.com/adham90/trail/internal/renderer"
	"github.com/spf13/cobra"
)

var doneCmd = &cobra.Command{
	Use:   "done N or done N.M",
	Short: "Mark task N (or sub-task N.M) as done",
	Args:  cobra.ExactArgs(1),
	RunE:  runDone,
}

func init() {
	rootCmd.AddCommand(doneCmd)
}

func runDone(cmd *cobra.Command, args []string) error {
	ref := args[0]
	isSubTask := strings.Contains(ref, ".")

	planPath, err := resolvePlanPathFromArgs(nil)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(planPath)
	if err != nil {
		return fmt.Errorf("reading plan: %w", err)
	}

	if err := plan.CreateBackup(planPath); err != nil {
		return fmt.Errorf("creating backup: %w", err)
	}

	var result []byte
	if isSubTask {
		result, err = plan.SetSubTaskDone(data, ref)
	} else {
		taskNum, parseErr := strconv.Atoi(ref)
		if parseErr != nil {
			return fmt.Errorf("invalid task number: %s", ref)
		}
		result, err = plan.SetTaskDone(data, taskNum)
	}
	if err != nil {
		return err
	}

	if err := plan.AtomicWriteFile(planPath, result); err != nil {
		return fmt.Errorf("writing plan: %w", err)
	}

	// Print confirmation
	if isSubTask {
		fmt.Printf("%s %s\n", renderer.SymbolDone, ref)
	} else {
		tasks := plan.ParseTasks(data)
		taskNum, _ := strconv.Atoi(ref)
		var taskText string
		for _, t := range tasks {
			if t.Index == taskNum {
				taskText = t.Text
				break
			}
		}
		fmt.Printf("%s %s\n", renderer.SymbolDone, taskText)
	}
	return nil
}
