package objects

import (
	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/token"
)

// enum of function types
type FunctionType int

const (
	FunctionTypeAnonymous FunctionType = iota
	FunctionTypeNamed
)

func (ft FunctionType) String() string {
	switch ft {
	case FunctionTypeAnonymous:
		return "anonymous"
	case FunctionTypeNamed:
		return "named"
	}
	return "unknown function type"
}

// Function represents a user-defined function value.
type Function struct {
	name          string
	params        []*token.Token
	body          *ast.BlockStmt
	functionType  FunctionType
	closure       *Environment
	isInitializer bool
}

func NewFunction(name string, params []*token.Token, body *ast.BlockStmt, closure *Environment, isInitializer bool, functionType FunctionType) *Function {
	return &Function{
		name:          name,
		params:        params,
		body:          body,
		functionType:  functionType,
		closure:       closure,
		isInitializer: isInitializer,
	}
}

func (cf *Function) Call(exec BlockExecutor, arguments []Object) (Object, error) {
	environment := NewEnvironment(cf.closure)
	for idx, parameter := range cf.params {
		environment.Define(parameter.Lexeme, arguments[idx])
	}

	result, err := exec.ExecuteBlock(cf.body, environment)
	if err != nil {
		if ret, ok := err.(*ReturnError); ok {
			if cf.isInitializer {
				return cf.closure.GetAt(0, "this")
			}
			return ret.Value, nil
		}
		return nil, err
	}
	if cf.isInitializer {
		return cf.closure.GetAt(0, "this")
	}
	return result, nil
}

func (cf *Function) Arity() int {
	return len(cf.params)
}

func (cf *Function) String() string {
	if cf.functionType == FunctionTypeAnonymous {
		return "<fun anonymous>"
	}
	return "<fun " + cf.name + ">"
}

func (cf *Function) Bind(instance *ClassInstance) *Function {
	environment := NewEnvironment(cf.closure)
	environment.Define("this", instance)
	return NewFunction(cf.name, cf.params, cf.body, environment, cf.isInitializer, cf.functionType)
}

func (cf *Function) Type() Type {
	return TypeFunction
}

func (cf *Function) Inspect() string {
	return cf.String()
}
