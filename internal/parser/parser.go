package parser

import (
	"errors"

	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/token"
)

var (
	ErrParse = errors.New("parse error")
)

type Parser struct {
	tokens             []token.Token
	current            int
	diagnosticHandler  objects.DiagnosticHandler
	hadError           bool
}

func NewParser(tokens []token.Token, diagnosticHandler objects.DiagnosticHandler) *Parser {
	return &Parser{
		tokens:            tokens,
		current:           0,
		diagnosticHandler: diagnosticHandler,
	}
}

func (p *Parser) Parse() ([]ast.Stmt, error) {
	statements := []ast.Stmt{}
	for !p.isAtEnd() {
		stmt := p.parseDeclaration()
		if stmt != nil {
			statements = append(statements, stmt)
		}
	}
	if p.hadError {
		return statements, ErrParse
	}
	return statements, nil
}

func (p *Parser) parseDeclaration() ast.Stmt {
	var (
		stmt ast.Stmt
		err  error
	)
	if p.match(token.VAR) {
		stmt, err = p.parseVarDecl()
	} else if p.match(token.FUN) {
		stmt, err = p.parseFunction()
	} else if p.match(token.CLASS) {
		stmt, err = p.parseClass()
	} else {
		stmt, err = p.parseStmt()
	}
	if err != nil {
		p.synchronize()
		return nil
	}
	return stmt
}

func (p *Parser) parseVarDecl() (ast.Stmt, error) {
	name, err := p.consume(token.IDENTIFIER, "Expect variable name.")
	if err != nil {
		return nil, err
	}

	var initializer ast.Expr
	if p.match(token.EQUAL) {
		initializer, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
	}

	if _, err = p.consume(token.SEMICOLON, "Expect ';' after variable declaration."); err != nil {
		return nil, err
	}

	return &ast.VarDeclStmt{
		Name:        *name,
		Initializer: initializer,
	}, nil
}

func (p *Parser) parseStmt() (ast.Stmt, error) {
	if p.match(token.PRINT) {
		return p.parsePrintStmt()
	}
	if p.match(token.LEFT_BRACE) {
		return p.parseBlockStmt()
	}
	if p.match(token.IF) {
		return p.parseIfStmt()
	}
	if p.match(token.WHILE) {
		return p.parseWhileStmt()
	}
	if p.match(token.FOR) {
		return p.parseForStmt()
	}
	if p.match(token.BREAK) {
		return p.parseBreakStmt()
	}
	if p.match(token.RETURN) {
		return p.parseReturnStmt()
	}

	return p.parseExprStmt()
}

func (p *Parser) parseReturnStmt() (ast.Stmt, error) {
	keyword := p.peekPrevious()
	var value ast.Expr
	var err error

	if !p.check(token.SEMICOLON) {
		value, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
	}

	if _, err = p.consume(token.SEMICOLON, "Expect ';' after return statement."); err != nil {
		return nil, err
	}
	return &ast.ReturnStmt{
		Keyword: *keyword,
		Value:   value,
	}, nil
}

func (p *Parser) parseClass() (ast.Stmt, error) {
	name, err := p.consume(token.IDENTIFIER, "Expect class name.")
	if err != nil {
		return nil, err
	}
	var superclass *ast.VariableExpr
	if p.match(token.LESS) {
		superClassName, err := p.consume(token.IDENTIFIER, "Expect superclass name.")
		if err != nil {
			return nil, err
		}
		superclass = &ast.VariableExpr{
			Name: *superClassName,
		}
	}
	if _, err = p.consume(token.LEFT_BRACE, "Expect '{' before class body."); err != nil {
		return nil, err
	}
	methods := make([]*ast.FunctionStmt, 0)
	for !p.isAtEnd() && !p.check(token.RIGHT_BRACE) {
		method, err := p.parseFunction()
		if err != nil {
			return nil, err
		}
		methods = append(methods, method.(*ast.FunctionStmt))
	}
	if _, err = p.consume(token.RIGHT_BRACE, "Expect '}' after class body."); err != nil {
		return nil, err
	}
	return &ast.ClassStmt{
		Name:    *name,
		Methods: methods,
		SuperClass: superclass,
	}, nil
}

