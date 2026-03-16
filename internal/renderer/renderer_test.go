package renderer

import (
	"strings"
	"testing"
)

func TestBold(t *testing.T) {
	result := Bold("hello")
	if !strings.Contains(result, "hello") {
		t.Error("Bold should contain the original text")
	}
	if !strings.HasPrefix(result, "\033[1m") {
		t.Error("Bold should start with ANSI bold escape")
	}
	if !strings.HasSuffix(result, "\033[0m") {
		t.Error("Bold should end with ANSI reset escape")
	}
}

func TestSymbols(t *testing.T) {
	if SymbolDone != "✓" {
		t.Errorf("SymbolDone = %q, want ✓", SymbolDone)
	}
	if SymbolTodo != "○" {
		t.Errorf("SymbolTodo = %q, want ○", SymbolTodo)
	}
}
