package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/adham90/trail/internal/plan"
	"github.com/spf13/cobra"
)

var decideCmd = &cobra.Command{
	Use:   "decide [reason]",
	Short: "Log a timestamped decision",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runDecide,
}

func init() {
	rootCmd.AddCommand(decideCmd)
}

func runDecide(cmd *cobra.Command, args []string) error {
	planPath, err := resolvePlanPathFromArgs(nil)
	if err != nil {
		return err
	}

	p, notes, err := plan.ReadFile(planPath)
	if err != nil {
		return fmt.Errorf("reading plan: %w", err)
	}

	now := time.Now().UTC()
	decision := plan.Decision{
		Time: now,
		Text: strings.Join(args, " "),
	}
	p.Decisions = append(p.Decisions, decision)
	p.Updated = time.Now().Format("2006-01-02")

	if err := plan.WriteFile(planPath, p, notes); err != nil {
		return fmt.Errorf("writing plan: %w", err)
	}

	fmt.Printf("Decision logged at %s\n", now.Format(time.RFC3339))
	return nil
}
