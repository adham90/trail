package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/adham90/trail/internal/plan"
	"github.com/adham90/trail/internal/renderer"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit N [new text]",
	Short: "Reword a task",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runEdit,
}

func init() {
	rootCmd.AddCommand(editCmd)
}

func runEdit(cmd *cobra.Command, args []string) error {
	idx, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("first argument must be a task index: %w", err)
	}

	planPath, err := resolvePlanPathFromArgs(nil)
	if err != nil {
		return err
	}

	p, notes, err := plan.ReadFile(planPath)
	if err != nil {
		return fmt.Errorf("reading plan: %w", err)
	}

	if idx < 0 || idx >= len(p.Tasks) {
		return fmt.Errorf("task index %d out of range (0-%d)", idx, len(p.Tasks)-1)
	}

	p.Tasks[idx].Text = strings.Join(args[1:], " ")

	if err := plan.WriteFile(planPath, p, notes); err != nil {
		return fmt.Errorf("writing plan: %w", err)
	}

	fmt.Printf("%s %02d · %s\n", renderer.StatusSymbol(p.Tasks[idx].Status), idx, p.Tasks[idx].Text)
	return nil
}
