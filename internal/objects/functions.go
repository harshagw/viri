package objects

import "github.com/harshagw/viri/internal/ast"

type CallableFunction struct {
	declaration   *ast.FunctionStmt
	closure       *Environment
	isInitializer bool
}

func NewCallableFunction(declaration *ast.FunctionStmt, closure *Environment, isInitializer bool) *CallableFunction {
	return &CallableFunction{declaration: declaration, closure: closure, isInitializer: isInitializer}
}

func (cf *CallableFunction) Call(exec BlockExecutor, arguments []interface{}) (interface{}, error) {
	environment := NewEnvironment(cf.closure)
	for idx, parameter := range cf.declaration.Params {
		environment.Define(parameter.Lexeme, arguments[idx])
	}
	result, err := exec.ExecuteBlock(cf.declaration.Body, environment)
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

func (cf *CallableFunction) Arity() int {
	return len(cf.declaration.Params)
}

func (cf *CallableFunction) String() string {
	return "<fun " + cf.declaration.Name.Lexeme + ">"
}

func (cf *CallableFunction) Bind(instance *ClassInstance) *CallableFunction {
	environment := NewEnvironment(cf.closure)
	environment.Define("this", instance)
	return NewCallableFunction(cf.declaration, environment, cf.isInitializer)
}
