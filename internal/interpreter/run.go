package interpreter

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/harshagw/viri/internal/interpreter/ast"
	"github.com/harshagw/viri/internal/interpreter/interp"
	"github.com/harshagw/viri/internal/interpreter/objects"
	"github.com/harshagw/viri/internal/interpreter/parser"
	"github.com/harshagw/viri/internal/interpreter/token"
)

// Options controls a single run.
type Options struct {
	DebugMode      bool
	StatsMode      bool
	DisableWarning bool
}

// diagnostics receives parse and resolve diagnostics and prints them. It also
// records whether any error was seen, which is what stops the program before
// it executes.
type diagnostics struct {
	disableWarning bool
	hadError       bool
}

var _ objects.DiagnosticHandler = (*diagnostics)(nil)

func (d *diagnostics) Error(tok token.Token, message string) {
	if tok.FilePath != nil {
		color.New(color.FgRed).Fprintf(color.Error, "Error in %s at line %d: %s\n", *tok.FilePath, tok.Line, message)
	} else {
		color.New(color.FgRed).Fprintf(color.Error, "Error at line %d: %s\n", tok.Line, message)
	}
	d.hadError = true
}

func (d *diagnostics) Warn(tok token.Token, message string) {
	if d.disableWarning {
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

// Run parses, resolves and executes the program at filePath, printing any
// diagnostics and runtime errors itself. It reports whether the run failed.
func Run(filePath string, opts Options) (failed bool) {
	handler := &diagnostics{disableWarning: opts.DisableWarning}

	mod, err := parser.LoadModuleFile(filePath, handler)
	if err != nil {
		// Per-token diagnostics were already printed through the handler;
		// only surface errors that carry extra information (e.g. scanning
		// or file-loading failures).
		if !handler.hadError {
			color.New(color.FgRed).Fprintln(color.Error, "Error parsing module:", err)
		}
		return true
	}

	if handler.hadError {
		return true
	}

	if opts.DebugMode {
		printer := ast.NewPrinter()
		fmt.Println(printer.PrintStatements(mod.GetAllStatements()))
	}

	res := parser.NewResolver(handler)
	locals, err := res.Resolve(mod)
	if err != nil || handler.hadError {
		return true
	}

	interpreter := interp.NewInterpreter(nil)
	interpreter.SetLocals(locals)
	interpreter.SetResolvedModules(res.GetResolvedModules())
	interpreter.SetCurrentModule(mod.Path)

	startTime := time.Now()
	if _, err := interpreter.Interpret(mod.GetAllStatements()); err != nil {
		if runtimeErr, ok := err.(*objects.RuntimeError); ok {
			errPath := ""
			line := 0
			if runtimeErr.Token != nil {
				line = runtimeErr.Token.Line
				if runtimeErr.Token.FilePath != nil {
					errPath = *runtimeErr.Token.FilePath
				}
			}
			printRuntimeError(errPath, line, runtimeErr.Message)
		} else {
			printRuntimeError("", 0, err.Error())
		}
		return true
	}

	elapsed := time.Since(startTime)
	if opts.StatsMode {
		fmt.Printf("Time taken: %s\n", elapsed)
	}
	return false
}
