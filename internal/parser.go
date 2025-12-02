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
	loopDepth int
}

func NewParser(tokens []Token, viri *Viri) *Parser {
	return &Parser{
		tokens: tokens,
		current: 0,
		viri: viri,
	}
}

func (p *Parser) parse() ([]Stmt) {
	statements := []Stmt{}
	for !p.isAtEnd() {
		stmt := p.parseDeclaration()
		if stmt != nil {
			statements = append(statements, stmt)
		}
	}
	return statements
}

func (p *Parser) parseDeclaration() (Stmt) {
	var (
		stmt Stmt
		err  error
	)
	if p.match(VAR) {
		stmt, err = p.parseVarDecl()
	} else {
		stmt, err = p.parseStmt()
	}
	if err != nil {
		p.synchronize()
		return nil
	}
	return stmt
}

func (p *Parser) parseVarDecl() (Stmt, error) {
	name, err := p.consume(IDENTIFIER, "Expect variable name.")
	if err != nil {
		return nil, err
	}
	
	var initializer Expr
	if p.match(EQUAL) {
		initializer, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
	}
	
	_, err = p.consume(SEMICOLON, "Expect ';' after variable declaration.")
	if err != nil {
		return nil, err
	}
	
	return &VarDeclStmt{
		tokenName:  name.Lexeme,
		initializer: initializer,
	}, nil
}

func (p *Parser) parseStmt() (Stmt, error) {
	if p.match(PRINT) {
		stmt, err := p.parsePrintStmt()
		if err != nil {
			return nil, err
		}
		return stmt, nil	
	}
	if p.match(LEFT_BRACE){
		stmt, err := p.parseBlockStmt()
		if err != nil {
			return nil, err
		}
		return stmt, nil
	}
	if p.match(IF){
		stmt, err := p.parseIfStmt()
		if err != nil {
			return nil, err
		}
		return stmt, nil
	}
	if p.match(WHILE){
		stmt, err := p.parseWhileStmt()
		if err != nil {
			return nil, err
		}
		return stmt, nil
	}
	if p.match(FOR){
		stmt, err := p.parseForStmt()
		if err != nil {
			return nil, err
		}
		return stmt, nil
	}
	if p.match(BREAK){
		stmt, err := p.parseBreakStmt()
		if err != nil {
			return nil, err
		}
		return stmt, nil
	}
	stmts, err := p.parseExprStmt()
	if err != nil {
		return nil, err
	}
	return stmts, nil
}

func (p *Parser) parseWhileStmt() (Stmt, error){
	_, err := p.consume(LEFT_PAREN, "Expect '(' after while.")
	if err != nil {
		return nil, err
	}
	condition, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	_, err = p.consume(RIGHT_PAREN, "Expect ')' after while condition.")
	if err != nil {
		return nil, err
	}
	p.loopDepth++
	body, err := p.parseStmt()
	p.loopDepth--
	if err != nil {
		return nil, err
	}
	return &WhileStmt{
		condition: condition,
		body: body,
	}, nil
}

func (p *Parser) parseForStmt() (Stmt, error){
	_, err := p.consume(LEFT_PAREN, "Expect '(' after for.")
	if err != nil {
		return nil, err
	}
	
	var intializer Stmt;

	if p.match(VAR){
		intializer, err = p.parseVarDecl()
		if err != nil {
			return nil, err
		}
	} else if p.match(SEMICOLON){
		intializer = nil
	} else {
		intializer, err = p.parseExprStmt()
		if err != nil {
			return nil, err
		}
	}
	
	var condition Expr;

	if !p.check(SEMICOLON){
		condition, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
	}

	_, err = p.consume(SEMICOLON, "Expect ';' after for condition.")
	if err != nil {
		return nil, err
	}

	var increment Expr;
	if !p.check(RIGHT_PAREN){
		increment, err = p.parseExpr()
		if err != nil {
			return nil, err
		}
		
	}

	_, err = p.consume(RIGHT_PAREN, "Expect ')' after for.")
	if err != nil {
		return nil, err
	}

	p.loopDepth++
	body, err := p.parseStmt()
	p.loopDepth--
	if err != nil {
		return nil, err
	}

	return &Block{ statements: []Stmt{intializer, &WhileStmt{
		condition: condition,
		body: &Block{
				statements: []Stmt{body, &ExprStmt{Expr: increment}},
			},
		}},
	}, nil
}