func (p *Parser) parseFunction() (ast.Stmt, error) {
	name, err := p.consume(token.IDENTIFIER, "Expect function name.")
	if err != nil {
		return nil, err
	}
	if _, err = p.consume(token.LEFT_PAREN, "Expect '(' after function name."); err != nil {
		return nil, err
	}
	parameters := make([]token.Token, 0)
	if !p.check(token.RIGHT_PAREN) {
		for {
			if len(parameters) >= 255 {
				return nil, p.error(p.peekPrevious(), "Can't have more than 255 parameters.")
			}

			parameter, err := p.consume(token.IDENTIFIER, "Expect parameter name.")
			if err != nil {
				return nil, err
			}
			parameters = append(parameters, *parameter)
			if !p.match(token.COMMA) {
				break
			}
		}
	}
	if _, err = p.consume(token.RIGHT_PAREN, "Expect ')' after parameters."); err != nil {
		return nil, err
	}
	if _, err = p.consume(token.LEFT_BRACE, "Expect '{' before block."); err != nil {
		return nil, err
	}
	body, err := p.parseBlockStmt()
	if err != nil {
		return nil, err
	}
	return &ast.FunctionStmt{
		Name:   *name,
		Params: parameters,
		Body:   body,
	}, nil
}

func (p *Parser) parseWhileStmt() (ast.Stmt, error) {
	if _, err := p.consume(token.LEFT_PAREN, "Expect '(' after while."); err != nil {
		return nil, err
	}
	condition, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if _, err = p.consume(token.RIGHT_PAREN, "Expect ')' after while condition."); err != nil {
		return nil, err
	}
	body, err := p.parseStmt()
	if err != nil {
		return nil, err
	}
	return &ast.WhileStmt{
		Condition: condition,
		Body:      body,
	}, nil
}

func (p *Parser) parseForStmt() (ast.Stmt, error) {
	if _, err := p.consume(token.LEFT_PAREN, "Expect '(' after for."); err != nil {
		return nil, err
	}

	var initializer ast.Stmt
	var err error

	if p.match(token.VAR) {
		initializer, err = p.parseVarDecl()
		if err != nil {
			return nil, err
		}
	} else if p.match(token.SEMICOLON) {
		initializer = nil
	} else {
		initializer, err = p.parseExprStmt()
		if err != nil {
			return nil, err
		}
	}

	var condition ast.Expr
	if !p.check(token.SEMICOLON) {
		condition, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
	}

	if _, err = p.consume(token.SEMICOLON, "Expect ';' after for condition."); err != nil {
		return nil, err
	}

	var increment ast.Expr
	if !p.check(token.RIGHT_PAREN) {
		increment, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
	}

	if _, err = p.consume(token.RIGHT_PAREN, "Expect ')' after for."); err != nil {
		return nil, err
	}

	body, err := p.parseStmt()
	if err != nil {
		return nil, err
	}

	if increment != nil {
		body = &ast.BlockStmt{
			Statements: []ast.Stmt{
				body,
				&ast.ExprStmt{Expr: increment},
			},
		}
	}

	if condition == nil {
		condition = &ast.LiteralExpr{Value: true}
	}

	body = &ast.WhileStmt{
		Condition: condition,
		Body:      body,
	}

	if initializer != nil {
		body = &ast.BlockStmt{
			Statements: []ast.Stmt{initializer, body},
		}
	}

	return body, nil
}

func (p *Parser) parseIfStmt() (ast.Stmt, error) {
	if _, err := p.consume(token.LEFT_PAREN, "Expect '(' after if."); err != nil {
		return nil, err
	}
	condition, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if _, err = p.consume(token.RIGHT_PAREN, "Expect ')' after if condition."); err != nil {
		return nil, err
	}
	thenBranch, err := p.parseStmt()
	if err != nil {
		return nil, err
	}
	if !p.match(token.ELSE) {
		return &ast.IfStmt{
			Condition:  condition,
			ThenBranch: thenBranch,
		}, nil
	}
	elseBranch, err := p.parseStmt()
	if err != nil {
		return nil, err
	}
	return &ast.IfStmt{
		Condition:  condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
	}, nil
}

