package main

import (
	"bytes"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/harshagw/viri/cmd/debugger/tui"
	"github.com/harshagw/viri/internal/compiler"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/parser"
	"github.com/harshagw/viri/internal/scanner"
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
	bytecode, err := compile(string(source), filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Compilation error: %v\n", err)
		os.Exit(1)
	}

	// Create debugger
	debugger := NewDebugger(bytecode)

	// Create TUI model
	model := tui.NewModel(debugger)

	// Run TUI
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

func compile(source, filename string) (*compiler.Bytecode, error) {
	handler := &errorHandler{hasErrors: false}

	// Scan
	buf := []byte(source)
	reader := bytes.NewBuffer(buf)
	sc := scanner.New(reader, &filename)
	tokens, err := sc.Scan()
	if err != nil {
		return nil, fmt.Errorf("scan error: %w", err)
	}

	// Parse
	p := parser.NewParser(tokens, handler)
	p.SetFilePath(filename)
	module, err := p.Parse()
	if err != nil || handler.hasErrors {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	// Compile
	comp := compiler.New(handler)
	stmts := module.GetAllStatements()
	for _, stmt := range stmts {
		if err := comp.Compile(stmt); err != nil {
			return nil, fmt.Errorf("compile error: %w", err)
		}
	}

	if handler.hasErrors {
		return nil, fmt.Errorf("compilation failed")
	}

	return comp.Bytecode(), nil
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

