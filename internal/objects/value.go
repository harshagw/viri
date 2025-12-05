package objects

import (
	"fmt"
	"math"
	"strconv"
)

// Number represents a numeric value.
type Number struct {
	Value float64
}

func NewNumber(v float64) *Number { return &Number{Value: v} }
func (n *Number) Type() Type      { return TypeNumber }
func (n *Number) Inspect() string {
	if n.Value == math.Trunc(n.Value) {
		return fmt.Sprintf("%.0f", n.Value)
	}
	return fmt.Sprintf("%g", n.Value)
}

// String represents a string value.
type String struct {
	Value string
}

func NewString(v string) *String { return &String{Value: v} }
func (s *String) Type() Type     { return TypeString }
func (s *String) Inspect() string {
	return s.Value
}

// Bool represents a boolean value.
type Bool struct {
	Value bool
}

func NewBool(v bool) *Bool { return &Bool{Value: v} }
func (b *Bool) Type() Type { return TypeBool }
func (b *Bool) Inspect() string {
	return strconv.FormatBool(b.Value)
}

// Nil is the singleton nil value.
type Nil struct{}

func NewNil() *Nil        { return &Nil{} }
func (n *Nil) Type() Type { return TypeNil }
func (n *Nil) Inspect() string {
	return "nil"
}
