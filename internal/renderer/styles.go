package renderer

import "fmt"

const (
	SymbolDone    = "✓"
	SymbolActive  = "▶"
	SymbolBlocked = "!"
	SymbolTodo    = "○"
)

// Bold wraps text in ANSI bold.
func Bold(s string) string {
	return fmt.Sprintf("\033[1m%s\033[0m", s)
}

// StatusSymbol returns the display symbol for a task status.
func StatusSymbol(status string) string {
	switch status {
	case "done":
		return SymbolDone
	case "active":
		return SymbolActive
	case "blocked":
		return SymbolBlocked
	case "todo":
		return SymbolTodo
	default:
		return "?"
	}
}
