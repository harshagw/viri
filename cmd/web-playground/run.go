// The playground pipeline: scan, parse, check, compile, run.
//
// It carries no build tag, so the behaviour the site depends on can be tested
// without a browser. main.go is the WASM entry point that calls into it.
package main

import (
	"bytes"
	"fmt"

	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/checker"
	"github.com/harshagw/viri/internal/compiler"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/parser"
	"github.com/harshagw/viri/internal/scanner"
	"github.com/harshagw/viri/internal/stdlib"
	"github.com/harshagw/viri/internal/token"
	"github.com/harshagw/viri/internal/types"
	"github.com/harshagw/viri/internal/vm"
)

const playgroundPath = "<playground>"

// Response is the JSON contract the playground front end reads.
type Response struct {
	Result   string   `json:"result"`
	Output   string   `json:"output"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

// execute runs one program and returns the value of its final statement, when
// that statement was an expression that produced one.
func execute(source string, out *bytes.Buffer, handler *playgroundHandler) string {
	if source == "" {
		return ""
	}

	path := playgroundPath
	tokens, err := scanner.New(bytes.NewBufferString(source), &path).Scan()
	if err != nil {
		handler.errors = append(handler.errors, fmt.Sprintf("Scanning error: %v", err))
		return ""
	}

	p := parser.NewParser(tokens, handler)
	p.SetFilePath(playgroundPath)
	mod, err := p.Parse()
	if err != nil || handler.hasErrors {
		return ""
	}

	statements := mod.Statements

	// Type-check before compiling: codegen assumes a well-typed program, so an
	// unchecked one would emit instructions the VM cannot run.
	ck := checker.New(handler)
	symbols := compiler.NewSymbolTable()
	if !declareImports(mod.Imports, ck, symbols, handler) {
		return ""
	}
	if !ck.CheckIncremental(statements) || handler.hasErrors {
		return ""
	}

	comp := compiler.NewWithState(handler, symbols)
	comp.SetTypeInfo(ck.ExprTypes(), ck.ClassTypes())
	for _, stmt := range statements {
		if err := comp.Compile(stmt); err != nil {
			handler.errors = append(handler.errors, fmt.Sprintf("Compilation error: %v", err))
			return ""
		}
	}
	if handler.hasErrors {
		return ""
	}

	machine := vm.New(comp.Result())
	machine.SetStdout(out)
	if err := machine.RunProgram(); err != nil {
		if vmErr, ok := err.(*objects.VMRuntimeError); ok && vmErr.Line > 0 {
			handler.errors = append(handler.errors,
				fmt.Sprintf("Line %d: %s", vmErr.Line, vmErr.Message))
		} else {
			handler.errors = append(handler.errors, fmt.Sprintf("Runtime error: %v", err))
		}
		return ""
	}

	return finalValue(statements, ck.ExprTypes(), machine)
}

// declareImports registers the program's imports with both the checker and the
// symbol table, and reports whether they were all usable.
//
// std: modules are built in, so they work here. A file import has nowhere to
// resolve to — the playground is a single buffer with no filesystem — so it is
// reported rather than silently dropped, which is what this used to do.
func declareImports(
	imports []*ast.ImportStmt,
	ck *checker.Checker,
	symbols *compiler.SymbolTable,
	handler *playgroundHandler,
) bool {
	ok := true
	for _, imp := range imports {
		path, isString := imp.Path.Literal.(string)
		if !isString {
			continue // already reported by the parser
		}
		if !stdlib.IsStdLib(path) {
			handler.Error(*imp.Path, "Only std: imports are available in the playground.")
			ok = false
			continue
		}
		moduleType := stdlib.ModuleType(path)
		names, known := stdlib.GetExportNameSet(path)
		if moduleType == nil || !known {
			handler.Error(*imp.Path, fmt.Sprintf("Unknown standard library module '%s'.", path))
			ok = false
			continue
		}
		ck.DeclareImport(imp.Alias.Lexeme, moduleType)
		symbols.DefineStdlibImport(imp.Alias.Lexeme, path, names)
	}
	return ok
}

// finalValue reports the value the program ended on, so a playground session
// can end in a bare expression and still show something.
//
// Only an expression statement qualifies. An expression statement's value is
// popped and discarded, so the last thing popped is exactly that value — but
// that is only true when the final statement was one: `print x;` also pops,
// and reporting what it printed as a result would be wrong.
func finalValue(statements []ast.Stmt, exprTypes map[ast.Expr]types.Type, machine *vm.VM) string {
	if len(statements) == 0 {
		return ""
	}
	last, ok := statements[len(statements)-1].(*ast.ExprStmt)
	if !ok {
		return ""
	}
	// A call to a function that returns nothing leaves no value to show.
	if t, known := exprTypes[last.Expr]; !known || t == types.Void {
		return ""
	}
	if value := machine.LastPoppedStackElem(); value != nil {
		return value.Inspect()
	}
	return ""
}

type playgroundHandler struct {
	hasErrors bool
	errors    []string
	warnings  []string
}

var _ objects.DiagnosticHandler = (*playgroundHandler)(nil)

func (h *playgroundHandler) Error(tok token.Token, msg string) {
	h.errors = append(h.errors, fmt.Sprintf("Line %d: %s", tok.Line, msg))
	h.hasErrors = true
}

func (h *playgroundHandler) Warn(tok token.Token, msg string) {
	h.warnings = append(h.warnings, fmt.Sprintf("Line %d: %s", tok.Line, msg))
}
