package objects

import "github.com/harshagw/viri/internal/ast"

// BlockExecutor executes a block with a provided environment.
type BlockExecutor interface {
	ExecuteBlock(block *ast.BlockStmt, env *Environment) (interface{}, error)
}

// Callable represents any callable value (function, class, native).
type Callable interface {
	Call(exec BlockExecutor, arguments []interface{}) (interface{}, error)
	Arity() int
	String() string
}
