package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/adham90/trail/internal/plan"
	"github.com/spf13/cobra"
)

var checkpointCmd = &cobra.Command{
	Use:   "checkpoint [plan-name]",
	Short: "Update context state in the plan",
	RunE:  runCheckpoint,
}

func init() {
	checkpointCmd.Flags().String("file", "", "Current file path")
	checkpointCmd.Flags().String("error", "", "Last error")
	checkpointCmd.Flags().String("tests", "", "Test state")
	checkpointCmd.Flags().String("note", "", "Open question or note")
	checkpointCmd.Flags().String("verify", "", "Mark verify step as passed (by text match)")
	rootCmd.AddCommand(checkpointCmd)
}

func runCheckpoint(cmd *cobra.Command, args []string) error {
	planPath, err := resolvePlanPathFromArgs(args)
	if err != nil {
		return err
	}

	p, notes, err := plan.ReadFile(planPath)
	if err != nil {
		return fmt.Errorf("reading plan: %w", err)
	}

	if cmd.Flags().Changed("file") {
		v, _ := cmd.Flags().GetString("file")
		p.Context.CurrentFile = plan.NullableString(v)
	}
	if cmd.Flags().Changed("error") {
		v, _ := cmd.Flags().GetString("error")
		p.Context.LastError = plan.NullableString(v)
	}
	if cmd.Flags().Changed("tests") {
		v, _ := cmd.Flags().GetString("tests")
		p.Context.TestState = plan.NullableString(v)
	}
	if cmd.Flags().Changed("note") {
		v, _ := cmd.Flags().GetString("note")
		p.Context.OpenQuestions = plan.NullableString(v)
	}

	if cmd.Flags().Changed("verify") {
		verifyText, _ := cmd.Flags().GetString("verify")
		found := false
		for i := range p.Tasks {
			if p.Tasks[i].Status == "active" {
				for j, v := range p.Tasks[i].Verify {
					if strings.Contains(v, verifyText) && !strings.HasPrefix(v, "✓ ") {
						p.Tasks[i].Verify[j] = "✓ " + v
						found = true
						fmt.Printf("Verified: %s\n", p.Tasks[i].Verify[j])
						break
					}
				}
				break
			}
		}
		if !found {
			return fmt.Errorf("no matching verify step for %q in active task", verifyText)
		}
	}

	p.Updated = time.Now().Format("2006-01-02")

	if err := plan.WriteFile(planPath, p, notes); err != nil {
		return fmt.Errorf("writing plan: %w", err)
	}

	fmt.Println("Checkpoint saved.")
	return nil
}
