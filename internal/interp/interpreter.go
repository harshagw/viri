package interp

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/token"
)

type Interpreter struct {
	environment *objects.Environment
	globals     *objects.Environment
	locals      map[ast.Expr]int
}

func NewInterpreter(globals *objects.Environment) *Interpreter {
	if globals == nil {
		globals = objects.NewEnvironment(nil)
	}
	globals.Define("clock", objects.NewClock())
	return &Interpreter{
		environment: globals,
		globals:     globals,
		locals:      make(map[ast.Expr]int),
	}
}

func (i *Interpreter) SetLocals(locals map[ast.Expr]int) {
	i.locals = locals
}

func (i *Interpreter) Interpret(stmts []ast.Stmt) error {
	for _, stmt := range stmts {
		if _, err := i.evalStmt(stmt); err != nil {
			return err
		}
	}
	return nil
}

// ExecuteBlock executes a block in a provided environment.
func (i *Interpreter) ExecuteBlock(block *ast.BlockStmt, env *objects.Environment) (interface{}, error) {
	previous := i.environment
	i.environment = env
	defer func() { i.environment = previous }()

	for _, stmt := range block.Statements {
		if _, err := i.evalStmt(stmt); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

// Statements

func (i *Interpreter) evalStmt(stmt ast.Stmt) (interface{}, error) {
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		newEnv := objects.NewEnvironment(i.environment)
		return i.ExecuteBlock(s, newEnv)
	case *ast.FunctionStmt:
		return i.visitFunction(s)
	case *ast.ClassStmt:
		return i.visitClass(s)
	case *ast.ReturnStmt:
		return i.visitReturnStmt(s)
	case *ast.VarDeclStmt:
		return i.visitVarDeclStmt(s)
	case *ast.PrintStmt:
		return i.visitPrintStmt(s)
	case *ast.ExprStmt:
		return i.visitExprStmt(s)
	case *ast.IfStmt:
		return i.visitIfStmt(s)
	case *ast.WhileStmt:
		return i.visitWhileStmt(s)
	case *ast.BreakStmt:
		return nil, &objects.BreakError{}
	default:
		return nil, nil
	}
}

func (i *Interpreter) visitFunction(function *ast.FunctionStmt) (interface{}, error) {
	fn := objects.NewCallableFunction(function, i.environment, false)
	i.environment.Define(function.Name.Lexeme, fn)
	return nil, nil
}

func (i *Interpreter) visitClass(class *ast.ClassStmt) (interface{}, error) {
	i.environment.Define(class.Name.Lexeme, nil)
	methods := objects.BuildMethods(class.Methods, i.environment)
	callableClass := objects.NewCallableClass(class.Name.Lexeme, methods)
	if err := i.environment.Assign(class.Name.Lexeme, callableClass); err != nil {
		return nil, err
	}
	return nil, nil
}

func (i *Interpreter) visitReturnStmt(ret *ast.ReturnStmt) (interface{}, error) {
	var value interface{}
	if ret.Value != nil {
		v, err := i.evalExpr(ret.Value)
		if err != nil {
			return nil, err
		}
		value = v
	}
	return value, &objects.ReturnError{Value: value}
}

func (i *Interpreter) visitVarDeclStmt(decl *ast.VarDeclStmt) (interface{}, error) {
	if decl.Initializer != nil {
		val, err := i.evalExpr(decl.Initializer)
		if err != nil {
			return nil, err
		}
		i.environment.Define(decl.Name.Lexeme, val)
	} else {
		i.environment.Define(decl.Name.Lexeme, nil)
	}
	return nil, nil
}

func (i *Interpreter) visitPrintStmt(printStmt *ast.PrintStmt) (interface{}, error) {
	value, err := i.evalExpr(printStmt.Expr)
	if err != nil {
		return nil, err
	}
	fmt.Println(stringify(value))
	return value, nil
}

func (i *Interpreter) visitExprStmt(exprStmt *ast.ExprStmt) (interface{}, error) {
	value, err := i.evalExpr(exprStmt.Expr)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (i *Interpreter) visitIfStmt(ifStmt *ast.IfStmt) (interface{}, error) {
	cond, err := i.evalExpr(ifStmt.Condition)
	if err != nil {
		return nil, err
	}
	if i.isTruthy(cond) {
		return i.evalStmt(ifStmt.ThenBranch)
	} else if ifStmt.ElseBranch != nil {
		return i.evalStmt(ifStmt.ElseBranch)
	}
	return nil, nil
}

func (i *Interpreter) visitWhileStmt(whileStmt *ast.WhileStmt) (interface{}, error) {
	for {
		cond, err := i.evalExpr(whileStmt.Condition)
		if err != nil {
			return nil, err
		}
		if !i.isTruthy(cond) {
			break
		}
		if _, err = i.evalStmt(whileStmt.Body); err != nil {
			if _, ok := err.(*objects.BreakError); ok {
				break
			}
			return nil, err
		}
	}
	return nil, nil
}

// Expressions

func (i *Interpreter) evalExpr(expr ast.Expr) (interface{}, error) {
	switch e := expr.(type) {
	case *ast.BinaryExpr:
		return i.visitBinaryExpr(e)
	case *ast.GroupingExpr:
		return i.evalExpr(e.Expr)
	case *ast.LiteralExpr:
		return e.Value, nil
	case *ast.UnaryExpr:
		return i.visitUnaryExpr(e)
	case *ast.VariableExpr:
		return i.visitVariableExpr(e)
	case *ast.AssignExpr:
		return i.visitAssignExpr(e)
	case *ast.LogicalExpr:
		return i.visitLogicalExpr(e)
	case *ast.CallExpr:
		return i.visitCallExpr(e)
	case *ast.GetExpr:
		return i.visitGetExpr(e)
	case *ast.SetExpr:
		return i.visitSetExpr(e)
	case *ast.ThisExpr:
		return i.visitThisExpr(e)
	default:
		return nil, nil
	}
}

func (i *Interpreter) visitBinaryExpr(exp *ast.BinaryExpr) (interface{}, error) {
	right, err := i.evalExpr(exp.Right)
	if err != nil {
		return nil, err
	}
	left, err := i.evalExpr(exp.Left)
	if err != nil {
		return nil, err
	}
	switch exp.Operator.Type {
	case token.PLUS:
		leftNum, leftIsNum := left.(float64)
		rightNum, rightIsNum := right.(float64)
		leftStr, leftIsStr := left.(string)
		rightStr, rightIsStr := right.(string)

		if leftIsNum && rightIsNum {
			return leftNum + rightNum, nil
		} else if leftIsStr && rightIsStr {
			return leftStr + rightStr, nil
		} else if leftIsStr && rightIsNum {
			return leftStr + fmt.Sprintf("%g", rightNum), nil
		} else if leftIsNum && rightIsStr {
			return fmt.Sprintf("%g", leftNum) + rightStr, nil
		}
		return nil, fmt.Errorf("operands to '+' must both be numbers or both be strings (left: %T=%v, right: %T=%v)", left, left, right, right)
	case token.MINUS, token.STAR, token.SLASH, token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL:
		if !i.isNumber(left) || !i.isNumber(right) {
			return nil, errors.New("operand must be a number")
		}
		lf := left.(float64)
		rf := right.(float64)
		switch exp.Operator.Type {
		case token.MINUS:
			return lf - rf, nil
		case token.STAR:
			return lf * rf, nil
		case token.SLASH:
			if rf == 0 {
				return nil, errors.New("division by zero")
			}
			return lf / rf, nil
		case token.GREATER:
			return lf > rf, nil
		case token.GREATER_EQUAL:
			return lf >= rf, nil
		case token.LESS:
			return lf < rf, nil
		case token.LESS_EQUAL:
			return lf <= rf, nil
		}
	case token.EQUAL_EQUAL:
		return i.isEqual(left, right), nil
	case token.BANG_EQUAL:
		return !i.isEqual(left, right), nil
	}
	return nil, errors.New("invalid operator")
}

func (i *Interpreter) visitUnaryExpr(unary *ast.UnaryExpr) (interface{}, error) {
	right, err := i.evalExpr(unary.Expr)
	if err != nil {
		return nil, err
	}
	switch unary.Operator.Type {
	case token.MINUS:
		return -right.(float64), nil
	case token.BANG:
		return !i.isTruthy(right), nil
	}
	return nil, errors.New("invalid operator")
}

func (i *Interpreter) visitVariableExpr(variable *ast.VariableExpr) (interface{}, error) {
	return i.findVariable(variable, variable.Name)
}

func (i *Interpreter) visitThisExpr(t *ast.ThisExpr) (interface{}, error) {
	return i.findVariable(t, t.Keyword)
}

func (i *Interpreter) findVariable(expr ast.Expr, name token.Token) (interface{}, error) {
	if dist, ok := i.locals[expr]; ok {
		return i.environment.GetAt(dist, name.Lexeme)
	}
	return i.globals.Get(name.Lexeme)
}

func (i *Interpreter) visitAssignExpr(assign *ast.AssignExpr) (interface{}, error) {
	value, err := i.evalExpr(assign.Value)
	if err != nil {
		return nil, err
	}
	if dist, ok := i.locals[assign]; ok {
		if err := i.environment.AssignAt(dist, assign.Name.Lexeme, value); err != nil {
			return nil, err
		}
	} else {
		if err := i.globals.Assign(assign.Name.Lexeme, value); err != nil {
			return nil, err
		}
	}
	return value, nil
}

func (i *Interpreter) visitLogicalExpr(logical *ast.LogicalExpr) (interface{}, error) {
	left, err := i.evalExpr(logical.Left)
	if err != nil {
		return nil, err
	}
	if (logical.Operator.Type == token.OR && i.isTruthy(left)) ||
		(logical.Operator.Type == token.AND && !i.isTruthy(left)) {
		return left, nil
	}
	return i.evalExpr(logical.Right)
}

func (i *Interpreter) visitCallExpr(call *ast.CallExpr) (interface{}, error) {
	callee, err := i.evalExpr(call.Callee)
	if err != nil {
		return nil, err
	}
	args := make([]interface{}, 0, len(call.Arguments))
	for _, arg := range call.Arguments {
		val, err := i.evalExpr(arg)
		if err != nil {
			return nil, err
		}
		args = append(args, val)
	}
	callable, ok := callee.(objects.Callable)
	if !ok {
		return nil, errors.New("can only call functions or classes")
	}
	if callable.Arity() != len(args) {
		return nil, errors.New("expected " + strconv.Itoa(callable.Arity()) + " arguments but got " + strconv.Itoa(len(args)))
	}
	return callable.Call(i, args)
}

func (i *Interpreter) visitGetExpr(get *ast.GetExpr) (interface{}, error) {
	object, err := i.evalExpr(get.Object)
	if err != nil {
		return nil, err
	}
	instance, ok := object.(*objects.ClassInstance)
	if !ok {
		return nil, errors.New("object is not a class instance")
	}
	return instance.Get(get.Name)
}

func (i *Interpreter) visitSetExpr(set *ast.SetExpr) (interface{}, error) {
	object, err := i.evalExpr(set.Object)
	if err != nil {
		return nil, err
	}
	instance, ok := object.(*objects.ClassInstance)
	if !ok {
		return nil, errors.New("object is not a class instance")
	}
	value, err := i.evalExpr(set.Value)
	if err != nil {
		return nil, err
	}
	if err := instance.Set(set.Name, value); err != nil {
		return nil, err
	}
	return value, nil
}

// Utilities

func (i *Interpreter) isTruthy(value interface{}) bool {
	if value == nil {
		return false
	}
	if b, ok := value.(bool); ok {
		return b
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

func stringify(value interface{}) string {
	switch v := value.(type) {
	case nil:
		return "nil"
	case fmt.Stringer:
		return v.String()
	case float64:
		if v == math.Trunc(v) {
			return fmt.Sprintf("%.0f", v)
		}
		return fmt.Sprintf("%g", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
