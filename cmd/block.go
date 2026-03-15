package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/adham90/trail/internal/plan"
	"github.com/adham90/trail/internal/renderer"
	"github.com/spf13/cobra"
)

var blockCmd = &cobra.Command{
	Use:   "block [reason] or block N [reason]",
	Short: "Mark a task as blocked (defaults to active task)",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runBlock,
}

func init() {
	rootCmd.AddCommand(blockCmd)
}

func runBlock(cmd *cobra.Command, args []string) error {
	planPath, err := resolvePlanPathFromArgs(nil)
	if err != nil {
		return err
	}

	p, notes, err := plan.ReadFile(planPath)
	if err != nil {
		return fmt.Errorf("reading plan: %w", err)
	}

	var taskIdx int
	var reason string

	if idx, parseErr := strconv.Atoi(args[0]); parseErr == nil {
		taskIdx = idx
		reason = strings.Join(args[1:], " ")
	} else {
		taskIdx = -1
		for i := range p.Tasks {
			if p.Tasks[i].Status == "active" {
				taskIdx = i
				break
			}
		}
		if taskIdx < 0 {
			return fmt.Errorf("no active task to block")
		}
		reason = strings.Join(args, " ")
	}

	if taskIdx < 0 || taskIdx >= len(p.Tasks) {
		return fmt.Errorf("task index %d out of range (0-%d)", taskIdx, len(p.Tasks)-1)
	}

	p.Tasks[taskIdx].Status = "blocked"
	p.Tasks[taskIdx].Reason = reason

	if err := plan.WriteFile(planPath, p, notes); err != nil {
		return fmt.Errorf("writing plan: %w", err)
	}

	fmt.Printf("%s %02d · %-30s  blocked: %s\n",
		renderer.SymbolBlocked, taskIdx, p.Tasks[taskIdx].Text, reason)
	return nil
}