func (p *Parser) parseIfStmt() (Stmt, error){
	_, err := p.consume(LEFT_PAREN, "Expect '(' after if.")
	if err != nil {
		return nil, err
	}
	condition, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	_, err = p.consume(RIGHT_PAREN, "Expect ')' after if condition.")
	if err != nil {
		return nil, err
	}
	ifStmt, err := p.parseStmt()
	if err != nil {
		return nil, err
	}
	if !p.match(ELSE){
		return &IfStmt{
			condition: condition,
			ifBranch: ifStmt,
			elseBranch: nil,
		}, nil
	}
	elseStmt, err := p.parseStmt()
	if err != nil {
		return nil, err
	}
	return &IfStmt{
		condition: condition,
		ifBranch: ifStmt,
		elseBranch: elseStmt,
	}, nil
}

func (p *Parser) parseBlockStmt() (Stmt, error){
	statements := []Stmt{}
	for !p.isAtEnd() && !p.check(RIGHT_BRACE) {
		stmt := p.parseDeclaration()
		if stmt != nil {
			statements = append(statements, stmt)
		}
	}
	_, err := p.consume(RIGHT_BRACE, "Expect '}' after block.")
	if err != nil {
		return nil, err
	}
	return &Block{
		statements: statements,
	}, nil
}

func (p *Parser) parsePrintStmt() (Stmt, error) {
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	_, err = p.consume(SEMICOLON, "Expect ';' after print expression.")
	if err != nil {
		return nil, err
	}
	return &PrintStmt{
		Expr: expr,
	}, nil
}

func (p *Parser) parseBreakStmt() (Stmt, error) {
	keyword := p.peekPrevious()
	if p.loopDepth == 0 {
		return nil, p.error(keyword, "break statement must be inside a loop.")
	}
	_, err := p.consume(SEMICOLON, "Expect ';' after break.")
	if err != nil {
		return nil, err
	}
	return &BreakStmt{
		keyword: *keyword,
	}, nil
}

func (p *Parser) parseExprStmt() (Stmt, error) {
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	_, err = p.consume(SEMICOLON, "Expect ';' after expression statement.")
	if err != nil {
		return nil, err
	}
	return &ExprStmt{
		Expr: expr,
	}, nil
}

func (p *Parser) parseExpr() (Expr, error) {
	return p.parseAssignment()
}

func (p *Parser) parseAssignment() (Expr, error) {
	expr,err := p.parseLogicOr()
	if err != nil {
		return nil, err
	}

	if p.match(EQUAL) {
		assignment := p.peekPrevious()
		value, err := p.parseAssignment()
		if err != nil {
			return nil, err
		}

		if expr.Type() != VARIABLE {
			return nil, p.error(assignment, "Expect variable name.")
		}

		variable := expr.(*Variable)
		return &Assignment{
			Name: variable.Name,
			Value: value,
		}, nil
	}

	return expr, nil
}

func (p *Parser) parseLogicOr() (Expr, error) {
	expr, err := p.parseLogicAnd()
	if err != nil {
		return nil, err
	}
	for p.match(OR) {
		operator := p.peekPrevious()
		right, err := p.parseLogicAnd()
		if err != nil {
			return nil, err
		}
		expr = &Logical{
			Left: expr,
			Operator: *operator,
			Right: right,
		}
	}
	return expr, nil
}

