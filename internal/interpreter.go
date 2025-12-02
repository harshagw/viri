package internal

import (
	"errors"
)

type Interpreter struct {
	viri *Viri
}

func NewInterpreter(viri *Viri) *Interpreter {
	return &Interpreter{
		viri: viri,
	}
}

func (i *Interpreter) Interpret(stmts []Stmt) (error) {
	for _, stmt := range stmts {
		_, err := stmt.Accept(i)
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Interpreter) evaluateExpr(expr Expr) (interface{}, error) {
	return expr.Accept(i)
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

func (i *Interpreter) visitBinaryExp(binaryExp *BinaryExp) (interface{}, error) {
	rightExp, err := i.evaluateExpr(binaryExp.Right);
	if err != nil {
		return nil, err
	}
	leftExp, err := i.evaluateExpr(binaryExp.Left);
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
			return leftStr + rightStr,nil
		} else {
			return nil, errors.New("operands to '+' must both be numbers or both be strings")
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
		return leftExp.(float64) * rightExp.(float64), nil
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
	return literal.Value, nil;
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
