package cmd

import (
	"fmt"

	"github.com/adhameldeeb/trail/internal/plan"
)

// resolvePlanPathFromArgs resolves a plan path from optional args.
// If args has a name, uses that. Otherwise uses current plan.
func resolvePlanPathFromArgs(args []string) (string, error) {
	var name string
	if len(args) > 0 {
		name = args[0]
	}
	resolved, err := plan.ResolveCurrentPlan(name)
	if err != nil {
		return "", err
	}
	return plan.ResolvePlanPath(resolved)
}

// resolvePlanName resolves a plan name from optional args.
func resolvePlanName(args []string) (string, error) {
	var name string
	if len(args) > 0 {
		name = args[0]
	}
	resolved, err := plan.ResolveCurrentPlan(name)
	if err != nil {
		return "", fmt.Errorf("%w\nTip: run 'trail use <name>' to set a current plan", err)
	}
	return resolved, nil
}
