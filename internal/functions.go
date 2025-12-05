package internal

type Callable interface {
	Call(i *Interpreter, arguments []interface{}) (interface{}, error)
	Arity() int
	ToString() string
}

type CallableFunction struct {
	declaration   *Function
	closure       *Environment
	isInitializer bool
}

func NewCallableFunction(declaration *Function, closure *Environment, isInitializer bool) *CallableFunction {
	return &CallableFunction{declaration: declaration, closure: closure, isInitializer: isInitializer}
}

func (cf *CallableFunction) Call(i *Interpreter, arguments []interface{}) (interface{}, error) {
	environment := NewEnvironment(cf.closure)
	for i, parameter := range cf.declaration.parameters {
		environment.define(parameter.Lexeme, arguments[i])
	}
	result, err := i.executeBlock(cf.declaration.body, environment)
	if err != nil {
		if _, ok := err.(*ReturnError); ok {
			return err.(*ReturnError).value, nil
		}
		return nil, err
	}
	if cf.isInitializer {
		return cf.closure.getAt(0, "this")
	}
	return result, nil
}

func (cf *CallableFunction) Arity() int {
	return len(cf.declaration.parameters)
}

func (cf *CallableFunction) ToString() string {
	return "<fun " + cf.declaration.token.Lexeme + ">"
}

func (cf *CallableFunction) Bind(instance *ClassInstance) *CallableFunction {
	environment := NewEnvironment(cf.closure)
	environment.define("this", instance)
	return NewCallableFunction(cf.declaration, environment, cf.isInitializer)
}
