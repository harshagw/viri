package internal

import (
	"errors"
)

var (
	ErrParse = errors.New("parse error")
)


type Parser struct {
	viri *Viri
	tokens []Token
	current int
}

func NewParser(tokens []Token, viri *Viri) *Parser {
	return &Parser{
		tokens: tokens,
		current: 0,
		viri: viri,
	}
}

func (p *Parser) parse() (Expr, error) {
	return p.parseExpr()
}

func (p *Parser) parseExpr() (Expr, error) {
	return p.parseEquality()
}

func (p *Parser) parseEquality() (Expr, error) {
		expr, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		for p.match(EQUAL_EQUAL, BANG_EQUAL) {
			operator := p.previous()
			right, err := p.parseComparison()
			if err != nil {
				return nil, err
			}
			expr = &BinaryExp{
				Left: expr,
				Right: right,
				Operator: operator,
			}
		}
		return expr, nil
}

func (p *Parser) parseComparison() (Expr, error) {
	expr, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	for p.match(GREATER, GREATER_EQUAL, LESS, LESS_EQUAL) {
		operator := p.previous()
		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		expr = &BinaryExp{
			Left: expr,
			Right: right,
			Operator: operator,
		}
	}
	return expr, nil
}

func (p *Parser) parseTerm() (Expr, error) {
	expr, err := p.parseFactor()
	if err != nil {
		return nil, err
	}
	for p.match(MINUS, PLUS) {
		operator := p.previous()
		right, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		expr = &BinaryExp{
			Left: expr,
			Right: right,
			Operator: operator,
		}
	}
	return expr, nil
}

func (p *Parser) parseFactor() (Expr, error) {
	expr, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.match(SLASH, STAR) {
		operator := p.previous()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		expr = &BinaryExp{
			Left: expr,
			Right: right,
			Operator: operator,
		}
	}
	return expr, nil
}

func (p *Parser) parseUnary() (Expr, error) {
	if p.match(BANG, MINUS) {
		operator := p.previous()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &Unary{
			Operator: operator,
			Expr: right,
		}, nil
	}
	return p.parsePrimary()
}

func (p *Parser) parsePrimary() (Expr, error) {
	if p.match(NUMBER, STRING, TRUE, FALSE, NIL) {
		return &Literal{
			Value: p.previous().Literal,
		},nil	
	}
	if p.match(LEFT_PAREN) {
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		p.consume(RIGHT_PAREN, "Expect ')' after expression.")
		return &Grouping{
			Expr: expr,
		}, nil
	}
	return nil, p.error(p.peek(), "Expect expression.")
}

// Utility Functions

// Check if we are at the end of file
func (p *Parser) isAtEnd() bool {
	return p.peek().TokenType == EOF
}

// Return the current token without consuming it
func (p *Parser) peek() Token {
	return p.tokens[p.current]
}

// Check if the current token is one of the given type
func (p *Parser) match(tokenTypes ...TokenType) bool {
	for _, tokenType := range tokenTypes {
		if p.check(tokenType) {
			p.advance()
			return true
		}
	}
	return false
}

// Check if the current token is of the given type
func (p *Parser) check(tokenType TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().TokenType == tokenType
}

// Return the previous token
func (p *Parser) previous() Token {
	return p.tokens[p.current - 1]
}

// Advance the current token
func (p *Parser) advance() Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

func (p *Parser) consume(tokenType TokenType, message string) Token {
	if p.check(tokenType) {
		return p.advance()
	}
	p.error(p.peek(), message)
	return p.previous()
}

func (p *Parser) error(token Token, message string) error {
	p.viri.Error(token, message)
	return ErrParse
}

func (p *Parser) synchronize() {
	p.advance()
	for !p.isAtEnd() {
		if p.previous().TokenType == SEMICOLON {
			return
		}
		switch p.peek().TokenType {
		case CLASS, FOR, FUN, IF, PRINT, RETURN, VAR, WHILE:
			return
		}
		p.advance()
	}
}