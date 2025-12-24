package compiler

import (
	"fmt"

	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/code"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/token"
)

type Compiler struct {
	instructions      code.Instructions
	constants         []objects.Object
	diagnosticHandler objects.DiagnosticHandler
}

func New(diagnosticHandler objects.DiagnosticHandler) *Compiler {
	return &Compiler{
		instructions:      code.Instructions{},
		constants:         []objects.Object{},
		diagnosticHandler: diagnosticHandler,
	}
}

func (c *Compiler) Compile(node interface{}) error {
	switch node := node.(type) {
	case ast.Stmt:
		return c.compileStatement(node)
	case ast.Expr:
		return c.compileExpression(node)
	default:
		return fmt.Errorf("unknown node type: %T", node)
	}
}

func (c *Compiler) compileStatement(stmt ast.Stmt) error {
	switch stmt := stmt.(type) {
	case *ast.ExprStmt:
		if err := c.compileExpression(stmt.Expr); err != nil {
			return err
		}
		c.emit(code.OpPop)
		return nil
	default:
		return fmt.Errorf("unsupported statement type: %T", stmt)
	}
}

func (c *Compiler) compileExpression(node ast.Expr) error {
	switch node := node.(type) {
	case *ast.LiteralExpr:
		switch value := node.Value.(type) {
		case int:
			integer := &objects.Number{Value: float64(value)}
			c.emitConstant(integer)
		case float64:
			number := &objects.Number{Value: value}
			c.emitConstant(number)
		case bool:
			if value {
				c.emit(code.OpTrue)
			} else {
				c.emit(code.OpFalse)
			}
		}

	case *ast.UnaryExpr:
		if err := c.compileExpression(node.Expr); err != nil {
			return err
		}

		switch node.Operator.Type {
		case token.BANG:
			c.emit(code.OpBang)
		case token.MINUS:
			c.emit(code.OpMinus)
		default:
			return c.error(node.Operator, fmt.Sprintf("unknown unary operator %s", node.Operator.Lexeme))
		}

	case *ast.GroupingExpr:
		return c.compileExpression(node.Expr)

	case *ast.BinaryExpr:
		if node.Operator.Type == token.LESS {
			if err := c.compileExpression(node.Right); err != nil {
				return err
			}
			if err := c.compileExpression(node.Left); err != nil {
				return err
			}
			c.emit(code.OpGreaterThan)
			return nil
		}

		if err := c.compileExpression(node.Left); err != nil {
			return err
		}
		if err := c.compileExpression(node.Right); err != nil {
			return err
		}

		switch node.Operator.Type {
		case token.PLUS:
			c.emit(code.OpAdd)
		case token.MINUS:
			c.emit(code.OpSub)
		case token.STAR:
			c.emit(code.OpMul)
		case token.SLASH:
			c.emit(code.OpDiv)
		case token.GREATER:
			c.emit(code.OpGreaterThan)
		case token.EQUAL_EQUAL:
			c.emit(code.OpEqual)
		case token.BANG_EQUAL:
			c.emit(code.OpNotEqual)
		default:
			return c.error(node.Operator, fmt.Sprintf("unknown operator %s", node.Operator.Lexeme))
		}
	}
	return nil
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)
	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return posNewInstruction
}

func (c *Compiler) emitConstant(obj objects.Object) int {
	c.constants = append(c.constants, obj)
	return c.emit(code.OpConstant, len(c.constants)-1)
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []objects.Object
}

func (c *Compiler) error(tok *token.Token, message string) error {
	if c.diagnosticHandler != nil && tok != nil {
		c.diagnosticHandler.Error(*tok, message)
	}
	return fmt.Errorf("%s", message)
}

func (c *Compiler) warn(tok *token.Token, message string) {
	if c.diagnosticHandler != nil && tok != nil {
		c.diagnosticHandler.Warn(*tok, message)
	}
}