func (p *Parser) parseBlockStmt() (*ast.BlockStmt, error) {
	statements := []ast.Stmt{}
	for !p.isAtEnd() && !p.check(token.RIGHT_BRACE) {
		stmt := p.parseDeclaration()
		if stmt != nil {
			statements = append(statements, stmt)
		}
	}
	if _, err := p.consume(token.RIGHT_BRACE, "Expect '}' after block."); err != nil {
		return nil, err
	}
	return &ast.BlockStmt{
		Statements: statements,
	}, nil
}

func (p *Parser) parsePrintStmt() (ast.Stmt, error) {
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if _, err = p.consume(token.SEMICOLON, "Expect ';' after print expression."); err != nil {
		return nil, err
	}
	return &ast.PrintStmt{
		Expr: expr,
	}, nil
}

func (p *Parser) parseBreakStmt() (ast.Stmt, error) {
	keyword := p.peekPrevious()
	if _, err := p.consume(token.SEMICOLON, "Expect ';' after break."); err != nil {
		return nil, err
	}
	return &ast.BreakStmt{
		Keyword: *keyword,
	}, nil
}

func (p *Parser) parseExprStmt() (ast.Stmt, error) {
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if _, err = p.consume(token.SEMICOLON, "Expect ';' after expression statement."); err != nil {
		return nil, err
	}
	return &ast.ExprStmt{
		Expr: expr,
	}, nil
}

func (p *Parser) parseExpr() (ast.Expr, error) {
	return p.parseAssignment()
}

func (p *Parser) parseAssignment() (ast.Expr, error) {
	expr, err := p.parseLogicOr()
	if err != nil {
		return nil, err
	}

	if p.match(token.EQUAL) {
		assignment := p.peekPrevious()
		value, err := p.parseAssignment()
		if err != nil {
			return nil, err
		}

		switch target := expr.(type) {
		case *ast.VariableExpr:
			return &ast.AssignExpr{
				Name:  target.Name,
				Value: value,
			}, nil
		case *ast.GetExpr:
			return &ast.SetExpr{
				Object: target.Object,
				Name:   target.Name,
				Value:  value,
			}, nil
		default:
			return nil, p.error(assignment, "Invalid assignment target")
		}
	}

	return expr, nil
}

func (p *Parser) parseLogicOr() (ast.Expr, error) {
	expr, err := p.parseLogicAnd()
	if err != nil {
		return nil, err
	}
	for p.match(token.OR) {
		operator := p.peekPrevious()
		right, err := p.parseLogicAnd()
		if err != nil {
			return nil, err
		}
		expr = &ast.LogicalExpr{
			Left:     expr,
			Operator: *operator,
			Right:    right,
		}
	}
	return expr, nil
}

func (p *Parser) parseLogicAnd() (ast.Expr, error) {
	expr, err := p.parseEquality()
	if err != nil {
		return nil, err
	}
	for p.match(token.AND) {
		operator := p.peekPrevious()
		right, err := p.parseEquality()
		if err != nil {
			return nil, err
		}
		expr = &ast.LogicalExpr{
			Left:     expr,
			Operator: *operator,
			Right:    right,
		}
	}
	return expr, nil
}

func (p *Parser) parseEquality() (ast.Expr, error) {
	expr, err := p.parseComparison()
	if err != nil {
		return nil, err
	}
	for p.match(token.EQUAL_EQUAL, token.BANG_EQUAL) {
		operator := p.peekPrevious()
		right, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		expr = &ast.BinaryExpr{
			Left:     expr,
			Right:    right,
			Operator: *operator,
		}
	}
	return expr, nil
}

func (p *Parser) parseComparison() (ast.Expr, error) {
	expr, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	for p.match(token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL) {
		operator := p.peekPrevious()
		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		expr = &ast.BinaryExpr{
			Left:     expr,
			Right:    right,
			Operator: *operator,
		}
	}
	return expr, nil
}

func (p *Parser) parseTerm() (ast.Expr, error) {
	expr, err := p.parseFactor()
	if err != nil {
		return nil, err
	}
	for p.match(token.MINUS, token.PLUS) {
		operator := p.peekPrevious()
		right, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		expr = &ast.BinaryExpr{
			Left:     expr,
			Right:    right,
			Operator: *operator,
		}
	}
	return expr, nil
}

func (p *Parser) parseFactor() (ast.Expr, error) {
	expr, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.match(token.SLASH, token.STAR) {
		operator := p.peekPrevious()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		expr = &ast.BinaryExpr{
			Left:     expr,
			Right:    right,
			Operator: *operator,
		}
	}
	return expr, nil
}

