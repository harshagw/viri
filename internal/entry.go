package internal

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/compiler"
	"github.com/harshagw/viri/internal/interp"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/parser"
	"github.com/harshagw/viri/internal/token"
	"github.com/harshagw/viri/internal/vm"
)

type ViriRuntimeConfig struct {
	DebugMode      bool
	DisableWarning bool
	Engine         string // "interpreter" or "vm"
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
	if tok.FilePath != nil {
		color.New(color.FgRed).Fprintf(color.Error, "Error in %s at line %d: %s\n", *tok.FilePath, tok.Line, message)
	} else {
		color.New(color.FgRed).Fprintf(color.Error, "Error at line %d: %s\n", tok.Line, message)
	}
	v.hasErrors = true
}

func (v *Viri) Warn(tok token.Token, message string) {
	if v.config.DisableWarning {
		return
	}
	if tok.FilePath != nil {
		color.New(color.FgYellow).Fprintf(color.Error, "Warning in %s at line %d: %s\n", *tok.FilePath, tok.Line, message)
	} else {
		color.New(color.FgYellow).Fprintf(color.Error, "Warning at line %d: %s\n", tok.Line, message)
	}
}

func (v *Viri) Run(filePath string) {
	if v.config.Engine == "vm" {
		v.runWithVM(filePath)
	} else {
		v.runWithInterpreter(filePath)
	}
}

func (v *Viri) runWithVM(filePath string) {
	mod, err := parser.LoadModuleFile(filePath, v)
	if err != nil {
		fmt.Println("Error parsing module:", err)
		v.hasErrors = true
		return
	}

	if v.hasErrors {
		return
	}

	if v.config.DebugMode {
		printer := ast.NewPrinter()
		tree := printer.PrintStatements(mod.GetAllStatements())
		fmt.Println(tree)
	}

	comp := compiler.New(v)
	for _, stmt := range mod.GetAllStatements() {
		if err := comp.Compile(stmt); err != nil {
			if !v.hasErrors {
				color.New(color.FgRed).Fprintln(color.Error, "Compilation error:", err)
			}
			v.hasErrors = true
			return
		}
	}

	machine := vm.New(comp.Bytecode())
	if err := machine.Run(); err != nil {
		color.New(color.FgRed).Fprintln(color.Error, "Runtime error:", err)
		v.hasErrors = true
		return
	}
}

func (v *Viri) runWithInterpreter(filePath string) {
	mod, err := parser.LoadModuleFile(filePath, v)
	if err != nil {
		fmt.Println("Error parsing module:", err)
		v.hasErrors = true
		return
	}

	if v.hasErrors {
		return
	}

	if v.config.DebugMode {
		printer := ast.NewPrinter()
		tree := printer.PrintStatements(mod.GetAllStatements())
		fmt.Println(tree)
	}

	res := parser.NewResolver(v)
	locals, err := res.Resolve(mod)
	if err != nil {
		v.hasErrors = true
		return
	}

	interpreter := interp.NewInterpreter(nil)
	interpreter.SetLocals(locals)
	interpreter.SetResolvedModules(res.GetResolvedModules())
	interpreter.SetCurrentModule(mod.Path)

	if _, err := interpreter.Interpret(mod.GetAllStatements()); err != nil {
		if runtimeErr, ok := err.(*objects.RuntimeError); ok {
			if runtimeErr.Token != nil {
				if runtimeErr.Token.FilePath != nil {
					color.New(color.FgRed).Fprintf(color.Error, "Runtime error in %s at line %d: %s\n", *runtimeErr.Token.FilePath, runtimeErr.Token.Line, runtimeErr.Message)
				} else {
					color.New(color.FgRed).Fprintf(color.Error, "Runtime error at line %d: %s\n", runtimeErr.Token.Line, runtimeErr.Message)
				}
			} else {
				color.New(color.FgRed).Fprintln(color.Error, "Runtime error:", runtimeErr.Message)
			}
		} else {
			fmt.Println("Runtime error:", err)
		}
		v.hasErrors = true
		return
	}
}
