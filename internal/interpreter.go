package internal

import (
	"errors"
	"fmt"
	"strconv"
)

// BreakError is a special error type used for control flow
type BreakError struct{}

func (e *BreakError) Error() string {
	return "break"
}

// ReturnError is a special error type used for control flow
type ReturnError struct {
	value interface{}
}

func (e *ReturnError) Error() string {
	return "return " + fmt.Sprintf("%v", e.value)
}

type Interpreter struct {
	viri        *Viri
	environment *Environment
	globals     *Environment
	localMap    map[Expr]int
}

func NewInterpreter(viri *Viri) *Interpreter {
	globals := NewEnvironment(nil)
	globals.define("clock", NewClock())
	return &Interpreter{
		viri:        viri,
		environment: globals,
		globals:     globals,
		localMap:    make(map[Expr]int),
	}
}

func (i *Interpreter) Interpret(stmts []Stmt) error {
	for _, stmt := range stmts {
		_, err := stmt.Accept(i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Interpreter) resolve(expr Expr, depth int) {
	i.localMap[expr] = depth
}

func (i *Interpreter) visitBlock(block *Block) (interface{}, error) {
	newEnvironment := NewEnvironment(i.environment)
	return i.executeBlock(block, newEnvironment)
}

func (i *Interpreter) executeBlock(block *Block, environment *Environment) (interface{}, error) {
	previousEnvironment := i.environment
	i.environment = environment
	defer func() {
		i.environment = previousEnvironment
	}()
	for _, stmt := range block.statements {
		_, err := i.evaluateStmt(stmt)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (i *Interpreter) visitFunction(function *Function) (interface{}, error) {
	callableFunction := NewCallableFunction(function, i.environment, false)
	i.environment.define(function.token.Lexeme, callableFunction)
	return nil, nil
}

func (i *Interpreter) visitClass(class *Class) (interface{}, error) {
	i.environment.define(class.name.Lexeme, nil)

	methods := make(map[string]*CallableFunction)

	for _, method := range class.methods {
		function := NewCallableFunction(method, i.environment, method.token.Lexeme == "init")
		i.environment.define(method.token.Lexeme, function)
		methods[method.token.Lexeme] = function
	}

	callableClass := NewCallableClass(class.name.Lexeme, methods)

	i.environment.assign(class.name.Lexeme, callableClass)

	return nil, nil
}

func (i *Interpreter) visitReturnStmt(returnStmt *ReturnStmt) (interface{}, error) {
	var value interface{}

	if returnStmt.value != nil {
		var err error
		value, err = i.evaluateExpr(returnStmt.value)
		if err != nil {
			return nil, err
		}
	}

	return value, &ReturnError{value: value}
}

func (i *Interpreter) visitVarDeclStmt(varDeclStmt *VarDeclStmt) (interface{}, error) {
	if varDeclStmt.initializer != nil {
		value, err := i.evaluateExpr(varDeclStmt.initializer)
		if err != nil {
			return nil, err
		}
		i.environment.define(varDeclStmt.token.Lexeme, value)
	} else {
		i.environment.define(varDeclStmt.token.Lexeme, nil)
	}
	return nil, nil
}

func (i *Interpreter) visitPrintStmt(printStmt *PrintStmt) (interface{}, error) {
	value, err := i.evaluateExpr(printStmt.Expr)
	if err != nil {
		return nil, err
	}
	err = printStmt.Print(value)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (i *Interpreter) visitExprStmt(exprStmt *ExprStmt) (interface{}, error) {
	value, err := i.evaluateExpr(exprStmt.Expr)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (i *Interpreter) visitIfStmt(ifStmt *IfStmt) (interface{}, error) {
	value, err := i.evaluateExpr(ifStmt.condition)
	if err != nil {
		return nil, err
	}
	if i.isTruthy(value) {
		return i.evaluateStmt(ifStmt.ifBranch)
	} else if ifStmt.elseBranch != nil {
		return i.evaluateStmt(ifStmt.elseBranch)
	}
	return nil, nil
}

func (i *Interpreter) visitWhileStmt(whileStmt *WhileStmt) (interface{}, error) {
	for {
		condition, err := i.evaluateExpr(whileStmt.condition)
		if err != nil {
			return nil, err
		}
		if !i.isTruthy(condition) {
			break
		}
		_, err = i.evaluateStmt(whileStmt.body)
		if err != nil {
			// Check if it's a break error
			if _, ok := err.(*BreakError); ok {
				// Break out of the loop
				break
			}
			return nil, err
		}
	}
	return nil, nil
}

func (i *Interpreter) evaluateStmt(stmt Stmt) (interface{}, error) {
	result, err := stmt.Accept(i)
	// Automatically propagate BreakError
	if _, ok := err.(*BreakError); ok {
		return nil, err
	}
	return result, err
}

func (i *Interpreter) evaluateExpr(expr Expr) (interface{}, error) {
	return expr.Accept(i)
}

func (i *Interpreter) visitBinaryExp(binaryExp *BinaryExp) (interface{}, error) {
	rightExp, err := i.evaluateExpr(binaryExp.Right)
	if err != nil {
		return nil, err
	}
	leftExp, err := i.evaluateExpr(binaryExp.Left)
	if err != nil {
		return nil, err
	}
	switch binaryExp.Operator.TokenType {
	case PLUS:
		leftNum, leftIsNum := leftExp.(float64)
		rightNum, rightIsNum := rightExp.(float64)
		leftStr, leftIsStr := leftExp.(string)
		rightStr, rightIsStr := rightExp.(string)

		if leftIsNum && rightIsNum {
			return leftNum + rightNum, nil
		} else if leftIsStr && rightIsStr {
			return leftStr + rightStr, nil
		} else {
			return nil, fmt.Errorf("operands to '+' must both be numbers or both be strings (left: %T=%v, right: %T=%v)", leftExp, leftExp, rightExp, rightExp)
		}
	case MINUS:
		if !i.isNumber(leftExp) || !i.isNumber(rightExp) {
			i.viri.Error(binaryExp.Operator, "Operand must be a number")
			return nil, errors.New("operand must be a number")
		}
		return leftExp.(float64) - rightExp.(float64), nil
	case STAR:
		if !i.isNumber(leftExp) || !i.isNumber(rightExp) {
			i.viri.Error(binaryExp.Operator, "Operand must be a number")
			return nil, errors.New("operand must be a number")
		}
		return leftExp.(float64) * rightExp.(float64), nil
	case SLASH:
		if !i.isNumber(leftExp) || !i.isNumber(rightExp) {
			i.viri.Error(binaryExp.Operator, "Operand must be a number")
			return nil, errors.New("operand must be a number")
		}
		if rightExp.(float64) == 0 {
			i.viri.Error(binaryExp.Operator, "Division by zero")
			return nil, errors.New("division by zero")
		}
		return leftExp.(float64) / rightExp.(float64), nil
	case GREATER:
		if !i.isNumber(leftExp) || !i.isNumber(rightExp) {
			i.viri.Error(binaryExp.Operator, "Operand must be a number")
			return nil, errors.New("operand must be a number")
		}
		return leftExp.(float64) > rightExp.(float64), nil
	case GREATER_EQUAL:
		if !i.isNumber(leftExp) || !i.isNumber(rightExp) {
			i.viri.Error(binaryExp.Operator, "Operand must be a number")
			return nil, errors.New("operand must be a number")
		}
		return leftExp.(float64) >= rightExp.(float64), nil
	case LESS:
		if !i.isNumber(leftExp) || !i.isNumber(rightExp) {
			i.viri.Error(binaryExp.Operator, "Operand must be a number")
			return nil, errors.New("operand must be a number")
		}
		return leftExp.(float64) < rightExp.(float64), nil
	case LESS_EQUAL:
		if !i.isNumber(leftExp) || !i.isNumber(rightExp) {
			i.viri.Error(binaryExp.Operator, "Operand must be a number")
			return nil, errors.New("operand must be a number")
		}
		return leftExp.(float64) <= rightExp.(float64), nil
	case EQUAL_EQUAL:
		return i.isEqual(leftExp, rightExp), nil
	case BANG_EQUAL:
		return !i.isEqual(leftExp, rightExp), nil
	}
	return nil, errors.New("invalid operator")
}

func (i *Interpreter) visitGrouping(grouping *Grouping) (interface{}, error) {
	return i.evaluateExpr(grouping.Expr)
}

func (i *Interpreter) visitLiteral(literal *Literal) (interface{}, error) {
	return literal.Value, nil
}

func (i *Interpreter) visitUnary(unary *Unary) (interface{}, error) {
	rightExpr, err := i.evaluateExpr(unary.Expr)
	if err != nil {
		return nil, err
	}
	switch unary.Operator.TokenType {
	case MINUS:
		return -rightExpr.(float64), nil
	case BANG:
		return !i.isTruthy(rightExpr), nil
	}
	return nil, errors.New("invalid operator")
}

func (i *Interpreter) visitVariable(variable *Variable) (interface{}, error) {
	return i.findVariable(variable, variable.Name)
}

func (i *Interpreter) visitThisExpr(this *This) (interface{}, error) {
	return i.findVariable(this, this.keyword)
}

func (i *Interpreter) findVariable(expr Expr, token Token) (interface{}, error) {
	if dist, ok := i.localMap[expr]; ok {
		return i.environment.getAt(dist, token.Lexeme)
	}
	return i.globals.get(token.Lexeme)
}

func (i *Interpreter) visitAssignment(assignment *Assignment) (interface{}, error) {
	value, err := i.evaluateExpr(assignment.Value)
	if err != nil {
		return nil, err
	}
	if dist, ok := i.localMap[assignment]; ok {
		err = i.environment.assignAt(dist, assignment.Name.Lexeme, value)
		if err != nil {
			return nil, err
		}
	} else {
		err = i.globals.assign(assignment.Name.Lexeme, value)
		if err != nil {
			return nil, err
		}
	}
	return value, nil
}

func (i *Interpreter) visitLogical(logical *Logical) (interface{}, error) {
	left, err := i.evaluateExpr(logical.Left)
	if err != nil {
		return nil, err
	}
	if (logical.Operator.TokenType == OR && i.isTruthy(left)) || (logical.Operator.TokenType == AND && !i.isTruthy(left)) {
		// short circuit
		return left, nil
	}

	return i.evaluateExpr(logical.Right)
}

func (i *Interpreter) visitCall(call *Call) (interface{}, error) {
	callee, err := i.evaluateExpr(call.callee)
	if err != nil {
		return nil, err
	}

	arguments := []interface{}{}
	for _, argument := range call.arguments {
		argument, err := i.evaluateExpr(argument)
		if err != nil {
			return nil, err
		}
		arguments = append(arguments, argument)
	}

	callable, ok := callee.(Callable)

	if !ok {
		return nil, errors.New("can only call which is a function")
	}

	if callable.Arity() != len(arguments) {
		return nil, errors.New("expected " + strconv.Itoa(callable.Arity()) + " arguments but got " + strconv.Itoa(len(arguments)))
	}

	return callable.Call(i, arguments)
}

func (i *Interpreter) visitGet(get *Get) (interface{}, error) {
	object, err := i.evaluateExpr(get.object)
	if err != nil {
		return nil, err
	}
	instance, ok := object.(*ClassInstance)
	if !ok {
		return nil, errors.New("object is not a class instance")
	}

	return instance.Get(get.name)
}

func (i *Interpreter) visitSet(set *Set) (interface{}, error) {
	object, err := i.evaluateExpr(set.object)
	if err != nil {
		return nil, err
	}
	instance, ok := object.(*ClassInstance)
	if !ok {
		return nil, errors.New("object is not a class instance")
	}
	value, err := i.evaluateExpr(set.value)
	if err != nil {
		return nil, err
	}
	err = instance.Set(set.name, value)
	if err != nil {
		return nil, err
	}
	return value, nil
}

// Utility functions

func (i *Interpreter) isTruthy(value interface{}) bool {
	if value == nil {
		return false
	}
	if value == false {
		return false
	}
	return true
}

func (i *Interpreter) isNumber(value interface{}) bool {
	_, ok := value.(float64)
	return ok
}

func (i *Interpreter) isEqual(a interface{}, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil {
		return false
	}
	return a == b
}

func (i *Interpreter) visitBreakStmt(breakStmt *BreakStmt) (interface{}, error) {
	return nil, &BreakError{}
}
