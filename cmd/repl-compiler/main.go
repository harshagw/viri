package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	prompt "github.com/c-bata/go-prompt"
	figure "github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/compiler"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/parser"
	"github.com/harshagw/viri/internal/scanner"
	"github.com/harshagw/viri/internal/token"
	"github.com/harshagw/viri/internal/vm"
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
	fmt.Printf("\n%s\n\n(type :quit to exit)\n\n", banner)

	handler := &replHandler{disableWarning: !showWarning}

	// Persist symbol table and globals across REPL inputs
	symbolTable := compiler.NewSymbolTable()
	globals := make([]objects.Object, vm.GlobalsSize)

	executor := func(line string) {
		if strings.TrimSpace(line) == "" {
			return
		}
		if strings.TrimSpace(line) == ":quit" {
			fmt.Println("bye")
			os.Exit(0)
		}

		code := bytes.NewBufferString(line + "\n")
		handler.hasErrors = false

		replPath := "<repl>"
		sc := scanner.New(code, &replPath)
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

		comp := compiler.NewWithState(handler, symbolTable)
		err = comp.Compile(newStmts[0])
		if err != nil {
			color.New(color.FgRed).Fprintf(color.Error, "Compilation error: %v\n", err)
			return
		}

		program := comp.Result()

		if debugMode {
			fmt.Println(program.Modules[0].Instructions.String())
			fmt.Println(program.Constants)
		}

		machine := vm.New(program)
		for i, g := range globals {
			if i < len(machine.GetModuleGlobals(0)) {
				machine.GetModuleGlobals(0)[i] = g
			}
		}
		err = machine.RunProgram()
		if err != nil {
			color.New(color.FgRed).Fprintf(color.Error, "Runtime error: %v\n", err)
			return
		}

		if _, isPrint := newStmts[0].(*ast.PrintStmt); !isPrint {
			fmt.Println(machine.LastPoppedStackElem().Inspect())
		}

		// Save globals for next execution
		globals = machine.GetModuleGlobals(0)
	}

	completer := func(d prompt.Document) []prompt.Suggest {
		return []prompt.Suggest{}
	}

	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix("> "),
		prompt.OptionTitle("Viri REPL"),
	)
	p.Run()
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
