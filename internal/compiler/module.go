package compiler

import (
	"fmt"
	"path/filepath"
	"slices"

	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/parser"
)

// CompileProgram compiles a program starting from the entry module
func (c *Compiler) CompileProgram(entryPath string) (*objects.CompiledProgram, error) {
	// Load all modules and build dependency graph
	if err := c.loadModule(entryPath, []string{}); err != nil {
		return nil, err
	}

	// Topologically sort modules (dependencies first)
	order, err := c.topologicalSort()
	if err != nil {
		return nil, err
	}
	c.moduleOrder = order

	// Assign module indices
	for i, path := range c.moduleOrder {
		c.moduleIndices[path] = i
	}

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
	}, nil
}

// compileModule compiles a single module using the shared constants table
func (c *Compiler) compileModule(path string) (objects.CompiledModule, error) {
	mod := c.modules[path]

	// Reset compiler state for this module
	c.reset(nil)

	// Register imports - we need to know what each imported module exports
	for _, importStmt := range mod.Imports {
		importPath, ok := importStmt.Path.Literal.(string)
		if !ok {
			return objects.CompiledModule{}, fmt.Errorf("import path must be a string")
		}

		targetPath, err := parser.ResolveModulePath(filepath.Dir(path), importPath)
		if err != nil {
			return objects.CompiledModule{}, err
		}

		moduleIdx := c.moduleIndices[targetPath]
		exportMap := c.buildExportMap(c.modules[targetPath])
		c.symbolTable.DefineImport(importStmt.Alias.Lexeme, moduleIdx, exportMap)
	}

	// Track exports as we compile
	var exportNames []string
	exportSlots := make(map[string]int)

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

		// Record global slot before compiling (for exports)
		globalSlot := c.symbolTable.NumDefinitions()

		if err := c.Compile(stmt); err != nil {
			return objects.CompiledModule{}, err
		}

		if exported {
			exportNames = append(exportNames, exportName)
			exportSlots[exportName] = globalSlot
		}
	}

	// Build exports array (export index -> global slot)
	exports := make([]int, len(exportNames))
	for i, name := range exportNames {
		exports[i] = exportSlots[name]
	}

	return objects.CompiledModule{
		Instructions: c.currentInstructions(),
		NumGlobals:   c.maxGlobalIndex + 1,
		Exports:      exports,
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

		targetPath, err := parser.ResolveModulePath(filepath.Dir(path), importPath)
		if err != nil {
			return err
		}

		if err := c.loadModule(targetPath, newStack); err != nil {
			return err
		}
	}

	return nil
}

// topologicalSort returns modules in dependency order (dependencies first)
func (c *Compiler) topologicalSort() ([]string, error) {
	adjList := make(map[string][]string)
	inDegree := make(map[string]int)

	for path := range c.modules {
		inDegree[path] = 0
		adjList[path] = []string{}
	}

	// Build edges: if A imports B, then B -> A (B must come before A)
	for path, mod := range c.modules {
		for _, importStmt := range mod.Imports {
			importPath, ok := importStmt.Path.Literal.(string)
			if !ok {
				continue
			}
			targetPath, err := parser.ResolveModulePath(filepath.Dir(path), importPath)
			if err != nil {
				return nil, err
			}
			adjList[targetPath] = append(adjList[targetPath], path)
			inDegree[path]++
		}
	}

	// Kahn's algorithm
	var queue []string
	for path, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, path)
		}
	}

	var result []string
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		for _, neighbor := range adjList[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(result) != len(c.modules) {
		return nil, fmt.Errorf("cycle detected in module dependencies")
	}

	return result, nil
}
