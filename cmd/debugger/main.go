package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/harshagw/viri/cmd/debugger/tui"
	"github.com/harshagw/viri/internal/compiler"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/token"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <file.viri>\n", os.Args[0])
		os.Exit(1)
	}

	filename := os.Args[1]

	// Read source file
	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Compile the source
	program, err := compile(string(source), filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Compilation error: %v\n", err)
		os.Exit(1)
	}

	// Create debugger
	debugger := NewDebugger(program)

	// Create TUI model
	model := tui.NewModel(debugger)

	// Run TUI
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

func compile(source, filename string) (*objects.CompiledProgram, error) {
	handler := &errorHandler{hasErrors: false}

	// Use Compiler with full module support
	comp := compiler.New(handler)
	program, err := comp.CompileProgram(filename)
	if err != nil || handler.hasErrors {
		return nil, fmt.Errorf("compilation failed: %w", err)
	}

	if len(program.Modules) == 0 {
		return nil, fmt.Errorf("no modules compiled")
	}

	return program, nil
}

type errorHandler struct {
	hasErrors bool
}

var _ objects.DiagnosticHandler = (*errorHandler)(nil)

func (h *errorHandler) Error(tok token.Token, msg string) {
	fmt.Fprintf(os.Stderr, "Error at line %d: %s\n", tok.Line, msg)
	h.hasErrors = true
}

func (h *errorHandler) Warn(tok token.Token, msg string) {
	fmt.Fprintf(os.Stderr, "Warning at line %d: %s\n", tok.Line, msg)
}
