package cmd

import (
	"github.com/adham90/trail/internal/plan"
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
