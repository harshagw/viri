package objects

import "time"

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
