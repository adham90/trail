package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/adham90/trail/internal/plan"
	"github.com/spf13/cobra"
)

var blockCmd = &cobra.Command{
	Use:   "block N \"reason\"",
	Short: "Mark task N as blocked (1-based)",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runBlock,
}

func init() {
	rootCmd.AddCommand(blockCmd)
}

func runBlock(cmd *cobra.Command, args []string) error {
	taskNum, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid task number: %s", args[0])
	}
	reason := strings.Join(args[1:], " ")

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

	result, err := plan.SetTaskBlocked(data, taskNum, reason)
	if err != nil {
		return err
	}

	if err := plan.AtomicWriteFile(planPath, result); err != nil {
		return fmt.Errorf("writing plan: %w", err)
	}

	fmt.Printf("! task %d blocked: %s\n", taskNum, reason)
	return nil
}
