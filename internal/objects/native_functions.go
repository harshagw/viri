package objects

import "time"

type Clock struct{}

func NewClock() *Clock {
	return &Clock{}
}

func (c *Clock) Call(exec BlockExecutor, arguments []interface{}) (interface{}, error) {
	return float64(time.Now().Unix()), nil
}

func (c *Clock) Arity() int {
	return 0
}

func (c *Clock) String() string {
	return "<native_fun clock>"
}
