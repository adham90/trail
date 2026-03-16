package renderer

import "fmt"

const (
	SymbolDone = "✓"
	SymbolTodo = "○"
)

// Bold wraps text in ANSI bold.
func Bold(s string) string {
	return fmt.Sprintf("\033[1m%s\033[0m", s)
}
