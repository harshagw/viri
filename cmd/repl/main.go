package main

import (
	"bufio"
	"bytes"
	"fmt"
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
	fmt.Println("  Press Enter twice (empty line) - Execute the code")
	fmt.Println("  Type :quit and Enter           - Exit the REPL")
	fmt.Println()

	interpreter := interp.NewInterpreter(nil)
	var programStmts []ast.Stmt
	handler := &replHandler{disableWarning: !showWarning}

	reader := bufio.NewReader(os.Stdin)
	var inputBuffer []string

	for {
		// Show appropriate prompt
		prompt := "> "
		if len(inputBuffer) > 0 {
			prompt = "... "
		}
		fmt.Print(prompt)

		// Read a line of input
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("\nbye")
			return
		}

		// Remove trailing newline
		line = strings.TrimSuffix(line, "\n")
		line = strings.TrimSuffix(line, "\r")

		// Handle :quit command
		if strings.TrimSpace(line) == ":quit" {
			fmt.Println("bye")
			return
		}

		// Check if line is empty
		if strings.TrimSpace(line) == "" {
			if len(inputBuffer) > 0 {
				// Empty line with accumulated input - execute it
				code := strings.Join(inputBuffer, "\n")
				executeCode(code, &programStmts, interpreter, handler, debugMode)
				inputBuffer = nil
			}
			// Empty line without accumulated input - just continue
			continue
		}

		// Add line to buffer
		inputBuffer = append(inputBuffer, line)
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
