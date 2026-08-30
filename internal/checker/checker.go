// Package checker type-checks a module before it is compiled.
//
// It runs as a separate pass between parsing and codegen, the shape used by
// Go, TypeScript, clang and javac. Two walks per module:
//
//  1. collect — read every top-level signature from its annotations, so
//     functions and classes can refer to each other in any order.
//  2. check — walk every body, giving each expression a type and checking it
//     against what the declarations promised.
//
// Nothing here emits code. The result is a verdict plus a side table of
// expression types, which Phase 3 and 4 read to pick narrower instructions.
package checker

import (
	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/token"
	"github.com/harshagw/viri/internal/types"
)

// Checker checks one module at a time, carrying the module types of everything
// that module imports.
type Checker struct {
	handler objects.DiagnosticHandler
	scopes  []*scope

	// exprTypes records the type of every expression the checker synthesises.
	// It is the payoff of the pass: Phase 3 reads it to lay out field access,
	// Phase 4 to choose typed arithmetic opcodes.
	exprTypes map[ast.Expr]types.Type

	// classes holds every class declared in the module being checked, so a
	// NamedType can resolve to it.
	classes map[string]*types.Class

	// imports maps an import alias to the module it names.
	imports map[string]*types.Module

	// Context for the body walk.
	currentFn    *types.Function // signature of the function being checked
	currentClass *types.Class    // class whose method is being checked
	inInit       bool

	hadError bool
}

// New creates a checker reporting through handler.
func New(handler objects.DiagnosticHandler) *Checker {
	return &Checker{
		handler:   handler,
		exprTypes: make(map[ast.Expr]types.Type),
		classes:   make(map[string]*types.Class),
		imports:   make(map[string]*types.Module),
	}
}

// ExprTypes returns the synthesised type of every expression checked so far.
func (c *Checker) ExprTypes() map[ast.Expr]types.Type {
	return c.exprTypes
}

// CheckModule checks one module. imports maps each import alias used by the
// module to the already-checked module it names, so dependencies must be
// checked first — which the compiler's post-order module list gives for free.
//
// It returns the module's own exported types, and whether checking succeeded.
func (c *Checker) CheckModule(mod *ast.Module, imports map[string]*types.Module) (*types.Module, bool) {
	c.hadError = false
	c.classes = make(map[string]*types.Class)
	c.imports = imports
	c.scopes = []*scope{newScope()}

	c.declareNatives()

	statements := mod.Statements
	c.collect(statements)
	c.checkStatements(statements)

	return c.buildModuleType(mod), !c.hadError
}

// declareNatives binds the globally available native functions. They are not
// imported, so they live in the outermost scope alongside the builtin types.
func (c *Checker) declareNatives() {
	c.declare("clock", &types.Function{Params: nil, Return: types.Number}, true)
	// len accepts a string, an array, or a map. Viri has no way to write that
	// signature yet, so it is special-cased in checkCall.
	c.declare("len", nativeLen, true)
}

// nativeLen is a sentinel: calls to it are checked by hand because its
// parameter is not expressible in the type system.
var nativeLen = &types.Function{Params: []types.Type{types.Invalid}, Return: types.Number}

// buildModuleType gathers the module's exported names and their types.
func (c *Checker) buildModuleType(mod *ast.Module) *types.Module {
	out := &types.Module{
		Path:    mod.Path,
		Exports: make(map[string]types.Type),
		Classes: make(map[string]*types.Class),
	}
	for _, stmt := range mod.Statements {
		var name string
		switch s := stmt.(type) {
		case *ast.VarDeclStmt:
			if s.Exported {
				name = s.Name.Lexeme
			}
		case *ast.FunctionStmt:
			if s.Exported {
				name = s.Name.Lexeme
			}
		case *ast.ClassStmt:
			if s.Exported {
				name = s.Name.Lexeme
			}
		}
		if name == "" {
			continue
		}
		if b, ok := c.lookup(name); ok {
			out.Exports[name] = b.typ
		}
		// An exported class is also a name importers can use as a type.
		if class, ok := c.classes[name]; ok {
			out.Classes[name] = class
		}
	}
	return out
}

// --- diagnostics -----------------------------------------------------------

func (c *Checker) error(node ast.Node, message string) {
	c.hadError = true
	if c.handler == nil {
		return
	}
	if tok := node.GetPrimaryToken(); tok != nil {
		c.handler.Error(*tok, message)
		return
	}
	c.handler.Error(token.Token{}, message)
}

func (c *Checker) errorAt(tok *token.Token, message string) {
	c.hadError = true
	if c.handler == nil {
		return
	}
	if tok != nil {
		c.handler.Error(*tok, message)
		return
	}
	c.handler.Error(token.Token{}, message)
}
