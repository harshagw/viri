package objects

import (
	"fmt"

	"github.com/harshagw/viri/internal/code"
)

// CompiledFunction holds the bytecode for a compiled function.
type CompiledFunction struct {
	Instructions  code.Instructions
	NumLocals     int
	NumParameters int
}

func (cf *CompiledFunction) Type() Type {
	return TypeCompiledFunction
}

func (cf *CompiledFunction) Inspect() string {
	return fmt.Sprintf("CompiledFunction[%p]", cf)
}
