package objects

import "fmt"

// Closure is a wrapper around a compiled function
// It has a list of free cell (containers around variables) used by the function
type Closure struct {
	Fn   *CompiledFunction
	Free []*Cell

	// DefiningClass is the class this closure was declared in (set when the
	// class is built, and inherited by closures created inside a method).
	DefiningClass *CompiledClass
}

func NewClosure(fn *CompiledFunction, free []*Cell) *Closure {
	return &Closure{Fn: fn, Free: free}
}

func (c *Closure) Type() Type {
	return TypeClosure
}

func (c *Closure) Inspect() string {
	if c.Fn.Name == "" {
		return "<fun anonymous>"
	}
	return fmt.Sprintf("<fun %s>", c.Fn.Name)
}
