package internal

import (
	"fmt"
)

type Environment struct {
	enclosing *Environment
	values map[string]interface{}
}

func NewEnvironment(enclosing *Environment) *Environment {
	return &Environment{
		enclosing: enclosing,
		values: make(map[string]interface{}),
	}
}

func (e *Environment) define(name string, value interface{}) {
	e.values[name] = value
}

func (e *Environment) assign(name string, value interface{}) error {
	if _, ok := e.values[name]; !ok {
		if e.enclosing != nil {
			return e.enclosing.assign(name, value)
		}
		return fmt.Errorf("%s: variable not found", name)
	}
	e.values[name] = value
	return nil
}

func (e *Environment) get(name string) (interface{}, error) {
	value, ok := e.values[name]
	if !ok {
		if e.enclosing != nil {
			return e.enclosing.get(name)
		}
		return nil, fmt.Errorf("%s: variable not found", name)
	}
	return value, nil
}