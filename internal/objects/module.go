package objects

import (
	"sync"

	"github.com/harshagw/viri/internal/ast"
)

// Module represents a parsed and executed module.
type Module struct {
	Path       string            // absolute, normalized path
	Imports    []*ast.ImportStmt // import statements
	Exports    map[string]Object // exported symbols
	Namespace  *Namespace        // namespace object
	Statements []ast.Stmt        // non-import statements
}

// NewModule creates a new module.
func NewModule(path string, imports []*ast.ImportStmt, statements []ast.Stmt) *Module {
	return &Module{
		Path:       path,
		Imports:    imports,
		Exports:    make(map[string]Object),
		Statements: statements,
	}
}

// GetAllStatements returns all statements in execution order (imports first, then other statements).
func (m *Module) GetAllStatements() []ast.Stmt {
	all := make([]ast.Stmt, 0, len(m.Imports)+len(m.Statements))
	for _, importStmt := range m.Imports {
		all = append(all, importStmt)
	}
	all = append(all, m.Statements...)
	return all
}

// ModuleCache caches loaded modules.
type ModuleCache struct {
	modules map[string]*Module
	mu      sync.Mutex
}

// NewModuleCache creates a new module cache.
func NewModuleCache() *ModuleCache {
	return &ModuleCache{
		modules: make(map[string]*Module),
	}
}

// Get retrieves a cached module.
func (c *ModuleCache) Get(path string) (*Module, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	mod, ok := c.modules[path]
	return mod, ok
}

// Put caches a module.
func (c *ModuleCache) Put(path string, mod *Module) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.modules[path] = mod
}

