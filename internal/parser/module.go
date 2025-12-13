package parser

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/scanner"
	"github.com/harshagw/viri/internal/token"
)

// ResolveModulePath resolves a module path relative to a base directory.
func ResolveModulePath(baseDir, importPath string) (string, error) {
	// Join the base directory with the import path
	fullPath := filepath.Join(baseDir, importPath)

	// Convert to absolute path
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path '%s': %w", importPath, err)
	}

	// Clean the path (normalize)
	absPath = filepath.Clean(absPath)

	return absPath, nil
}

func LoadModuleFile(path string, diagnosticHandler objects.DiagnosticHandler) (*ast.Module, error) {
	code, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read module '%s': %w", path, err)
	}
	
	var filePathPtr *string
	if path != "" {
		filePathPtr = &path
	}

	sc := scanner.New(bytes.NewBuffer(code), filePathPtr)
	tokens, err := sc.Scan()
	if err != nil {
		return nil, fmt.Errorf("failed to scan module '%s': %w", path, err)
	}

	p := NewParser(tokens, diagnosticHandler)
	p.SetFilePath(path)
	mod, err := p.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse module '%s': %w", path, err)
	}
	
	return mod, nil
}

func ParseModule(tokens []token.Token, path string, diagnosticHandler objects.DiagnosticHandler) (*ast.Module, error) {
	p := NewParser(tokens, diagnosticHandler)
	p.SetFilePath(path)
	return p.Parse()
}
