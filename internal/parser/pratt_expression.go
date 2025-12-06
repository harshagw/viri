package parser

import (
	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/token"
)

// precedence establishes binding power for infix operators.
// Higher value means tighter binding.
type precedence int

const (
	precNone precedence = iota
	precAssignment
	precOr
	precAnd
	precEquality
	precComparison
	precTerm
	precFactor
	precUnary
	precCall
)

var infixPrecedence = map[token.Type]precedence{
	token.EQUAL:         precAssignment,
	token.OR:            precOr,
	token.AND:           precAnd,
	token.EQUAL_EQUAL:   precEquality,
	token.BANG_EQUAL:    precEquality,
	token.GREATER:       precComparison,
	token.GREATER_EQUAL: precComparison,
	token.LESS:          precComparison,
	token.LESS_EQUAL:    precComparison,
	token.MINUS:         precTerm,
	token.PLUS:          precTerm,
	token.SLASH:         precFactor,
	token.STAR:          precFactor,
	token.LEFT_PAREN:    precCall,
	token.DOT:           precCall,
}

func infixBindingPower(tt token.Type) precedence {
	if prec, ok := infixPrecedence[tt]; ok {
		return prec
	}
	return precNone
}

// parseExpr is the public entrypoint used by the statement parser.
func (p *Parser) parseExpr() (ast.Expr, error) {
	return p.parseExpression(precNone)
}

// parseExpression implements a Pratt parser for expressions.
func (p *Parser) parseExpression(minPrec precedence) (ast.Expr, error) {
	tok := p.peekCurrent()
	p.advance()
	left, err := p.parsePrefix(tok)
	if err != nil {
		return nil, err
	}

	for {
		next := p.peekCurrent() // this usually gives out operator token
		nextPrec := infixBindingPower(next.Type)
		if minPrec >= nextPrec {
			break
		}
		op := p.advance() // op = next actually
		if op == nil {
			break
		}
		left, err = p.parseInfix(left, op)
		if err != nil {
			return nil, err
		}
	}

	return left, nil
}

func (p *Parser) parsePrefix(tok *token.Token) (ast.Expr, error) {
	switch tok.Type {
	case token.NUMBER, token.STRING, token.TRUE, token.FALSE, token.NIL:
		return &ast.LiteralExpr{Value: tok.Literal}, nil
	case token.IDENTIFIER:
		return &ast.VariableExpr{Name: *tok}, nil
	case token.THIS:
		return &ast.ThisExpr{Keyword: *tok}, nil
	case token.SUPER:
		if _, err := p.consume(token.DOT, "Expect '.' after 'super'."); err != nil {
			return nil, err
		}
		method, err := p.consume(token.IDENTIFIER, "Expect superclass method name.")
		if err != nil {
			return nil, err
		}
		return &ast.SuperExpr{Keyword: *tok, Method: *method}, nil
	case token.BANG, token.MINUS:
		right, err := p.parseExpression(precUnary)
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpr{Operator: *tok, Expr: right}, nil
	case token.LEFT_PAREN:
		expr, err := p.parseExpression(precNone)
		if err != nil {
			return nil, err
		}
		if _, err = p.consume(token.RIGHT_PAREN, "Expect ')' after expression."); err != nil {
			return nil, err
		}
		return &ast.GroupingExpr{Expr: expr}, nil
	default:
		return nil, p.error(tok, "Expect expression.")
	}
}

func (p *Parser) parseInfix(left ast.Expr, operator *token.Token) (ast.Expr, error) {
	switch operator.Type {
	case token.LEFT_PAREN:
		return p.finishCall(left)
	case token.DOT:
		ident, err := p.consume(token.IDENTIFIER, "Expect property name after '.'.")
		if err != nil {
			return nil, err
		}
		return &ast.GetExpr{Object: left, Name: *ident}, nil
	case token.EQUAL:
		right, err := p.parseExpression(infixBindingPower(operator.Type) - 1) // -1 because equal is right associative
		if err != nil {
			return nil, err
		}
		switch target := left.(type) {
		case *ast.VariableExpr:
			return &ast.AssignExpr{Name: target.Name, Value: right}, nil
		case *ast.GetExpr:
			return &ast.SetExpr{Object: target.Object, Name: target.Name, Value: right}, nil
		default:
			return nil, p.error(operator, "Invalid assignment target")
		}
	case token.OR, token.AND:
		right, err := p.parseExpression(infixBindingPower(operator.Type))
		if err != nil {
			return nil, err
		}
		return &ast.LogicalExpr{Left: left, Operator: *operator, Right: right}, nil
	case token.EQUAL_EQUAL, token.BANG_EQUAL,
		token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL,
		token.MINUS, token.PLUS, token.SLASH, token.STAR:
		right, err := p.parseExpression(infixBindingPower(operator.Type))
		if err != nil {
			return nil, err
		}
		return &ast.BinaryExpr{Left: left, Right: right, Operator: *operator}, nil
	default:
		return left, nil
	}
}

func (p *Parser) finishCall(callee ast.Expr) (ast.Expr, error) {
	arguments := make([]ast.Expr, 0)

	if !p.check(token.RIGHT_PAREN) {
		for {
			if len(arguments) >= 255 {
				return nil, p.error(p.peekPrevious(), "Can't have more than 255 arguments.")
			}
			arg, err := p.parseExpression(precNone)
			if err != nil {
				return nil, err
			}
			arguments = append(arguments, arg)
			if !p.match(token.COMMA) {
				break
			}
		}
	}

	if _, err := p.consume(token.RIGHT_PAREN, "Expect ')' after arguments."); err != nil {
		return nil, err
	}

	return &ast.CallExpr{
		Callee:       callee,
		Arguments:    arguments,
		ClosingParen: *p.peekPrevious(),
	}, nil
}
