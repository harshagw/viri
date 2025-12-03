package internal

import (
	"time"
)

type Clock struct {}

func NewClock() *Clock {
	return &Clock{}
}

func (c *Clock) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	return time.Now().Unix(), nil
}

func (c *Clock) Arity() int {
	return 0
}

func (c *Clock) ToString() string {
	return "<native_fun clock>"
}