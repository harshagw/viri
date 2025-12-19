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
	tokens            []token.Token
	current           int
	diagnosticHandler objects.DiagnosticHandler
	hadError          bool
	filePath          string
}

func NewParser(tokens []token.Token, diagnosticHandler objects.DiagnosticHandler) *Parser {
	return &Parser{
		tokens:            tokens,
		current:           0,
		diagnosticHandler: diagnosticHandler,
	}
}

func (p *Parser) SetFilePath(path string) {
	p.filePath = path
}

func (p *Parser) Parse() (*ast.Module, error) {
	var imports []*ast.ImportStmt
	var statements []ast.Stmt
	hasSeenNonImport := false

	for !p.isAtEnd() {
		if p.check(token.IMPORT) {
			if hasSeenNonImport {
				p.error(p.peekCurrent(), "Imports must appear at the top of the file, before any other statements.")
				p.advance()
				p.synchronize()
				continue
			}
			p.advance()
			stmt, err := p.parseImportStmt()
			if err == nil && stmt != nil {
				if importStmt, ok := stmt.(*ast.ImportStmt); ok {
					imports = append(imports, importStmt)
				}
			}
		} else {
			hasSeenNonImport = true
			stmt := p.parseDeclaration()
			if stmt != nil {
				statements = append(statements, stmt)
			}
		}
	}

	mod := ast.NewModule(p.filePath, imports, statements)

	if p.hadError {
		return mod, ErrParse
	}
	return mod, nil
}

func (p *Parser) parseDeclaration() ast.Stmt {
	var (
		stmt     ast.Stmt
		err      error
		exported bool
	)

	// Check for export keyword
	if p.match(token.EXPORT) {
		exported = true
	}

	if p.match(token.VAR) {
		stmt, err = p.parseVarDecl(false)
		if exported && stmt != nil {
			stmt.(*ast.VarDeclStmt).Exported = true
		}
	} else if p.match(token.CONST) {
		stmt, err = p.parseVarDecl(true)
		if exported && stmt != nil {
			stmt.(*ast.VarDeclStmt).Exported = true
		}
	} else if p.check(token.FUN) && p.checkNext(token.IDENTIFIER) {
		p.advance()
		stmt, err = p.parseFunction()
		if exported && stmt != nil {
			stmt.(*ast.FunctionStmt).Exported = true
		}
	} else if p.match(token.CLASS) {
		stmt, err = p.parseClass()
		if exported && stmt != nil {
			stmt.(*ast.ClassStmt).Exported = true
		}
	} else {
		if exported {
			p.error(p.peekPrevious(), "Only var, const, fun, and class declarations can be exported.")
			return nil
		}
		stmt, err = p.parseStmt()
	}
	if err != nil {
		p.synchronize()
		return nil
	}
	return stmt
}

func (p *Parser) parseVarDecl(isConst bool) (ast.Stmt, error) {
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
		Name:        name,
		Initializer: initializer,
		IsConst:     isConst,
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
	if p.match(token.CONTINUE) {
		return p.parseContinueStmt()
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
		Keyword: keyword,
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
			Name: superClassName,
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
		Name:       name,
		Methods:    methods,
		SuperClass: superclass,
	}, nil
}

func (p *Parser) parseFunction() (ast.Stmt, error) {
	name, err := p.consume(token.IDENTIFIER, "Expect function name.")
	if err != nil {
		return nil, err
	}
	params, body, err := p.parseFunctionBody(objects.FunctionTypeNamed)
	if err != nil {
		return nil, err
	}
	return &ast.FunctionStmt{
		Name:   name,
		Params: params,
		Body:   body,
	}, nil
}

func (p *Parser) parseFunctionBody(functionType objects.FunctionType) ([]*token.Token, *ast.BlockStmt, error) {
	if _, err := p.consume(token.LEFT_PAREN, "Expect '(' after "+functionType.String()+"."); err != nil {
		return nil, nil, err
	}
	parameters := make([]*token.Token, 0)
	if !p.check(token.RIGHT_PAREN) {
		for {
			if len(parameters) >= 255 {
				return nil, nil, p.error(p.peekPrevious(), "Can't have more than 255 parameters.")
			}

			parameter, err := p.consume(token.IDENTIFIER, "Expect parameter name.")
			if err != nil {
				return nil, nil, err
			}
			parameters = append(parameters, parameter)
			if !p.match(token.COMMA) {
				break
			}
		}
	}
	if _, err := p.consume(token.RIGHT_PAREN, "Expect ')' after parameters."); err != nil {
		return nil, nil, err
	}
	if _, err := p.consume(token.LEFT_BRACE, "Expect '{' before block."); err != nil {
		return nil, nil, err
	}
	body, err := p.parseBlockStmt()
	if err != nil {
		return nil, nil, err
	}
	return parameters, body, nil
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
		initializer, err = p.parseVarDecl(false)
		if err != nil {
			return nil, err
		}
	} else if p.match(token.CONST) {
		initializer, err = p.parseVarDecl(true)
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

	return &ast.ForStmt{
		Initializer: initializer,
		Condition:   condition,
		Increment:   increment,
		Body:        body,
	}, nil
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
		Keyword: keyword,
	}, nil
}

func (p *Parser) parseContinueStmt() (ast.Stmt, error) {
	keyword := p.peekPrevious()
	if _, err := p.consume(token.SEMICOLON, "Expect ';' after continue."); err != nil {
		return nil, err
	}
	return &ast.ContinueStmt{
		Keyword: keyword,
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

func (p *Parser) parseImportStmt() (ast.Stmt, error) {
	pathToken, err := p.consume(token.STRING, "Expect string path after 'import'.")
	if err != nil {
		return nil, err
	}

	if _, err = p.consume(token.AS, "Expect 'as' after import path."); err != nil {
		return nil, err
	}

	aliasToken, err := p.consume(token.IDENTIFIER, "Expect alias identifier after 'as'.")
	if err != nil {
		return nil, err
	}

	if _, err = p.consume(token.SEMICOLON, "Expect ';' after import statement."); err != nil {
		return nil, err
	}

	return &ast.ImportStmt{
		Path:  pathToken,
		Alias: aliasToken,
	}, nil
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

func (p *Parser) checkNext(tokenType token.Type) bool {
	if p.isAtEnd() {
		return false
	}
	if p.current+1 >= len(p.tokens) {
		return false
	}
	return p.tokens[p.current+1].Type == tokenType
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
		case token.CLASS, token.FOR, token.FUN, token.IF, token.PRINT, token.RETURN, token.VAR, token.CONST, token.WHILE:
			return
		}

		p.advance()
	}
}
