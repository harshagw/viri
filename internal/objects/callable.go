package objects

import "github.com/harshagw/viri/internal/ast"

// BlockExecutor executes a block with a provided environment.
type BlockExecutor interface {
	ExecuteBlock(block *ast.BlockStmt, env *Environment) (Object, error)
}

// Callable represents any callable value (function, class, native).
type Callable interface {
	Call(exec BlockExecutor, arguments []Object) (Object, error)
	Arity() int
	String() string
}
