package internal

import "errors"

type CallableClass struct {
	name    string
	methods map[string]*CallableFunction
}

func NewCallableClass(name string, methods map[string]*CallableFunction) *CallableClass {
	return &CallableClass{name: name, methods: methods}
}

func (cc *CallableClass) ToString() string {
	return "<class " + cc.name + ">"
}

func (cc *CallableClass) Arity() int {
	initializer := cc.methods["init"]
	if initializer != nil {
		return initializer.Arity()
	}
	return 0
}

func (cc *CallableClass) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	newInstance := NewClassInstance(cc)
	initializer := cc.methods["init"]
	if initializer != nil {
		initializer.Bind(newInstance).Call(i, arguments)
	}
	return newInstance, nil
}

type ClassInstance struct {
	class  *CallableClass
	fields map[string]interface{}
}

func NewClassInstance(class *CallableClass) *ClassInstance {
	return &ClassInstance{class: class, fields: make(map[string]interface{})}
}

func (ci *ClassInstance) ToString() string {
	return "<instance " + ci.class.name + ">"
}

func (ci *ClassInstance) Get(name Token) (interface{}, error) {
	if value, ok := ci.fields[name.Lexeme]; ok {
		return value, nil
	}
	if method, ok := ci.class.methods[name.Lexeme]; ok {
		return method.Bind(ci), nil
	}
	return nil, errors.New("instance does not have field " + name.Lexeme)
}

func (ci *ClassInstance) Set(name Token, value interface{}) error {
	ci.fields[name.Lexeme] = value
	return nil
}
