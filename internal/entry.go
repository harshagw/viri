package internal

import (
	"bytes"
	"fmt"

	"github.com/fatih/color"
	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/interp"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/parser"
	"github.com/harshagw/viri/internal/scanner"
	"github.com/harshagw/viri/internal/token"
)

type ViriRuntimeConfig struct {
	DebugMode      bool
	DisableWarning bool
}

type Viri struct {
	hasErrors bool
	config    *ViriRuntimeConfig
}

var _ objects.DiagnosticHandler = (*Viri)(nil)

func NewViriRuntime(config *ViriRuntimeConfig) *Viri {
	if config == nil {
		config = &ViriRuntimeConfig{
			DebugMode:      false,
			DisableWarning: false,
		}
	}
	return &Viri{
		hasErrors: false,
		config:    config,
	}
}

func (v *Viri) HasErrors() bool {
	return v.hasErrors
}

func (v *Viri) Error(tok token.Token, message string) {
	color.New(color.FgRed).Fprintf(color.Error, "Error at line %d: %s\n", tok.Line, message)
	v.hasErrors = true
}

func (v *Viri) Warn(tok token.Token, message string) {
	if v.config.DisableWarning {
		return
	}
	color.New(color.FgYellow).Fprintf(color.Error, "Warning at line %d: %s\n", tok.Line, message)
}

func (v *Viri) Run(bytes *bytes.Buffer) {
	sc := scanner.New(bytes)
	tokens, err := sc.Scan()
	if err != nil {
		fmt.Println("Error parsing tokens:", err)
		v.hasErrors = true
		return
	}

	p := parser.NewParser(tokens, v)
	statements, err := p.Parse()
	if err != nil {
		v.hasErrors = true
	}

	if v.hasErrors {
		return
	}

	if v.config.DebugMode {
		printer := ast.NewPrinter()
		tree := printer.PrintStatements(statements)
		fmt.Println(tree)
	}

	res := parser.NewResolver(v)
	locals, err := res.Resolve(statements)
	if err != nil {
		v.hasErrors = true
		return
	}

	interpreter := interp.NewInterpreter(nil)
	interpreter.SetLocals(locals)

	if err := interpreter.Interpret(statements); err != nil {
		fmt.Println("Error interpreting expression:", err)
		v.hasErrors = true
		return
	}
}
