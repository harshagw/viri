package parser

import (
	"testing"

	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/token"
)

// createParserFromTokens creates a parser from tokens directly (unit test - isolated)
func createParserFromTokens(tokens []token.Token) (*Parser, *objects.DiagnosticCollector) {
	collector := &objects.DiagnosticCollector{}
	p := NewParser(tokens, collector)
	p.SetFilePath("test.viri")
	return p, collector
}

func TestParseVarDecl(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []token.Token
		wantErr bool
		checkFn func(*testing.T, *ast.Module)
	}{
		{
			name: "simple declaration",
			// var x;
			tokens: []token.Token{
				token.New(token.VAR, "var", nil, 1, nil),
				token.New(token.IDENTIFIER, "x", nil, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				stmt, ok := mod.Statements[0].(*ast.VarDeclStmt)
				if !ok {
					t.Fatalf("expected VarDeclStmt, got %T", mod.Statements[0])
				}
				if stmt.Name.Lexeme != "x" {
					t.Errorf("name = %s, want x", stmt.Name.Lexeme)
				}
				if stmt.Initializer != nil {
					t.Errorf("expected nil initializer, got %T", stmt.Initializer)
				}
			},
		},
		{
			name: "declaration with initializer",
			// var x = 42;
			tokens: []token.Token{
				token.New(token.VAR, "var", nil, 1, nil),
				token.New(token.IDENTIFIER, "x", nil, 1, nil),
				token.New(token.EQUAL, "=", nil, 1, nil),
				token.New(token.NUMBER, "42", 42.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				stmt, ok := mod.Statements[0].(*ast.VarDeclStmt)
				if !ok {
					t.Fatalf("expected VarDeclStmt, got %T", mod.Statements[0])
				}
				if stmt.Name.Lexeme != "x" {
					t.Errorf("name = %s, want x", stmt.Name.Lexeme)
				}
				lit, ok := stmt.Initializer.(*ast.LiteralExpr)
				if !ok || lit.Value != 42.0 {
					t.Errorf("value = %v, want 42.0", lit.Value)
				}
			},
		},
		{
			name: "non ending declaration",
			// var x
			tokens: []token.Token{
				token.New(token.VAR, "var", nil, 1, nil),
				token.New(token.IDENTIFIER, "x", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, collector := createParserFromTokens(tt.tokens)
			mod, err := p.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if len(collector.Errors) == 0 {
					t.Error("expected errors but got none")
				}
				return
			}
			if tt.checkFn != nil {
				tt.checkFn(t, mod)
			}
		})
	}
}

func TestParsePrintStmt(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []token.Token
		wantErr bool
		checkFn func(*testing.T, *ast.Module)
	}{
		{
			name: "simple print statement",
			// print 42;
			tokens: []token.Token{
				token.New(token.PRINT, "print", nil, 1, nil),
				token.New(token.NUMBER, "42", 42.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				stmt, ok := mod.Statements[0].(*ast.PrintStmt)
				if !ok {
					t.Fatalf("expected PrintStmt, got %T", mod.Statements[0])
				}
				lit, ok := stmt.Expr.(*ast.LiteralExpr)
				if !ok || lit.Value != 42.0 {
					t.Errorf("expression = %v, want 42.0", lit.Value)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, collector := createParserFromTokens(tt.tokens)
			mod, err := p.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if len(collector.Errors) == 0 {
					t.Error("expected errors but got none")
				}
				return
			}
			if tt.checkFn != nil {
				tt.checkFn(t, mod)
			}
		})
	}
}

func TestParseExprStmt(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []token.Token
		wantErr bool
		checkFn func(*testing.T, *ast.Module)
	}{
		{
			name: "simple expression statement",
			// x = 10;
			tokens: []token.Token{
				token.New(token.IDENTIFIER, "x", nil, 1, nil),
				token.New(token.EQUAL, "=", nil, 1, nil),
				token.New(token.NUMBER, "10", 10.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				if len(mod.Statements) != 1 {
					t.Fatalf("expected 1 statement, got %d", len(mod.Statements))
				}
				_, ok := mod.Statements[0].(*ast.ExprStmt)
				if !ok {
					t.Fatalf("expected ExprStmt, got %T", mod.Statements[0])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, collector := createParserFromTokens(tt.tokens)
			mod, err := p.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if len(collector.Errors) == 0 {
					t.Error("expected errors but got none")
				}
				return
			}
			if tt.checkFn != nil {
				tt.checkFn(t, mod)
			}
		})
	}
}

func TestParseBlockStmt(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []token.Token
		wantErr bool
		checkFn func(*testing.T, *ast.Module)
	}{
		{
			name: "simple block",
			// { var x = 1; }
			tokens: []token.Token{
				token.New(token.LEFT_BRACE, "{", nil, 1, nil),
				token.New(token.VAR, "var", nil, 1, nil),
				token.New(token.IDENTIFIER, "x", nil, 1, nil),
				token.New(token.EQUAL, "=", nil, 1, nil),
				token.New(token.NUMBER, "1", 1.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.RIGHT_BRACE, "}", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				block, ok := mod.Statements[0].(*ast.BlockStmt)
				if !ok {
					t.Fatalf("expected BlockStmt, got %T", mod.Statements[0])
				}
				if len(block.Statements) != 1 {
					t.Fatalf("expected 1 statement in block, got %d", len(block.Statements))
				}
			},
		},
		{
			name: "nested block",
			// { { print 1; } }
			tokens: []token.Token{
				token.New(token.LEFT_BRACE, "{", nil, 1, nil),
				token.New(token.LEFT_BRACE, "{", nil, 1, nil),
				token.New(token.PRINT, "print", nil, 1, nil),
				token.New(token.NUMBER, "1", 1.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.RIGHT_BRACE, "}", nil, 1, nil),
				token.New(token.RIGHT_BRACE, "}", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				outer, ok := mod.Statements[0].(*ast.BlockStmt)
				if !ok {
					t.Fatalf("expected BlockStmt, got %T", mod.Statements[0])
				}
				if len(outer.Statements) == 0 {
					t.Fatal("expected at least one statement in outer block")
				}
				inner, ok := outer.Statements[0].(*ast.BlockStmt)
				if !ok {
					t.Fatalf("expected inner BlockStmt, got %T", outer.Statements[0])
				}
				if len(inner.Statements) == 0 {
					t.Fatal("expected at least one statement in inner block")
				}
				if _, ok := inner.Statements[0].(*ast.PrintStmt); !ok {
					t.Errorf("expected PrintStmt in inner block, got %T", inner.Statements[0])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, collector := createParserFromTokens(tt.tokens)
			mod, err := p.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if len(collector.Errors) == 0 {
					t.Error("expected errors but got none")
				}
				return
			}
			if tt.checkFn != nil {
				tt.checkFn(t, mod)
			}
		})
	}
}

func TestParseIfStmt(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []token.Token
		wantErr bool
		checkFn func(*testing.T, *ast.Module)
	}{
		{
			name: "if without else",
			// if (true) print 1;
			tokens: []token.Token{
				token.New(token.IF, "if", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.TRUE, "true", true, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.PRINT, "print", nil, 1, nil),
				token.New(token.NUMBER, "1", 1.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				ifStmt, ok := mod.Statements[0].(*ast.IfStmt)
				if !ok {
					t.Fatalf("expected IfStmt, got %T", mod.Statements[0])
				}
				cond, ok := ifStmt.Condition.(*ast.LiteralExpr)
				if !ok {
					t.Fatalf("expected LiteralExpr condition, got %T", ifStmt.Condition)
				}
				if cond.Value != true {
					t.Errorf("condition value = %v, want true", cond.Value)
				}
				if _, ok := ifStmt.ThenBranch.(*ast.PrintStmt); !ok {
					t.Errorf("expected PrintStmt then branch, got %T", ifStmt.ThenBranch)
				}
				if ifStmt.ElseBranch != nil {
					t.Error("expected no else branch")
				}
			},
		},
		{
			name: "if with else",
			// if (true) print 1; else print 2;
			tokens: []token.Token{
				token.New(token.IF, "if", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.TRUE, "true", true, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.PRINT, "print", nil, 1, nil),
				token.New(token.NUMBER, "1", 1.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.ELSE, "else", nil, 1, nil),
				token.New(token.PRINT, "print", nil, 1, nil),
				token.New(token.NUMBER, "2", 2.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				ifStmt, ok := mod.Statements[0].(*ast.IfStmt)
				if !ok {
					t.Fatalf("expected IfStmt, got %T", mod.Statements[0])
				}
				if ifStmt.ElseBranch == nil {
					t.Fatal("expected else branch")
				}
				if _, ok := ifStmt.ElseBranch.(*ast.PrintStmt); !ok {
					t.Errorf("expected PrintStmt else branch, got %T", ifStmt.ElseBranch)
				}
			},
		},
		{
			name: "nested if",
			// if (true) if (false) print 1; else print 2;
			tokens: []token.Token{
				token.New(token.IF, "if", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.TRUE, "true", true, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.IF, "if", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.FALSE, "false", false, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.PRINT, "print", nil, 1, nil),
				token.New(token.NUMBER, "1", 1.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.ELSE, "else", nil, 1, nil),
				token.New(token.PRINT, "print", nil, 1, nil),
				token.New(token.NUMBER, "2", 2.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				ifStmt, ok := mod.Statements[0].(*ast.IfStmt)
				if !ok {
					t.Fatalf("expected IfStmt, got %T", mod.Statements[0])
				}
				innerIf, ok := ifStmt.ThenBranch.(*ast.IfStmt)
				if !ok {
					t.Fatalf("expected inner IfStmt, got %T", ifStmt.ThenBranch)
				}
				if innerIf.ElseBranch == nil {
					t.Error("expected inner else branch (dangling else should bind to inner if)")
				}
				cond, ok := innerIf.Condition.(*ast.LiteralExpr)
				if !ok || cond.Value != false {
					t.Errorf("inner condition = %v, want false", cond.Value)
				}
			},
		},
		{
			name: "if-else if-else",
			// if (1) print 1; else if (2) print 2; else print 3;
			tokens: []token.Token{
				token.New(token.IF, "if", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.NUMBER, "1", 1.0, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.PRINT, "print", nil, 1, nil),
				token.New(token.NUMBER, "1", 1.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.ELSE, "else", nil, 1, nil),
				token.New(token.IF, "if", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.NUMBER, "2", 2.0, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.PRINT, "print", nil, 1, nil),
				token.New(token.NUMBER, "2", 2.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.ELSE, "else", nil, 1, nil),
				token.New(token.PRINT, "print", nil, 1, nil),
				token.New(token.NUMBER, "3", 3.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				ifStmt, ok := mod.Statements[0].(*ast.IfStmt)
				if !ok {
					t.Fatalf("expected IfStmt, got %T", mod.Statements[0])
				}
				elseIf, ok := ifStmt.ElseBranch.(*ast.IfStmt)
				if !ok {
					t.Fatalf("expected else if (IfStmt), got %T", ifStmt.ElseBranch)
				}
				cond2, ok := elseIf.Condition.(*ast.LiteralExpr)
				if !ok || cond2.Value != 2.0 {
					t.Errorf("else if condition = %v, want 2.0", cond2.Value)
				}
				if elseIf.ElseBranch == nil {
					t.Error("expected final else branch")
				}
				if _, ok := elseIf.ElseBranch.(*ast.PrintStmt); !ok {
					t.Errorf("expected final PrintStmt else branch, got %T", elseIf.ElseBranch)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, collector := createParserFromTokens(tt.tokens)
			mod, err := p.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if len(collector.Errors) == 0 {
					t.Error("expected errors but got none")
				}
				return
			}
			if tt.checkFn != nil {
				tt.checkFn(t, mod)
			}
		})
	}
}

func TestParseWhileStmt(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []token.Token
		wantErr bool
		checkFn func(*testing.T, *ast.Module)
	}{
		{
			name: "simple while loop",
			// while (true) print 1;
			tokens: []token.Token{
				token.New(token.WHILE, "while", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.TRUE, "true", true, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.PRINT, "print", nil, 1, nil),
				token.New(token.NUMBER, "1", 1.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				whileStmt, ok := mod.Statements[0].(*ast.WhileStmt)
				if !ok {
					t.Fatalf("expected WhileStmt, got %T", mod.Statements[0])
				}
				cond, ok := whileStmt.Condition.(*ast.LiteralExpr)
				if !ok || cond.Value != true {
					t.Errorf("condition = %v, want true", cond.Value)
				}
				if _, ok := whileStmt.Body.(*ast.PrintStmt); !ok {
					t.Errorf("expected PrintStmt body, got %T", whileStmt.Body)
				}
			},
		},
		{
			name: "while with complex condition",
			// while (x < 10 and y == 0) print x;
			tokens: []token.Token{
				token.New(token.WHILE, "while", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.IDENTIFIER, "x", nil, 1, nil),
				token.New(token.LESS, "<", nil, 1, nil),
				token.New(token.NUMBER, "10", 10.0, 1, nil),
				token.New(token.AND, "and", nil, 1, nil),
				token.New(token.IDENTIFIER, "y", nil, 1, nil),
				token.New(token.EQUAL_EQUAL, "==", nil, 1, nil),
				token.New(token.NUMBER, "0", 0.0, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.PRINT, "print", nil, 1, nil),
				token.New(token.IDENTIFIER, "x", nil, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				whileStmt, ok := mod.Statements[0].(*ast.WhileStmt)
				if !ok {
					t.Fatalf("expected WhileStmt, got %T", mod.Statements[0])
				}
				logical, ok := whileStmt.Condition.(*ast.LogicalExpr)
				if !ok {
					t.Fatalf("expected logical condition, got %T", whileStmt.Condition)
				}
				if logical.Operator.Type != token.AND {
					t.Errorf("operator = %v, want AND", logical.Operator.Type)
				}
				if _, ok := whileStmt.Body.(*ast.PrintStmt); !ok {
					t.Errorf("expected PrintStmt body, got %T", whileStmt.Body)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, collector := createParserFromTokens(tt.tokens)
			mod, err := p.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if len(collector.Errors) == 0 {
					t.Error("expected errors but got none")
				}
				return
			}
			if tt.checkFn != nil {
				tt.checkFn(t, mod)
			}
		})
	}
}

func TestParseForStmt(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []token.Token
		wantErr bool
		checkFn func(*testing.T, *ast.Module)
	}{
		{
			name: "simple for loop",
			// for (;;) print 1;
			tokens: []token.Token{
				token.New(token.FOR, "for", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.PRINT, "print", nil, 1, nil),
				token.New(token.NUMBER, "1", 1.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				forStmt, ok := mod.Statements[0].(*ast.ForStmt)
				if !ok {
					t.Fatalf("expected ForStmt, got %T", mod.Statements[0])
				}
				if forStmt.Initializer != nil {
					t.Errorf("expected nil initializer, got %T", forStmt.Initializer)
				}
				if forStmt.Condition != nil {
					t.Errorf("expected nil condition, got %T", forStmt.Condition)
				}
				if forStmt.Increment != nil {
					t.Errorf("expected nil increment, got %T", forStmt.Increment)
				}
				if _, ok := forStmt.Body.(*ast.PrintStmt); !ok {
					t.Errorf("expected PrintStmt body, got %T", forStmt.Body)
				}
			},
		},
		{
			name: "full for loop",
			// for (var i = 0; i < 10; i = i + 1) print i;
			tokens: []token.Token{
				token.New(token.FOR, "for", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.VAR, "var", nil, 1, nil),
				token.New(token.IDENTIFIER, "i", nil, 1, nil),
				token.New(token.EQUAL, "=", nil, 1, nil),
				token.New(token.NUMBER, "0", 0.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.IDENTIFIER, "i", nil, 1, nil),
				token.New(token.LESS, "<", nil, 1, nil),
				token.New(token.NUMBER, "10", 10.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.IDENTIFIER, "i", nil, 1, nil),
				token.New(token.EQUAL, "=", nil, 1, nil),
				token.New(token.IDENTIFIER, "i", nil, 1, nil),
				token.New(token.PLUS, "+", nil, 1, nil),
				token.New(token.NUMBER, "1", 1.0, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.PRINT, "print", nil, 1, nil),
				token.New(token.IDENTIFIER, "i", nil, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				forStmt, ok := mod.Statements[0].(*ast.ForStmt)
				if !ok {
					t.Fatalf("expected ForStmt, got %T", mod.Statements[0])
				}
				if _, ok := forStmt.Initializer.(*ast.VarDeclStmt); !ok {
					t.Errorf("expected VarDeclStmt initializer, got %T", forStmt.Initializer)
				}
				if _, ok := forStmt.Condition.(*ast.BinaryExpr); !ok {
					t.Errorf("expected BinaryExpr condition, got %T", forStmt.Condition)
				}
				if _, ok := forStmt.Increment.(*ast.AssignExpr); !ok {
					t.Errorf("expected AssignExpr increment, got %T", forStmt.Increment)
				}
			},
		},
		{
			name: "for without init",
			// for (; i < 10; i = i + 1) print i;
			tokens: []token.Token{
				token.New(token.FOR, "for", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.IDENTIFIER, "i", nil, 1, nil),
				token.New(token.LESS, "<", nil, 1, nil),
				token.New(token.NUMBER, "10", 10.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.IDENTIFIER, "i", nil, 1, nil),
				token.New(token.EQUAL, "=", nil, 1, nil),
				token.New(token.IDENTIFIER, "i", nil, 1, nil),
				token.New(token.PLUS, "+", nil, 1, nil),
				token.New(token.NUMBER, "1", 1.0, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.PRINT, "print", nil, 1, nil),
				token.New(token.IDENTIFIER, "i", nil, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				forStmt, ok := mod.Statements[0].(*ast.ForStmt)
				if !ok {
					t.Fatalf("expected ForStmt, got %T", mod.Statements[0])
				}
				if forStmt.Initializer != nil {
					t.Errorf("expected nil initializer, got %T", forStmt.Initializer)
				}
				if forStmt.Condition == nil {
					t.Error("expected condition")
				}
				if forStmt.Increment == nil {
					t.Error("expected increment")
				}
			},
		},
		{
			name: "for without condition",
			// for (var i = 0; ; i = i + 1) print i;
			tokens: []token.Token{
				token.New(token.FOR, "for", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.VAR, "var", nil, 1, nil),
				token.New(token.IDENTIFIER, "i", nil, 1, nil),
				token.New(token.EQUAL, "=", nil, 1, nil),
				token.New(token.NUMBER, "0", 0.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.IDENTIFIER, "i", nil, 1, nil),
				token.New(token.EQUAL, "=", nil, 1, nil),
				token.New(token.IDENTIFIER, "i", nil, 1, nil),
				token.New(token.PLUS, "+", nil, 1, nil),
				token.New(token.NUMBER, "1", 1.0, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.PRINT, "print", nil, 1, nil),
				token.New(token.IDENTIFIER, "i", nil, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				forStmt, ok := mod.Statements[0].(*ast.ForStmt)
				if !ok {
					t.Fatalf("expected ForStmt, got %T", mod.Statements[0])
				}
				if forStmt.Initializer == nil {
					t.Error("expected initializer")
				}
				if forStmt.Condition != nil {
					t.Errorf("expected nil condition, got %T", forStmt.Condition)
				}
				if forStmt.Increment == nil {
					t.Error("expected increment")
				}
			},
		},
		{
			name: "for without increment",
			// for (var i = 0; i < 10; ) print i;
			tokens: []token.Token{
				token.New(token.FOR, "for", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.VAR, "var", nil, 1, nil),
				token.New(token.IDENTIFIER, "i", nil, 1, nil),
				token.New(token.EQUAL, "=", nil, 1, nil),
				token.New(token.NUMBER, "0", 0.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.IDENTIFIER, "i", nil, 1, nil),
				token.New(token.LESS, "<", nil, 1, nil),
				token.New(token.NUMBER, "10", 10.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.PRINT, "print", nil, 1, nil),
				token.New(token.IDENTIFIER, "i", nil, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				forStmt, ok := mod.Statements[0].(*ast.ForStmt)
				if !ok {
					t.Fatalf("expected ForStmt, got %T", mod.Statements[0])
				}
				if forStmt.Initializer == nil {
					t.Error("expected initializer")
				}
				if forStmt.Condition == nil {
					t.Error("expected condition")
				}
				if forStmt.Increment != nil {
					t.Errorf("expected nil increment, got %T", forStmt.Increment)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, collector := createParserFromTokens(tt.tokens)
			mod, err := p.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if len(collector.Errors) == 0 {
					t.Error("expected errors but got none")
				}
				return
			}
			if tt.checkFn != nil {
				tt.checkFn(t, mod)
			}
		})
	}
}

func TestParseBreakStmt(t *testing.T) {
	// break;
	tokens := []token.Token{
		token.New(token.BREAK, "break", nil, 1, nil),
		token.New(token.SEMICOLON, ";", nil, 1, nil),
		token.New(token.EOF, "", nil, 1, nil),
	}
	p, _ := createParserFromTokens(tokens)
	mod, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if _, ok := mod.Statements[0].(*ast.BreakStmt); !ok {
		t.Fatalf("expected BreakStmt, got %T", mod.Statements[0])
	}
}

func TestParseContinueStmt(t *testing.T) {
	// continue;
	tokens := []token.Token{
		token.New(token.CONTINUE, "continue", nil, 1, nil),
		token.New(token.SEMICOLON, ";", nil, 1, nil),
		token.New(token.EOF, "", nil, 1, nil),
	}
	p, _ := createParserFromTokens(tokens)
	mod, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if _, ok := mod.Statements[0].(*ast.ContinueStmt); !ok {
		t.Fatalf("expected ContinueStmt, got %T", mod.Statements[0])
	}
}

func TestParseReturnStmt(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []token.Token
		wantErr bool
		checkFn func(*testing.T, *ast.Module)
	}{
		{
			name: "return with value",
			// return 42;
			tokens: []token.Token{
				token.New(token.RETURN, "return", nil, 1, nil),
				token.New(token.NUMBER, "42", 42.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				stmt, ok := mod.Statements[0].(*ast.ReturnStmt)
				if !ok {
					t.Fatalf("expected ReturnStmt, got %T", mod.Statements[0])
				}
				lit, ok := stmt.Value.(*ast.LiteralExpr)
				if !ok || lit.Value != 42.0 {
					t.Errorf("return value = %v, want 42.0", lit.Value)
				}
			},
		},
		{
			name: "return without value",
			// return;
			tokens: []token.Token{
				token.New(token.RETURN, "return", nil, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				stmt, ok := mod.Statements[0].(*ast.ReturnStmt)
				if !ok {
					t.Fatalf("expected ReturnStmt, got %T", mod.Statements[0])
				}
				if stmt.Value != nil {
					t.Errorf("expected nil return value, got %T", stmt.Value)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, collector := createParserFromTokens(tt.tokens)
			mod, err := p.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if len(collector.Errors) == 0 {
					t.Error("expected errors but got none")
				}
				return
			}
			if tt.checkFn != nil {
				tt.checkFn(t, mod)
			}
		})
	}
}

func TestParseFunction(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []token.Token
		wantErr bool
		checkFn func(*testing.T, *ast.Module)
	}{
		{
			name: "simple function",
			// fun foo() {}
			tokens: []token.Token{
				token.New(token.FUN, "fun", nil, 1, nil),
				token.New(token.IDENTIFIER, "foo", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.LEFT_BRACE, "{", nil, 1, nil),
				token.New(token.RIGHT_BRACE, "}", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				funStmt, ok := mod.Statements[0].(*ast.FunctionStmt)
				if !ok {
					t.Fatalf("expected FunctionStmt, got %T", mod.Statements[0])
				}
				if funStmt.Name.Lexeme != "foo" {
					t.Errorf("name = %s, want foo", funStmt.Name.Lexeme)
				}
				if len(funStmt.Params) != 0 {
					t.Errorf("params = %d, want 0", len(funStmt.Params))
				}
				if funStmt.Body == nil || len(funStmt.Body.Statements) != 0 {
					t.Error("expected empty body")
				}
			},
		},
		{
			name: "function with parameters",
			// fun add(a, b) { return a + b; }
			tokens: []token.Token{
				token.New(token.FUN, "fun", nil, 1, nil),
				token.New(token.IDENTIFIER, "add", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.IDENTIFIER, "a", nil, 1, nil),
				token.New(token.COMMA, ",", nil, 1, nil),
				token.New(token.IDENTIFIER, "b", nil, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.LEFT_BRACE, "{", nil, 1, nil),
				token.New(token.RETURN, "return", nil, 1, nil),
				token.New(token.IDENTIFIER, "a", nil, 1, nil),
				token.New(token.PLUS, "+", nil, 1, nil),
				token.New(token.IDENTIFIER, "b", nil, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.RIGHT_BRACE, "}", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				funStmt, ok := mod.Statements[0].(*ast.FunctionStmt)
				if !ok {
					t.Fatalf("expected FunctionStmt, got %T", mod.Statements[0])
				}
				if len(funStmt.Params) != 2 {
					t.Errorf("params = %d, want 2", len(funStmt.Params))
				}
				if funStmt.Params[0].Lexeme != "a" || funStmt.Params[1].Lexeme != "b" {
					t.Errorf("params = %v, want [a, b]", funStmt.Params)
				}
				if len(funStmt.Body.Statements) != 1 {
					t.Errorf("body statements = %d, want 1", len(funStmt.Body.Statements))
				}
				ret, ok := funStmt.Body.Statements[0].(*ast.ReturnStmt)
				if !ok {
					t.Fatalf("expected ReturnStmt, got %T", funStmt.Body.Statements[0])
				}
				if _, ok := ret.Value.(*ast.BinaryExpr); !ok {
					t.Errorf("expected BinaryExpr return value, got %T", ret.Value)
				}
			},
		},
		{
			name: "function with multiple statements",
			// fun foo() { print 1; print 2; }
			tokens: []token.Token{
				token.New(token.FUN, "fun", nil, 1, nil),
				token.New(token.IDENTIFIER, "foo", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.LEFT_BRACE, "{", nil, 1, nil),
				token.New(token.PRINT, "print", nil, 1, nil),
				token.New(token.NUMBER, "1", 1.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.PRINT, "print", nil, 1, nil),
				token.New(token.NUMBER, "2", 2.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.RIGHT_BRACE, "}", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				funStmt, ok := mod.Statements[0].(*ast.FunctionStmt)
				if !ok {
					t.Fatalf("expected FunctionStmt, got %T", mod.Statements[0])
				}
				if len(funStmt.Body.Statements) != 2 {
					t.Errorf("body statements = %d, want 2", len(funStmt.Body.Statements))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, collector := createParserFromTokens(tt.tokens)
			mod, err := p.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if len(collector.Errors) == 0 {
					t.Error("expected errors but got none")
				}
				return
			}
			if tt.checkFn != nil {
				tt.checkFn(t, mod)
			}
		})
	}
}

func TestParseFunctionExpr(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []token.Token
		wantErr bool
		checkFn func(*testing.T, *ast.Module)
	}{
		{
			name: "anonymous function",
			// var a = fun() { return 0; };
			tokens: []token.Token{
				token.New(token.VAR, "var", nil, 1, nil),
				token.New(token.IDENTIFIER, "a", nil, 1, nil),
				token.New(token.EQUAL, "=", nil, 1, nil),
				token.New(token.FUN, "fun", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.LEFT_BRACE, "{", nil, 1, nil),
				token.New(token.RETURN, "return", nil, 1, nil),
				token.New(token.NUMBER, "0", 0.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.RIGHT_BRACE, "}", nil, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				stmt, ok := mod.Statements[0].(*ast.VarDeclStmt)
				if !ok {
					t.Fatalf("expected VarDeclStmt, got %T", mod.Statements[0])
				}
				expr, ok := stmt.Initializer.(*ast.FunctionExpr)
				if !ok {
					t.Fatalf("expected FunctionExpr, got %T", stmt.Initializer)
				}
				if len(expr.Params) != 0 {
					t.Errorf("params = %d, want 0", len(expr.Params))
				}
				if len(expr.Body.Statements) != 1 {
					t.Errorf("body statements = %d, want 1", len(expr.Body.Statements))
				}
			},
		},
		{
			name: "function with parameters",
			// var add = fun(a, b) { return a + b; };
			tokens: []token.Token{
				token.New(token.VAR, "var", nil, 1, nil),
				token.New(token.IDENTIFIER, "add", nil, 1, nil),
				token.New(token.EQUAL, "=", nil, 1, nil),
				token.New(token.FUN, "fun", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.IDENTIFIER, "a", nil, 1, nil),
				token.New(token.COMMA, ",", nil, 1, nil),
				token.New(token.IDENTIFIER, "b", nil, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.LEFT_BRACE, "{", nil, 1, nil),
				token.New(token.RETURN, "return", nil, 1, nil),
				token.New(token.IDENTIFIER, "a", nil, 1, nil),
				token.New(token.PLUS, "+", nil, 1, nil),
				token.New(token.IDENTIFIER, "b", nil, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.RIGHT_BRACE, "}", nil, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				stmt := mod.Statements[0].(*ast.VarDeclStmt)
				expr := stmt.Initializer.(*ast.FunctionExpr)
				if len(expr.Params) != 2 {
					t.Errorf("params = %d, want 2", len(expr.Params))
				}
			},
		},
		{
			name: "IIFE",
			// fun(x) { print x; }(10);
			tokens: []token.Token{
				token.New(token.FUN, "fun", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.IDENTIFIER, "x", nil, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.LEFT_BRACE, "{", nil, 1, nil),
				token.New(token.PRINT, "print", nil, 1, nil),
				token.New(token.IDENTIFIER, "x", nil, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.RIGHT_BRACE, "}", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.NUMBER, "10", 10.0, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				stmt, ok := mod.Statements[0].(*ast.ExprStmt)
				if !ok {
					t.Fatalf("expected ExprStmt, got %T", mod.Statements[0])
				}
				call, ok := stmt.Expr.(*ast.CallExpr)
				if !ok {
					t.Fatalf("expected CallExpr, got %T", stmt.Expr)
				}
				if _, ok := call.Callee.(*ast.FunctionExpr); !ok {
					t.Errorf("expected Callee to be FunctionExpr, got %T", call.Callee)
				}
			},
		},
		{
			name: "function with multiple statements",
			// var f = fun() { print 1; print 2; };
			tokens: []token.Token{
				token.New(token.VAR, "var", nil, 1, nil),
				token.New(token.IDENTIFIER, "f", nil, 1, nil),
				token.New(token.EQUAL, "=", nil, 1, nil),
				token.New(token.FUN, "fun", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.LEFT_BRACE, "{", nil, 1, nil),
				token.New(token.PRINT, "print", nil, 1, nil),
				token.New(token.NUMBER, "1", 1.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.PRINT, "print", nil, 1, nil),
				token.New(token.NUMBER, "2", 2.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.RIGHT_BRACE, "}", nil, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				stmt := mod.Statements[0].(*ast.VarDeclStmt)
				expr := stmt.Initializer.(*ast.FunctionExpr)
				if len(expr.Body.Statements) != 2 {
					t.Errorf("body statements = %d, want 2", len(expr.Body.Statements))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _ := createParserFromTokens(tt.tokens)
			mod, err := p.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.checkFn != nil {
				tt.checkFn(t, mod)
			}
		})
	}
}

func TestParseClass(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []token.Token
		wantErr bool
		checkFn func(*testing.T, *ast.Module)
	}{
		{
			name: "simple class",
			// class Foo {}
			tokens: []token.Token{
				token.New(token.CLASS, "class", nil, 1, nil),
				token.New(token.IDENTIFIER, "Foo", nil, 1, nil),
				token.New(token.LEFT_BRACE, "{", nil, 1, nil),
				token.New(token.RIGHT_BRACE, "}", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				classStmt, ok := mod.Statements[0].(*ast.ClassStmt)
				if !ok {
					t.Fatalf("expected ClassStmt, got %T", mod.Statements[0])
				}
				if classStmt.Name.Lexeme != "Foo" {
					t.Errorf("name = %s, want Foo", classStmt.Name.Lexeme)
				}
				if len(classStmt.Methods) != 0 {
					t.Errorf("methods = %d, want 0", len(classStmt.Methods))
				}
				if classStmt.SuperClass != nil {
					t.Error("expected no super class")
				}
			},
		},
		{
			name: "class with methods",
			// class Foo { bar() {} baz(x) { print x; } }
			tokens: []token.Token{
				token.New(token.CLASS, "class", nil, 1, nil),
				token.New(token.IDENTIFIER, "Foo", nil, 1, nil),
				token.New(token.LEFT_BRACE, "{", nil, 1, nil),
				token.New(token.IDENTIFIER, "bar", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.LEFT_BRACE, "{", nil, 1, nil),
				token.New(token.RIGHT_BRACE, "}", nil, 1, nil),
				token.New(token.IDENTIFIER, "baz", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.IDENTIFIER, "x", nil, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.LEFT_BRACE, "{", nil, 1, nil),
				token.New(token.PRINT, "print", nil, 1, nil),
				token.New(token.IDENTIFIER, "x", nil, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.RIGHT_BRACE, "}", nil, 1, nil),
				token.New(token.RIGHT_BRACE, "}", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				classStmt, ok := mod.Statements[0].(*ast.ClassStmt)
				if !ok {
					t.Fatalf("expected ClassStmt, got %T", mod.Statements[0])
				}
				if len(classStmt.Methods) != 2 {
					t.Errorf("methods = %d, want 2", len(classStmt.Methods))
				}
				if classStmt.Methods[0].Name.Lexeme != "bar" {
					t.Errorf("method 0 name = %s, want bar", classStmt.Methods[0].Name.Lexeme)
				}
				if classStmt.Methods[1].Name.Lexeme != "baz" {
					t.Errorf("method 1 name = %s, want baz", classStmt.Methods[1].Name.Lexeme)
				}
				if len(classStmt.Methods[1].Params) != 1 {
					t.Errorf("method 1 params = %d, want 1", len(classStmt.Methods[1].Params))
				}
			},
		},
		{
			name: "class with inheritance",
			// class B < A {}
			tokens: []token.Token{
				token.New(token.CLASS, "class", nil, 1, nil),
				token.New(token.IDENTIFIER, "B", nil, 1, nil),
				token.New(token.LESS, "<", nil, 1, nil),
				token.New(token.IDENTIFIER, "A", nil, 1, nil),
				token.New(token.LEFT_BRACE, "{", nil, 1, nil),
				token.New(token.RIGHT_BRACE, "}", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			wantErr: false,
			checkFn: func(t *testing.T, mod *ast.Module) {
				classStmt, ok := mod.Statements[0].(*ast.ClassStmt)
				if !ok {
					t.Fatalf("expected ClassStmt, got %T", mod.Statements[0])
				}
				if classStmt.SuperClass == nil {
					t.Fatal("expected super class")
				}
				if classStmt.SuperClass.Name.Lexeme != "A" {
					t.Errorf("super class name = %s, want A", classStmt.SuperClass.Name.Lexeme)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, collector := createParserFromTokens(tt.tokens)
			mod, err := p.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if len(collector.Errors) == 0 {
					t.Error("expected errors but got none")
				}
				return
			}
			if tt.checkFn != nil {
				tt.checkFn(t, mod)
			}
		})
	}
}

func TestParseImportStmt(t *testing.T) {
	// import "path" as mod;
	tokens := []token.Token{
		token.New(token.IMPORT, "import", nil, 1, nil),
		token.New(token.STRING, `"path"`, "path", 1, nil),
		token.New(token.AS, "as", nil, 1, nil),
		token.New(token.IDENTIFIER, "mod", nil, 1, nil),
		token.New(token.SEMICOLON, ";", nil, 1, nil),
		token.New(token.EOF, "", nil, 1, nil),
	}
	p, _ := createParserFromTokens(tokens)
	mod, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(mod.Imports) != 1 {
		t.Fatalf("expected 1 import, got %d", len(mod.Imports))
	}
}

func TestParseMultipleStatements(t *testing.T) {
	// var x = 1; var y = 2;
	tokens := []token.Token{
		token.New(token.VAR, "var", nil, 1, nil),
		token.New(token.IDENTIFIER, "x", nil, 1, nil),
		token.New(token.EQUAL, "=", nil, 1, nil),
		token.New(token.NUMBER, "1", 1.0, 1, nil),
		token.New(token.SEMICOLON, ";", nil, 1, nil),
		token.New(token.VAR, "var", nil, 1, nil),
		token.New(token.IDENTIFIER, "y", nil, 1, nil),
		token.New(token.EQUAL, "=", nil, 1, nil),
		token.New(token.NUMBER, "2", 2.0, 1, nil),
		token.New(token.SEMICOLON, ";", nil, 1, nil),
		token.New(token.EOF, "", nil, 1, nil),
	}
	p, _ := createParserFromTokens(tokens)
	mod, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(mod.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(mod.Statements))
	}
}

func TestParseInvalidSyntax(t *testing.T) {
	tests := []struct {
		name   string
		tokens []token.Token
	}{
		{
			"missing semicolon after print",
			// print 1
			[]token.Token{
				token.New(token.PRINT, "print", nil, 1, nil),
				token.New(token.NUMBER, "1", 1.0, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
		},
		{
			"missing closing brace in block",
			// { var x = 1;
			[]token.Token{
				token.New(token.LEFT_BRACE, "{", nil, 1, nil),
				token.New(token.VAR, "var", nil, 1, nil),
				token.New(token.IDENTIFIER, "x", nil, 1, nil),
				token.New(token.EQUAL, "=", nil, 1, nil),
				token.New(token.NUMBER, "1", 1.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, collector := createParserFromTokens(tt.tokens)
			_, _ = p.Parse()
			if len(collector.Errors) == 0 {
				t.Error("expected diagnostic errors but got none")
			}
		})
	}
}
