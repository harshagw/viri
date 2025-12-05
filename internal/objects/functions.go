package objects

import "github.com/harshagw/viri/internal/ast"

// Function represents a user-defined function value.
type Function struct {
	declaration   *ast.FunctionStmt
	closure       *Environment
	isInitializer bool
}

func NewFunction(declaration *ast.FunctionStmt, closure *Environment, isInitializer bool) *Function {
	return &Function{declaration: declaration, closure: closure, isInitializer: isInitializer}
}

func (cf *Function) Call(exec BlockExecutor, arguments []Object) (Object, error) {
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

func (cf *Function) Arity() int {
	return len(cf.declaration.Params)
}

func (cf *Function) String() string {
	return "<fun " + cf.declaration.Name.Lexeme + ">"
}

func (cf *Function) Bind(instance *ClassInstance) *Function {
	environment := NewEnvironment(cf.closure)
	environment.Define("this", instance)
	return NewFunction(cf.declaration, environment, cf.isInitializer)
}

func (cf *Function) Type() Type {
	return TypeFunction
}

func (cf *Function) Inspect() string {
	return cf.String()
}
