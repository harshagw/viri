package objects

import "fmt"

// Closure is a wrapper around a compiled function
// It has a list of free cell (containers around variables) used by the function
type Closure struct {
	Fn   *CompiledFunction
	Free []*Cell
}

func NewClosure(fn *CompiledFunction, free []*Cell) *Closure {
	return &Closure{Fn: fn, Free: free}
}

func (c *Closure) Type() Type {
	return TypeClosure
}

func (c *Closure) Inspect() string {
	return fmt.Sprintf("closure(%s)", c.Fn.Inspect())
}