func (p *Parser) parseLogicAnd() (Expr, error) {
	expr, err := p.parseEquality()
	if err != nil {
		return nil, err
	}
	for p.match(AND) {
		operator := p.peekPrevious()
		right, err := p.parseEquality()
		if err != nil {
			return nil, err
		}
		expr = &Logical{
			Left: expr,
			Operator: *operator,
			Right: right,
		}
	}
	return expr, nil
}

func (p *Parser) parseEquality() (Expr, error) {
		expr, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		for p.match(EQUAL_EQUAL, BANG_EQUAL) {
			operator := p.peekPrevious()
			right, err := p.parseComparison()
			if err != nil {
				return nil, err
			}
			expr = &BinaryExp{
				Left: expr,
				Right: right,
				Operator: *operator,
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
		operator := p.peekPrevious()
		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		expr = &BinaryExp{
			Left: expr,
			Right: right,
			Operator: *operator,
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
		operator := p.peekPrevious()
		right, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		expr = &BinaryExp{
			Left: expr,
			Right: right,
			Operator: *operator,
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
		operator := p.peekPrevious()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		expr = &BinaryExp{
			Left: expr,
			Right: right,
			Operator: *operator,
		}
	}
	return expr, nil
}

func (p *Parser) parseUnary() (Expr, error) {
	if p.match(BANG, MINUS) {
		operator := p.peekPrevious()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &Unary{
			Operator: *operator,
			Expr: right,
		}, nil
	}
	return p.parsePrimary()
}

func (p *Parser) parsePrimary() (Expr, error) {
	if p.match(NUMBER, STRING, TRUE, FALSE, NIL) {
		return &Literal{
			Value: p.peekPrevious().Literal,
		},nil	
	}
	if (p.match(IDENTIFIER)) {
		return &Variable{
			Name: *p.peekPrevious(),
		}, nil
	}
	if p.match(LEFT_PAREN) {
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		_, err = p.consume(RIGHT_PAREN, "Expect ')' after expression.")
		if err != nil {
			return nil, err
		}
		return &Grouping{
			Expr: expr,
		}, nil
	}
	return nil, p.error(p.peekCurrent(), "Expect expression.")
}

// Utility Functions

// Check if we are at the end of file
func (p *Parser) isAtEnd() bool {
	return p.peekCurrent().TokenType == EOF
}

// Return the current token without consuming it
func (p *Parser) peekCurrent() *Token {
	if p.current >= len(p.tokens) {
		// Return EOF token if we're past the end
		return &p.tokens[len(p.tokens)-1]
	}
	return &p.tokens[p.current]
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
	return p.peekCurrent().TokenType == tokenType
}

// Return the previous token
func (p *Parser) peekPrevious() *Token {
	if p.current == 0 {
		return nil
	}
	return &p.tokens[p.current - 1]
}

// Advance the current token
func (p *Parser) advance() *Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.peekPrevious()
}

// consume attempts to consume a token of the expected type.
// If the token matches, it advances and returns the token.
// If it doesn't match, it reports an error and returns nil.
func (p *Parser) consume(tokenType TokenType, message string) (*Token, error) {
	if p.check(tokenType) {
		return p.advance(), nil
	}
	return nil, p.error(p.peekPrevious(), message)
}

func (p *Parser) error(token *Token, message string) error {
	p.viri.Error(*token, message)
	return ErrParse
}

// synchronize attempts to recover from a parse error by discarding tokens
// until we reach a synchronization point (semicolon or statement start).
// This allows the parser to continue and report multiple errors.
func (p *Parser) synchronize() {
	// First, advance past the token that caused the error
	p.advance()
	
	// Discard tokens until we find a synchronization point
	for !p.isAtEnd() {
		// If we just passed a semicolon, we're synchronized
		if p.peekPrevious().TokenType == SEMICOLON {
			return
		}
		
		// If we see a statement-starting token, we're synchronized
		switch p.peekCurrent().TokenType {
		case CLASS, FOR, FUN, IF, PRINT, RETURN, VAR, WHILE:
			return
		}
		
		p.advance()
	}
}