func (p *Parser) parseUnary() (ast.Expr, error) {
	if p.match(token.BANG, token.MINUS) {
		operator := p.peekPrevious()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryExpr{
			Operator: *operator,
			Expr:     right,
		}, nil
	}
	return p.parseCall()
}

func (p *Parser) parseCall() (ast.Expr, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	for {
		if p.match(token.LEFT_PAREN) {
			expr, err = p.parseFinishCall(expr)
			if err != nil {
				return nil, err
			}
		} else if p.match(token.DOT) {
			ident, err := p.consume(token.IDENTIFIER, "Expect property name after '.'.")
			if err != nil {
				return nil, err
			}
			expr = &ast.GetExpr{
				Object: expr,
				Name:   *ident,
			}
		} else {
			break
		}
	}
	return expr, nil
}

func (p *Parser) parseFinishCall(expr ast.Expr) (ast.Expr, error) {
	arguments := make([]ast.Expr, 0)

	if !p.check(token.RIGHT_PAREN) {
		for {
			if len(arguments) >= 255 {
				return nil, p.error(p.peekPrevious(), "Can't have more than 255 arguments.")
			}
			arg, err := p.parseExpr()
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
		Callee:       expr,
		Arguments:    arguments,
		ClosingParen: *p.peekPrevious(),
	}, nil
}

func (p *Parser) parsePrimary() (ast.Expr, error) {
	if p.match(token.NUMBER, token.STRING, token.TRUE, token.FALSE, token.NIL) {
		return &ast.LiteralExpr{
			Value: p.peekPrevious().Literal,
		}, nil
	}
	if p.match(token.IDENTIFIER) {
		return &ast.VariableExpr{
			Name: *p.peekPrevious(),
		}, nil
	}
	if p.match(token.THIS) {
		return &ast.ThisExpr{
			Keyword: *p.peekPrevious(),
		}, nil
	}
	if p.match(token.SUPER){
		superToken := p.peekPrevious()
		if _, err := p.consume(token.DOT, "Expect '.' after 'super'."); err != nil {
			return nil, err
		}
		ident, err := p.consume(token.IDENTIFIER, "Expect superclass method name.")
		if err != nil {
			return nil, err
		}
		return &ast.SuperExpr{
			Keyword: *superToken,
			Method:  *ident,
		}, nil
	}
	if p.match(token.LEFT_PAREN) {
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if _, err = p.consume(token.RIGHT_PAREN, "Expect ')' after expression."); err != nil {
			return nil, err
		}
		return &ast.GroupingExpr{
			Expr: expr,
		}, nil
	}
	return nil, p.error(p.peekCurrent(), "Expect expression.")
}

// Utility functions

func (p *Parser) isAtEnd() bool {
	return p.peekCurrent().Type == token.EOF
}

func (p *Parser) peekCurrent() *token.Token {
	if p.current >= len(p.tokens) {
		return &p.tokens[len(p.tokens)-1]
	}
	return &p.tokens[p.current]
}

func (p *Parser) match(tokenTypes ...token.Type) bool {
	for _, tokenType := range tokenTypes {
		if p.check(tokenType) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) check(tokenType token.Type) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peekCurrent().Type == tokenType
}

func (p *Parser) peekPrevious() *token.Token {
	if p.current == 0 {
		return nil
	}
	return &p.tokens[p.current-1]
}

func (p *Parser) advance() *token.Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.peekPrevious()
}

func (p *Parser) consume(tokenType token.Type, message string) (*token.Token, error) {
	if p.check(tokenType) {
		return p.advance(), nil
	}
	return nil, p.error(p.peekPrevious(), message)
}

func (p *Parser) error(tok *token.Token, message string) error {
	p.hadError = true
	if tok != nil && p.diagnosticHandler != nil {
		p.diagnosticHandler.Error(*tok, message)
	}
	return ErrParse
}

func (p *Parser) synchronize() {
	p.advance()

	for !p.isAtEnd() {
		if p.peekPrevious().Type == token.SEMICOLON {
			return
		}

		switch p.peekCurrent().Type {
		case token.CLASS, token.FOR, token.FUN, token.IF, token.PRINT, token.RETURN, token.VAR, token.WHILE:
			return
		}

		p.advance()
	}
}
