package renderer

import (
	"fmt"
	"strings"

	"github.com/adham90/trail/internal/plan"
)

// Summary renders the full session summary (trail plan output).
func Summary(p *plan.Plan) string {
	var b strings.Builder

	// Header
	fmt.Fprintf(&b, "%s  %s\n", Bold("plan:"), p.Name)
	fmt.Fprintf(&b, "%s  %s\n", Bold("goal:"), p.Goal)
	if p.Branch != "" {
		fmt.Fprintf(&b, "%s  %s\n", Bold("branch:"), p.Branch)
	}
	fmt.Fprintf(&b, "%s  %d\n", Bold("session:"), p.SessionCount)
	b.WriteString("\n")

	// Blocked tasks first
	hasBlocked := false
	for i, t := range p.Tasks {
		if t.Status == "blocked" {
			if !hasBlocked {
				hasBlocked = true
			}
			fmt.Fprintf(&b, "%s %02d · %-30s  blocked: %s\n", SymbolBlocked, i, t.Text, t.Reason)
		}
	}
	if hasBlocked {
		b.WriteString("\n")
	}

	// All tasks
	for i, t := range p.Tasks {
		fmt.Fprintf(&b, "%s %02d · %s\n", StatusSymbol(t.Status), i, t.Text)
	}
	b.WriteString("\n")

	// Context
	b.WriteString(Bold("context:") + "\n")
	fmt.Fprintf(&b, "  current_file:  %s\n", nullDisplay(string(p.Context.CurrentFile)))
	fmt.Fprintf(&b, "  last_error:    %s\n", nullDisplay(string(p.Context.LastError)))
	fmt.Fprintf(&b, "  test_state:    %s\n", nullDisplay(string(p.Context.TestState)))
	b.WriteString("\n")

	// Decisions count
	fmt.Fprintf(&b, "%s %d logged\n", Bold("decisions:"), len(p.Decisions))
	if len(p.Constraints) > 0 {
		fmt.Fprintf(&b, "%s %d defined\n", Bold("constraints:"), len(p.Constraints))
	}
	if len(p.PlanFiles) > 0 {
		fmt.Fprintf(&b, "%s %d tracked\n", Bold("files:"), len(p.PlanFiles))
	}

	return b.String()
}

func nullDisplay(s string) string {
	if s == "" {
		return "~"
	}
	return s
}
