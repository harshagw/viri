package interp

import (
	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/objects"
)

func (i *Interpreter) ExecuteModule(astMod *ast.Module, importStmt *ast.ImportStmt) (*objects.Module, error) {
	moduleEnv := objects.NewEnvironment(i.globals)

	previousEnv := i.environment
	previousExports := i.moduleExports
	previousModule := i.currentModule

	i.environment = moduleEnv
	i.moduleExports = make(map[string]objects.Object)
	i.currentModule = astMod.Path

	defer func() {
		i.environment = previousEnv
		i.moduleExports = previousExports
		i.currentModule = previousModule
	}()

	for _, stmt := range astMod.GetAllStatements() {
		if _, err := i.evalStmt(stmt); err != nil {
			return nil, err
		}
	}

	runtimeMod := objects.NewModule(astMod.Path, astMod.Imports, astMod.Statements)
	runtimeMod.Exports = i.moduleExports
	runtimeMod.Namespace = objects.NewNamespace(importStmt.Alias.Lexeme, i.moduleExports)

	return runtimeMod, nil
}
