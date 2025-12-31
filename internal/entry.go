package internal

import (
	"fmt"
	"time"

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
	StatsMode      bool
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
			StatsMode:      false,
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

func printRuntimeError(filePath string, line int, message string) {
	if filePath != "" && line > 0 {
		color.New(color.FgRed).Fprintf(color.Error, "Runtime error in %s at line %d: %s\n", filePath, line, message)
	} else if line > 0 {
		color.New(color.FgRed).Fprintf(color.Error, "Runtime error at line %d: %s\n", line, message)
	} else {
		color.New(color.FgRed).Fprintln(color.Error, "Runtime error:", message)
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
	comp := compiler.New(v)
	program, err := comp.CompileProgram(filePath)
	if err != nil {
		if !v.hasErrors {
			color.New(color.FgRed).Fprintln(color.Error, "Compilation error:", err)
		}
		v.hasErrors = true
		return
	}

	if v.hasErrors {
		return
	}

	if v.config.DebugMode {
		for i, compiledMod := range program.Modules {
			fmt.Printf("Module %d:\n", i)
			fmt.Println(compiledMod.Instructions.String())
		}
	}

	startTime := time.Now()
	machine := vm.New(program)
	if err := machine.RunProgram(); err != nil {
		if vmErr, ok := err.(*objects.VMRuntimeError); ok {
			printRuntimeError(vmErr.FilePath, vmErr.Line, vmErr.Message)
		} else {
			printRuntimeError("", 0, err.Error())
		}
		v.hasErrors = true
		return
	}

	elapsed := time.Since(startTime)
	if v.config.StatsMode {
		fmt.Printf("Time taken: %s\n", elapsed)
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

	startTime := time.Now()
	if _, err := interpreter.Interpret(mod.GetAllStatements()); err != nil {
		if runtimeErr, ok := err.(*objects.RuntimeError); ok {
			filePath := ""
			line := 0
			if runtimeErr.Token != nil {
				line = runtimeErr.Token.Line
				if runtimeErr.Token.FilePath != nil {
					filePath = *runtimeErr.Token.FilePath
				}
			}
			printRuntimeError(filePath, line, runtimeErr.Message)
		} else {
			printRuntimeError("", 0, err.Error())
		}
		v.hasErrors = true
		return
	}
	elapsed := time.Since(startTime)
	if v.config.StatsMode {
		fmt.Printf("Time taken: %s\n", elapsed)
	}
}
