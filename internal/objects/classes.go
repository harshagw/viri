package objects

import (
	"errors"

	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/token"
)

type CallableClass struct {
	name    string
	methods map[string]*CallableFunction
}

func NewCallableClass(name string, methods map[string]*CallableFunction) *CallableClass {
	return &CallableClass{name: name, methods: methods}
}

func (cc *CallableClass) String() string {
	return "<class " + cc.name + ">"
}

func (cc *CallableClass) Arity() int {
	initializer := cc.methods["init"]
	if initializer != nil {
		return initializer.Arity()
	}
	return 0
}

func (cc *CallableClass) Call(exec BlockExecutor, arguments []interface{}) (interface{}, error) {
	newInstance := NewClassInstance(cc)
	initializer := cc.methods["init"]
	if initializer != nil {
		_, _ = initializer.Bind(newInstance).Call(exec, arguments)
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

func (ci *ClassInstance) String() string {
	return "<instance " + ci.class.name + ">"
}

func (ci *ClassInstance) Get(name token.Token) (interface{}, error) {
	if value, ok := ci.fields[name.Lexeme]; ok {
		return value, nil
	}
	if method, ok := ci.class.methods[name.Lexeme]; ok {
		return method.Bind(ci), nil
	}
	return nil, errors.New("instance does not have field " + name.Lexeme)
}

func (ci *ClassInstance) Set(name token.Token, value interface{}) error {
	ci.fields[name.Lexeme] = value
	return nil
}

// BindMethod resolves a method name to a bound callable.
func (ci *ClassInstance) BindMethod(method *CallableFunction) *CallableFunction {
	return method.Bind(ci)
}

// LookupMethod finds a method by name on the class (without binding).
func (cc *CallableClass) LookupMethod(name string) (*CallableFunction, bool) {
	m, ok := cc.methods[name]
	return m, ok
}

// BuildMethods constructs callable functions from AST method declarations.
func BuildMethods(methods []*ast.FunctionStmt, closure *Environment) map[string]*CallableFunction {
	result := make(map[string]*CallableFunction)
	for _, method := range methods {
		function := NewCallableFunction(method, closure, method.Name.Lexeme == "init")
		result[method.Name.Lexeme] = function
	}
	return result
}
