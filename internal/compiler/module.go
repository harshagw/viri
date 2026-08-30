package compiler

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/parser"
	"github.com/harshagw/viri/internal/stdlib"
	"github.com/harshagw/viri/internal/token"
)

// CompileProgram compiles a program starting from the entry module
func (c *Compiler) CompileProgram(entryPath string) (*objects.CompiledProgram, error) {
	// Load all modules and build the initialization order: loadModule
	// records depth-first post-order, which puts dependencies first and is
	// deterministic — the same order the interpreter executes imports in.
	if err := c.loadModule(entryPath, []string{}); err != nil {
		return nil, err
	}

	// Assign module indices
	for i, path := range c.moduleOrder {
		c.moduleIndices[path] = i
	}

	// Type-check every module before compiling any of them. Codegen is
	// untyped and assumes a well-typed program, so nothing may be emitted
	// until the whole program checks.
	exprTypes, err := c.checkProgram()
	if err != nil {
		return nil, err
	}
	c.exprTypes = exprTypes

	// Compile all modules using shared constants table
	compiledModules := make([]objects.CompiledModule, len(c.moduleOrder))

	for i, path := range c.moduleOrder {
		mod, err := c.compileModule(path)
		if err != nil {
			return nil, err
		}
		compiledModules[i] = mod
	}

	return &objects.CompiledProgram{
		Modules:   compiledModules,
		Constants: c.constants, // shared constants table
		DebugInfo: c.debugInfo,
	}, nil
}

// compileModule compiles a single module using the shared constants table
func (c *Compiler) compileModule(path string) (objects.CompiledModule, error) {
	mod := c.modules[path]

	// Reset compiler state for this module
	c.reset(nil)
	c.currentModuleIdx = c.moduleIndices[path]

	c.SetFilePath(path)

	// Register imports - we need to know what each imported module exports
	for _, importStmt := range mod.Imports {
		importPath, ok := importStmt.Path.Literal.(string)
		if !ok {
			return objects.CompiledModule{}, fmt.Errorf("import path must be a string")
		}

		// Handle stdlib imports (e.g., "std:math")
		if stdlib.IsStdLib(importPath) {
			exportNames, exists := stdlib.GetExportNameSet(importPath)
			if !exists {
				return objects.CompiledModule{}, fmt.Errorf("unknown stdlib module: %s", importPath)
			}
			c.symbolTable.DefineStdlibImport(importStmt.Alias.Lexeme, importPath, exportNames)
			continue
		}

		targetPath, err := parser.ResolveModulePath(filepath.Dir(path), importPath)
		if err != nil {
			return objects.CompiledModule{}, err
		}

		moduleIdx := c.moduleIndices[targetPath]
		exportMap := c.buildExportMap(c.modules[targetPath])
		c.symbolTable.DefineImport(importStmt.Alias.Lexeme, moduleIdx, exportMap)
	}

	// Hoist all module-level declarations so function bodies can reference
	// globals declared later in the file (mutual recursion works). Each
	// symbol stays Pending until its declaration statement is compiled.
	hoistedSlots := make(map[string]int)
	for _, stmt := range mod.Statements {
		var nameTok *token.Token
		var isConst bool
		switch s := stmt.(type) {
		case *ast.VarDeclStmt:
			nameTok, isConst = s.Name, s.IsConst
		case *ast.FunctionStmt:
			nameTok = s.Name
		case *ast.ClassStmt:
			nameTok = s.Name
		}
		if nameTok == nil {
			continue
		}
		symbol, ok := c.symbolTable.Hoist(nameTok.Lexeme, isConst)
		if !ok {
			return objects.CompiledModule{}, c.error(nameTok, "Cannot declare variable with this name again.")
		}
		hoistedSlots[nameTok.Lexeme] = symbol.Index
	}

	// Track exports as we compile
	var exportNames []string

	// Compile all statements
	for _, stmt := range mod.Statements {
		// Check for exports before compiling
		var exported bool
		var exportName string

		switch s := stmt.(type) {
		case *ast.VarDeclStmt:
			exported = s.Exported
			exportName = s.Name.Lexeme
		case *ast.FunctionStmt:
			exported = s.Exported
			exportName = s.Name.Lexeme
		case *ast.ClassStmt:
			exported = s.Exported
			exportName = s.Name.Lexeme
		}

		if err := c.Compile(stmt); err != nil {
			return objects.CompiledModule{}, err
		}

		if exported {
			exportNames = append(exportNames, exportName)
		}
	}

	if c.overflowErr != nil {
		return objects.CompiledModule{}, c.overflowErr
	}

	// Build exports array (export index -> global slot)
	exports := make([]int, len(exportNames))
	for i, name := range exportNames {
		exports[i] = hoistedSlots[name]
	}

	// Add debug info for module-level code
	debugIdx := c.debugInfo.Add(c.currentLineTable(), path)

	return objects.CompiledModule{
		Instructions: c.currentInstructions(),
		NumGlobals:   c.maxGlobalIndex + 1,
		NumLocals:    c.symbolTable.NumBlockLocals(),
		Exports:      exports,
		DebugInfoIdx: debugIdx,
	}, nil
}

// buildExportMap builds a map of export names to export indices for a module
func (c *Compiler) buildExportMap(mod *ast.Module) map[string]int {
	exportMap := make(map[string]int)
	exportIdx := 0

	for _, stmt := range mod.Statements {
		var exported bool
		var exportName string

		switch s := stmt.(type) {
		case *ast.VarDeclStmt:
			exported = s.Exported
			exportName = s.Name.Lexeme
		case *ast.FunctionStmt:
			exported = s.Exported
			exportName = s.Name.Lexeme
		case *ast.ClassStmt:
			exported = s.Exported
			exportName = s.Name.Lexeme
		}

		if exported {
			exportMap[exportName] = exportIdx
			exportIdx++
		}
	}

	return exportMap
}

// loadModule loads a module and its dependencies, checking for cycles
func (c *Compiler) loadModule(path string, stack []string) error {
	if _, ok := c.modules[path]; ok {
		return nil // already loaded
	}

	if slices.Contains(stack, path) {
		return fmt.Errorf("circular dependency detected: %v -> %s", stack, path)
	}

	mod, err := parser.LoadModuleFile(path, c.diagnosticHandler)
	if err != nil {
		return err
	}

	c.modules[path] = mod

	// Load dependencies
	newStack := append(stack, path)
	for _, importStmt := range mod.Imports {
		importPath, ok := importStmt.Path.Literal.(string)
		if !ok {
			return fmt.Errorf("import path must be a string")
		}

		// Skip stdlib modules - they don't need to be loaded as files
		if stdlib.IsStdLib(importPath) {
			continue
		}

		targetPath, err := parser.ResolveModulePath(filepath.Dir(path), importPath)
		if err != nil {
			return err
		}

		if err := c.loadModule(targetPath, newStack); err != nil {
			return err
		}
	}

	// Post-order: all dependencies are recorded before this module
	c.moduleOrder = append(c.moduleOrder, path)

	return nil
}
