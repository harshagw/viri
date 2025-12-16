package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	figure "github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/interp"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/parser"
	"github.com/harshagw/viri/internal/scanner"
	"github.com/harshagw/viri/internal/token"
	"golang.org/x/term"
)

func main() {
	debugMode := false
	showWarning := false

	for _, arg := range os.Args[1:] {
		switch arg {
		case "--debug":
			debugMode = true
		case "--no-warning":
			showWarning = false
		}
	}

	banner := figure.NewFigure("Viri", "", true).String()
	fmt.Printf("\n%s\n\n", banner)
	fmt.Println("Multi-line REPL:")
	fmt.Println("  Press Enter  - Add a new line")
	fmt.Println("  Press Ctrl+D - Execute the accumulated code")
	fmt.Println("  Type :quit   - Exit the REPL")
	fmt.Println()

	interpreter := interp.NewInterpreter(nil)
	var programStmts []ast.Stmt
	handler := &replHandler{disableWarning: !showWarning}

	// Check if stdin is a terminal
	isTerminal := term.IsTerminal(int(os.Stdin.Fd()))
	
	if isTerminal {
		// Terminal mode with Ctrl+D support
		runTerminalMode(&programStmts, interpreter, handler, debugMode)
	} else {
		// Non-terminal mode (for testing/scripting)
		runNonTerminalMode(&programStmts, interpreter, handler, debugMode)
	}
}

func runTerminalMode(programStmts *[]ast.Stmt, interpreter *interp.Interpreter, handler *replHandler, debugMode bool) {
	// Set up terminal for raw input
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set terminal to raw mode: %v\n", err)
		os.Exit(1)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	terminal := term.NewTerminal(os.Stdin, "> ")
	var inputBuffer []string

	for {
		// Set appropriate prompt
		if len(inputBuffer) > 0 {
			terminal.SetPrompt("... ")
		} else {
			terminal.SetPrompt("> ")
		}

		// Read a line of input
		line, err := terminal.ReadLine()
		if err != nil {
			if err == io.EOF {
				// Ctrl+D pressed
				if len(inputBuffer) > 0 {
					// Execute the accumulated code
					code := strings.Join(inputBuffer, "\n")
					fmt.Println() // Add newline after Ctrl+D
					executeCode(code, programStmts, interpreter, handler, debugMode)
					inputBuffer = inputBuffer[:0]
				} else {
					// Ctrl+D on empty buffer - exit
					fmt.Println("\nbye")
					return
				}
				continue
			}
			fmt.Println("\nbye")
			return
		}

		// Handle :quit command
		if strings.TrimSpace(line) == ":quit" {
			fmt.Println("bye")
			return
		}

		// Add line to buffer
		inputBuffer = append(inputBuffer, line)
	}
}

func runNonTerminalMode(programStmts *[]ast.Stmt, interpreter *interp.Interpreter, handler *replHandler, debugMode bool) {
	// Simple line-by-line mode for testing (without Ctrl+D detection)
	// Uses bufio.Scanner for efficient line-by-line reading
	scanner := bufio.NewScanner(os.Stdin)
	var inputBuffer []string
	
	for scanner.Scan() {
		line := scanner.Text()
		
		// Handle :quit command
		if strings.TrimSpace(line) == ":quit" {
			fmt.Println("bye")
			return
		}
		
		// Empty line triggers execution
		if strings.TrimSpace(line) == "" {
			if len(inputBuffer) > 0 {
				code := strings.Join(inputBuffer, "\n")
				executeCode(code, programStmts, interpreter, handler, debugMode)
				inputBuffer = inputBuffer[:0]
			}
			continue
		}
		
		// Accumulate non-empty lines
		inputBuffer = append(inputBuffer, line)
	}
	
	// Handle any scanner errors
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
	}
	
	// Execute any remaining accumulated code
	if len(inputBuffer) > 0 {
		code := strings.Join(inputBuffer, "\n")
		executeCode(code, programStmts, interpreter, handler, debugMode)
	}
}

func executeCode(code string, programStmts *[]ast.Stmt, interpreter *interp.Interpreter, handler *replHandler, debugMode bool) {
	if strings.TrimSpace(code) == "" {
		return
	}

	codeBuffer := bytes.NewBufferString(code + "\n")
	handler.hasErrors = false

	sc := scanner.New(codeBuffer, nil)
	tokens, err := sc.Scan()
	if err != nil {
		fmt.Println("Error parsing tokens:", err)
		return
	}

	p := parser.NewParser(tokens, handler)
	p.SetFilePath("<repl>")
	lineModule, err := p.Parse()
	if err != nil || handler.hasErrors {
		return
	}

	if debugMode {
		pr := ast.NewPrinter()
		fmt.Println(pr.PrintStatements(lineModule.GetAllStatements()))
	}

	newStmts := lineModule.GetAllStatements()
	*programStmts = append(*programStmts, newStmts...)
	replModule := ast.NewModule("<repl>", nil, *programStmts)

	res := parser.NewResolver(handler)
	locals, err := res.Resolve(replModule)
	if err != nil || handler.hasErrors {
		*programStmts = (*programStmts)[:len(*programStmts)-len(newStmts)]
		return
	}

	interpreter.SetLocals(locals)
	results, err := interpreter.Interpret(newStmts)
	if err != nil {
		color.New(color.FgRed).Fprintf(color.Error, "Runtime error: %v\n", err)
		*programStmts = (*programStmts)[:len(*programStmts)-len(newStmts)]
		return
	}
	for idx, result := range results {
		if result == nil {
			continue
		}
		if _, ok := newStmts[idx].(*ast.PrintStmt); ok {
			continue
		}
		fmt.Println(result.Inspect())
	}
}

type replHandler struct {
	disableWarning bool
	hasErrors      bool
}

var _ objects.DiagnosticHandler = (*replHandler)(nil)

func (h *replHandler) Error(tok token.Token, msg string) {
	color.New(color.FgRed).Fprintf(color.Error, "Error at line %d: %s\n", tok.Line, msg)
	h.hasErrors = true
}

func (h *replHandler) Warn(tok token.Token, msg string) {
	if h.disableWarning {
		return
	}
	color.New(color.FgYellow).Fprintf(color.Error, "Warning at line %d: %s\n", tok.Line, msg)
}
