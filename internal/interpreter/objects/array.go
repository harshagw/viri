package objects

import (
	"fmt"
	"strings"
)

type Array struct {
	Elements []Object
}

func NewArray(elements []Object) *Array {
	return &Array{Elements: elements}
}

func (a *Array) Type() Type {
	return TypeArray
}

func (a *Array) Inspect() string {
	var out strings.Builder

	elements := []string{}
	for _, e := range a.Elements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}

// Get returns the element at the given index with bounds checking.
func (a *Array) Get(index int) (Object, error) {
	if index < 0 || index >= len(a.Elements) {
		return nil, fmt.Errorf("index out of bounds")
	}
	return a.Elements[index], nil
}

// Set writes the value at the given index with bounds checking.
func (a *Array) Set(index int, value Object) error {
	if index < 0 || index >= len(a.Elements) {
		return fmt.Errorf("index out of bounds")
	}
	a.Elements[index] = value
	return nil
}
