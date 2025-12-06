package objects

import (
	"errors"
	"time"
)

type Clock struct{}

func NewClock() *Clock {
	return &Clock{}
}

func (c *Clock) Call(exec BlockExecutor, arguments []Object) (Object, error) {
	return NewNumber(float64(time.Now().Unix())), nil
}

func (c *Clock) Arity() int {
	return 0
}

func (c *Clock) String() string {
	return "<native_fun clock>"
}

func (c *Clock) Inspect() string { return c.String() }
func (c *Clock) Type() Type      { return TypeNativeFun }

type Len struct{}

func NewLen() *Len {
	return &Len{}
}

func (l *Len) Call(exec BlockExecutor, arguments []Object) (Object, error) {
	value := arguments[0]
	switch value.Type() {
	case TypeString:
		return NewNumber(float64(len(value.Inspect()))), nil
	case TypeArray:
		if arr, ok := value.(*Array); ok {
			return NewNumber(float64(len(arr.Elements))), nil
		}
	}
	return nil, errors.New("invalid argument type for len function: " + string(value.Type()))
}

func (l *Len) Arity() int {
	return 1
}

func (l *Len) String() string {
	return "<native_fun len>"
}

func (l *Len) Inspect() string { return l.String() }
func (l *Len) Type() Type      { return TypeNativeFun }