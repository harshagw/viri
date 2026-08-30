package compiler

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/checker"
	"github.com/harshagw/viri/internal/parser"
	"github.com/harshagw/viri/internal/stdlib"
	"github.com/harshagw/viri/internal/types"
)

// ErrTypeCheck reports that the program did not type-check. The individual
// diagnostics have already gone to the handler, so this only signals that
// codegen must not run.
var ErrTypeCheck = errors.New("type error")

// checkProgram type-checks every loaded module before any of them is compiled.
//
// Modules are checked in c.moduleOrder, which loadModule fills depth-first
// post-order — dependencies first. That is exactly what the checker needs: a
// module's imports are already checked, so their exported types are available
// when it refers to them.
//
// Every module is checked even after one fails, so a single run reports every
// error in the program rather than only those in the first bad file.
func (c *Compiler) checkProgram() (map[ast.Expr]types.Type, map[*ast.ClassStmt]*types.Class, error) {
	checked := make(map[string]*types.Module, len(c.moduleOrder))
	exprTypes := make(map[ast.Expr]types.Type)
	classTypes := make(map[*ast.ClassStmt]*types.Class)
	ok := true

	for _, path := range c.moduleOrder {
		mod := c.modules[path]

		imports, importsOK, err := c.importedModuleTypes(mod, path, checked)
		if err != nil {
			return nil, nil, err
		}
		ok = ok && importsOK

		ck := checker.New(c.diagnosticHandler)
		moduleType, moduleOK := ck.CheckModule(mod, imports)
		checked[path] = moduleType
		ok = ok && moduleOK

		for expr, t := range ck.ExprTypes() {
			exprTypes[expr] = t
		}
		for decl, class := range ck.ClassTypes() {
			classTypes[decl] = class
		}
	}

	if !ok {
		return nil, nil, ErrTypeCheck
	}
	return exprTypes, classTypes, nil
}

// importedModuleTypes maps each of a module's import aliases to the types of
// what that import exports.
func (c *Compiler) importedModuleTypes(
	mod *ast.Module, path string, checked map[string]*types.Module,
) (map[string]*types.Module, bool, error) {
	imports := make(map[string]*types.Module, len(mod.Imports))
	ok := true

	for _, importStmt := range mod.Imports {
		importPath, ok := importStmt.Path.Literal.(string)
		if !ok {
			continue // the parser already rejected this
		}

		if stdlib.IsStdLib(importPath) {
			moduleType := stdlib.ModuleType(importPath)
			if moduleType == nil {
				// Report it here, where the import path token is to hand. The
				// alias stays unbound, so nothing downstream dereferences a
				// nil module.
				c.error(importStmt.Path, fmt.Sprintf("Unknown standard library module '%s'.", importPath))
				ok = false
				continue
			}
			imports[importStmt.Alias.Lexeme] = moduleType
			continue
		}

		targetPath, err := parser.ResolveModulePath(filepath.Dir(path), importPath)
		if err != nil {
			return nil, false, err
		}
		if target, ok := checked[targetPath]; ok {
			imports[importStmt.Alias.Lexeme] = target
		}
	}
	return imports, ok, nil
}

// fieldSlot reports the instance slot a property access resolves to, using the
// checker's record of what the object's type is.
//
// It answers false for a method — those still resolve by name at runtime,
// because they live on the class rather than in the instance — and for any
// expression the checker did not type, which happens only in the unit tests
// that compile a statement directly without running the checker.
func (c *Compiler) fieldSlot(object ast.Expr, name string) (int, bool) {
	if c.exprTypes == nil {
		return 0, false
	}
	class, ok := c.exprTypes[object].(*types.Class)
	if !ok {
		return 0, false
	}
	return class.FieldSlot(name)
}

// classFieldCount is the size of an instance's field storage, taken from the
// checker's layout for the class.
func (c *Compiler) classFieldCount(stmt *ast.ClassStmt) int {
	if c.classTypes == nil {
		return 0
	}
	if class, ok := c.classTypes[stmt]; ok {
		return class.NumFields()
	}
	return 0
}

// SetTypeInfo installs the checker's results for callers that drive checking
// themselves rather than going through CompileProgram — the REPL, which checks
// one line at a time against a scope that persists between them.
func (c *Compiler) SetTypeInfo(exprTypes map[ast.Expr]types.Type, classTypes map[*ast.ClassStmt]*types.Class) {
	c.exprTypes = exprTypes
	c.classTypes = classTypes
}
