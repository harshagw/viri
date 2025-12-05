package objects

import "fmt"

type Environment struct {
	enclosing *Environment
	values    map[string]Object
}

func NewEnvironment(enclosing *Environment) *Environment {
	return &Environment{
		enclosing: enclosing,
		values:    make(map[string]Object),
	}
}

func (e *Environment) Define(name string, value Object) {
	e.values[name] = value
}

func (e *Environment) Assign(name string, value Object) error {
	if _, ok := e.values[name]; !ok {
		if e.enclosing != nil {
			return e.enclosing.Assign(name, value)
		}
		return fmt.Errorf("'%s' variable not found", name)
	}
	e.values[name] = value
	return nil
}

func (e *Environment) Get(name string) (Object, error) {
	value, ok := e.values[name]
	if !ok {
		if e.enclosing != nil {
			return e.enclosing.Get(name)
		}
		return nil, fmt.Errorf("'%s' variable not found", name)
	}
	return value, nil
}

func (e *Environment) GetAt(distance int, name string) (Object, error) {
	env := e.ancestor(distance)
	value, ok := env.values[name]
	if !ok {
		return nil, fmt.Errorf("'%s' variable not found at distance %d", name, distance)
	}
	return value, nil
}

func (e *Environment) AssignAt(distance int, name string, value Object) error {
	e.ancestor(distance).values[name] = value
	return nil
}

func (e *Environment) ancestor(distance int) *Environment {
	environment := e
	for i := 0; i < distance; i++ {
		environment = environment.enclosing
	}
	return environment
}
