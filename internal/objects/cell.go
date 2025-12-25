package objects

import "fmt"

// Cell is a mutable container for a value that can modify captured variables.
type Cell struct {
	Value Object
}

func NewCell(value Object) *Cell {
	return &Cell{Value: value}
}

func (c *Cell) Type() Type {
	return TypeCell
}

func (c *Cell) Inspect() string {
	return fmt.Sprintf("cell(%s)", c.Value.Inspect())
}
