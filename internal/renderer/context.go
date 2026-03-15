package renderer

import (
	"fmt"
	"strings"

	"github.com/adham90/trail/internal/plan"
)

// ContextBlock renders the compact context view (trail next output).
// Shows the just-completed task (if any) and the new active task.
func ContextBlock(completed *plan.Task, active *plan.Task, ctx plan.Context) string {
	var b strings.Builder

	if completed != nil {
		fmt.Fprintf(&b, "%s %s\n", SymbolDone, completed.Text)
	}
	if active != nil {
		fmt.Fprintf(&b, "%s %s\n", SymbolActive, active.Text)
		if active.Spec != "" {
			fmt.Fprintf(&b, "  spec: %s\n", active.Spec)
		}
		if len(active.Verify) > 0 {
			b.WriteString("  verify:\n")
			for _, v := range active.Verify {
				fmt.Fprintf(&b, "    %s %s\n", SymbolTodo, v)
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(Bold("context:") + "\n")
	fmt.Fprintf(&b, "  current_file:  %s\n", nullDisplay(string(ctx.CurrentFile)))
	fmt.Fprintf(&b, "  last_error:    %s\n", nullDisplay(string(ctx.LastError)))
	fmt.Fprintf(&b, "  test_state:    %s\n", nullDisplay(string(ctx.TestState)))

	return b.String()
}
