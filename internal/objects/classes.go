package objects

import (
	"errors"

	"github.com/harshagw/viri/internal/token"
)

// Class represents a user-defined class.
type Class struct {
	name       string
	superClass *Class
	methods    map[string]*Function
}

func NewClass(name string, superClass *Class, methods map[string]*Function) *Class {
	return &Class{name: name, superClass: superClass, methods: methods}
}

func (cc *Class) String() string {
	return "<class " + cc.name + ">"
}

func (cc *Class) Inspect() string { return cc.String() }
func (cc *Class) Type() Type      { return TypeClass }

func (cc *Class) Arity() int {
	initializer := cc.methods["init"]
	if initializer != nil {
		return initializer.Arity()
	}
	return 0
}

func (cc *Class) Call(exec BlockExecutor, arguments []Object) (Object, error) {
	newInstance := NewClassInstance(cc)
	initializer := cc.methods["init"]
	if initializer != nil {
		if _, err := initializer.Bind(newInstance).Call(exec, arguments); err != nil {
			return nil, err
		}
	}
	return newInstance, nil
}

type ClassInstance struct {
	class  *Class
	fields map[string]Object
}

func NewClassInstance(class *Class) *ClassInstance {
	return &ClassInstance{class: class, fields: make(map[string]Object)}
}

func (ci *ClassInstance) String() string {
	return "<instance " + ci.class.name + ">"
}

func (ci *ClassInstance) Inspect() string { return ci.String() }
func (ci *ClassInstance) Type() Type      { return TypeInstance }

func (ci *ClassInstance) Get(name *token.Token) (Object, error) {
	if value, ok := ci.fields[name.Lexeme]; ok {
		return value, nil
	}
	if method, ok := ci.class.LookupMethod(name.Lexeme); ok {
		return method.Bind(ci), nil
	}
	return nil, errors.New("instance does not have field " + name.Lexeme)
}

func (ci *ClassInstance) Set(name *token.Token, value Object) error {
	ci.fields[name.Lexeme] = value
	return nil
}

// LookupMethod finds a method by name on the class (without binding).
func (cc *Class) LookupMethod(name string) (*Function, bool) {
	m, ok := cc.methods[name]
	if !ok && cc.superClass != nil {
		return cc.superClass.LookupMethod(name)
	}
	return m, ok
}
