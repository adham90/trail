package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/adhameldeeb/trail/internal/plan"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "List all plans with progress",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

type planInfo struct {
	name    string
	plan    *plan.Plan
	updated string
}

func runStatus(cmd *cobra.Command, args []string) error {
	plansDir, err := plan.PlansDir()
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(plansDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No plans found.")
			return nil
		}
		return fmt.Errorf("reading plans/: %w", err)
	}

	current, _ := plan.GetCurrent()

	var plans []planInfo
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		path := filepath.Join(plansDir, e.Name())
		p, _, err := plan.ReadFile(path)
		if err != nil {
			continue
		}
		plans = append(plans, planInfo{name: p.Name, plan: p, updated: p.Updated})
	}

	if len(plans) == 0 {
		fmt.Println("No plans found.")
		return nil
	}

	sort.Slice(plans, func(i, j int) bool {
		return plans[i].updated > plans[j].updated
	})

	for _, pi := range plans {
		doneCount := 0
		for _, t := range pi.plan.Tasks {
			if t.Status == "done" {
				doneCount++
			}
		}
		total := len(pi.plan.Tasks)
		marker := " "
		if pi.name == current {
			marker = "*"
		}
		branchInfo := ""
		if pi.plan.Branch != "" {
			branchInfo = fmt.Sprintf("  (%s)", pi.plan.Branch)
		}
		fmt.Printf("%s %-25s %-10s %d/%d    session %d%s\n",
			marker, pi.name, pi.plan.Status, doneCount, total, pi.plan.SessionCount, branchInfo)
	}

	return nil
}
