package compiler

import (
	"errors"
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
func (c *Compiler) checkProgram() (map[ast.Expr]types.Type, error) {
	checked := make(map[string]*types.Module, len(c.moduleOrder))
	exprTypes := make(map[ast.Expr]types.Type)
	ok := true

	for _, path := range c.moduleOrder {
		mod := c.modules[path]

		imports, err := c.importedModuleTypes(mod, path, checked)
		if err != nil {
			return nil, err
		}

		ck := checker.New(c.diagnosticHandler)
		moduleType, moduleOK := ck.CheckModule(mod, imports)
		checked[path] = moduleType
		ok = ok && moduleOK

		for expr, t := range ck.ExprTypes() {
			exprTypes[expr] = t
		}
	}

	if !ok {
		return nil, ErrTypeCheck
	}
	return exprTypes, nil
}

// importedModuleTypes maps each of a module's import aliases to the types of
// what that import exports.
func (c *Compiler) importedModuleTypes(
	mod *ast.Module, path string, checked map[string]*types.Module,
) (map[string]*types.Module, error) {
	imports := make(map[string]*types.Module, len(mod.Imports))

	for _, importStmt := range mod.Imports {
		importPath, ok := importStmt.Path.Literal.(string)
		if !ok {
			continue // the parser already rejected this
		}

		if stdlib.IsStdLib(importPath) {
			imports[importStmt.Alias.Lexeme] = stdlib.ModuleType(importPath)
			continue
		}

		targetPath, err := parser.ResolveModulePath(filepath.Dir(path), importPath)
		if err != nil {
			return nil, err
		}
		if target, ok := checked[targetPath]; ok {
			imports[importStmt.Alias.Lexeme] = target
		}
	}
	return imports, nil
}
