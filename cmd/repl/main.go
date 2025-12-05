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
	"github.com/harshagw/viri/internal/interp"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/parser"
	"github.com/harshagw/viri/internal/scanner"
	"github.com/harshagw/viri/internal/token"
)

func main() {
	debugMode := false
	showWarning := true

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

	interpreter := interp.NewInterpreter(nil)
	var program []ast.Stmt
	handler := &replHandler{disableWarning: !showWarning}

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

		sc := scanner.New(code)
		tokens, err := sc.Scan()
		if err != nil {
			fmt.Println("Error parsing tokens:", err)
			return
		}

		p := parser.NewParser(tokens, handler)
		stmts, err := p.Parse()
		if err != nil || handler.hasErrors {
			return
		}

		if debugMode {
			pr := ast.NewPrinter()
			fmt.Println(pr.PrintStatements(stmts))
		}

		program = append(program, stmts...)
		res := parser.NewResolver(handler)
		locals, err := res.Resolve(program)
		if err != nil || handler.hasErrors {
			program = program[:len(program)-len(stmts)]
			return
		}

		interpreter.SetLocals(locals)
		if err := interpreter.Interpret(stmts); err != nil {
			color.New(color.FgRed).Fprintf(color.Error, "Runtime error: %v\n", err)
			program = program[:len(program)-len(stmts)]
			return
		}
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
