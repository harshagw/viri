package parser

import (
	"testing"

	"reflect"

	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/token"
)

func parseExpressionFromTokens(tokens []token.Token) (ast.Expr, *objects.DiagnosticCollector, error) {
	collector := &objects.DiagnosticCollector{}
	p := NewParser(tokens, collector)
	p.SetFilePath("test.viri")
	expr, err := p.parseExpr()
	return expr, collector, err
}

func TestParseLiteralExpressions(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []token.Token
		expected interface{}
	}{
		{
			"number",
			[]token.Token{
				token.New(token.NUMBER, "42", 42.0, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			42.0,
		},
		{
			"float",
			[]token.Token{
				token.New(token.NUMBER, "3.14", 3.14, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			3.14,
		},
		{
			"string",
			[]token.Token{
				token.New(token.STRING, `"hello"`, "hello", 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			"hello",
		},
		{
			"true",
			[]token.Token{
				token.New(token.TRUE, "true", true, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			true,
		},
		{
			"false",
			[]token.Token{
				token.New(token.FALSE, "false", false, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			false,
		},
		{
			"nil",
			[]token.Token{
				token.New(token.NIL, "nil", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, _, err := parseExpressionFromTokens(tt.tokens)
			if err != nil {
				t.Fatalf("parseExpr() error = %v", err)
			}
			assertLiteral(t, expr, tt.expected)
		})
	}
}

func TestParseVariableExpression(t *testing.T) {
	tokens := []token.Token{
		token.New(token.IDENTIFIER, "x", nil, 1, nil),
		token.New(token.EOF, "", nil, 1, nil),
	}
	expr, _, err := parseExpressionFromTokens(tokens)
	if err != nil {
		t.Fatalf("parseExpr() error = %v", err)
	}
	assertVariable(t, expr, "x")
}

func TestParseUnaryExpressions(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []token.Token
		operator token.Type
		operand  any
	}{
		{
			"negation",
			[]token.Token{
				token.New(token.MINUS, "-", nil, 1, nil),
				token.New(token.NUMBER, "42", 42.0, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			token.MINUS,
			42.0,
		},
		{
			"logical not",
			[]token.Token{
				token.New(token.BANG, "!", nil, 1, nil),
				token.New(token.TRUE, "true", true, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			token.BANG,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, _, err := parseExpressionFromTokens(tt.tokens)
			if err != nil {
				t.Fatalf("parseExpr() error = %v", err)
			}
			_, operand := assertUnary(t, expr, tt.operator)
			assertLiteral(t, operand, tt.operand)
		})
	}
}

func TestParseBinaryExpressions(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []token.Token
		left     any
		operator token.Type
		right    any
	}{
		{
			"addition",
			[]token.Token{
				token.New(token.NUMBER, "1", 1.0, 1, nil),
				token.New(token.PLUS, "+", nil, 1, nil),
				token.New(token.NUMBER, "2", 2.0, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			1.0,
			token.PLUS,
			2.0,
		},
		{
			"equality",
			[]token.Token{
				token.New(token.IDENTIFIER, "x", nil, 1, nil),
				token.New(token.EQUAL_EQUAL, "==", nil, 1, nil),
				token.New(token.IDENTIFIER, "y", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			"x",
			token.EQUAL_EQUAL,
			"y",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, _, err := parseExpressionFromTokens(tt.tokens)
			if err != nil {
				t.Fatalf("parseExpr() error = %v", err)
			}
			_, left, right := assertBinary(t, expr, tt.operator)

			if s, ok := tt.left.(string); ok {
				assertVariable(t, left, s)
			} else {
				assertLiteral(t, left, tt.left)
			}

			if s, ok := tt.right.(string); ok {
				assertVariable(t, right, s)
			} else {
				assertLiteral(t, right, tt.right)
			}
		})
	}
}

func TestParsePrecedence(t *testing.T) {
	t.Run("multiplication before addition", func(t *testing.T) {
		// 1 + 2 * 3
		tokens := []token.Token{
			token.New(token.NUMBER, "1", 1.0, 1, nil),
			token.New(token.PLUS, "+", nil, 1, nil),
			token.New(token.NUMBER, "2", 2.0, 1, nil),
			token.New(token.STAR, "*", nil, 1, nil),
			token.New(token.NUMBER, "3", 3.0, 1, nil),
			token.New(token.EOF, "", nil, 1, nil),
		}
		expr, _, err := parseExpressionFromTokens(tokens)
		if err != nil {
			t.Fatalf("parseExpr() error = %v", err)
		}

		// (1 + (2 * 3))
		_, left, right := assertBinary(t, expr, token.PLUS)
		assertLiteral(t, left, 1.0)
		_, rLeft, rRight := assertBinary(t, right, token.STAR)
		assertLiteral(t, rLeft, 2.0)
		assertLiteral(t, rRight, 3.0)
	})

	t.Run("parentheses override precedence", func(t *testing.T) {
		// (1 + 2) * 3
		tokens := []token.Token{
			token.New(token.LEFT_PAREN, "(", nil, 1, nil),
			token.New(token.NUMBER, "1", 1.0, 1, nil),
			token.New(token.PLUS, "+", nil, 1, nil),
			token.New(token.NUMBER, "2", 2.0, 1, nil),
			token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
			token.New(token.STAR, "*", nil, 1, nil),
			token.New(token.NUMBER, "3", 3.0, 1, nil),
			token.New(token.EOF, "", nil, 1, nil),
		}
		expr, _, err := parseExpressionFromTokens(tokens)
		if err != nil {
			t.Fatalf("parseExpr() error = %v", err)
		}

		// ((1 + 2) * 3)
		_, left, right := assertBinary(t, expr, token.STAR)
		inner := assertGrouping(t, left)
		_, iLeft, iRight := assertBinary(t, inner, token.PLUS)
		assertLiteral(t, iLeft, 1.0)
		assertLiteral(t, iRight, 2.0)
		assertLiteral(t, right, 3.0)
	})
}

func TestParseLogicalExpressions(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []token.Token
		left     any
		operator token.Type
		right    any
	}{
		{
			"and",
			[]token.Token{
				token.New(token.TRUE, "true", true, 1, nil),
				token.New(token.AND, "and", nil, 1, nil),
				token.New(token.FALSE, "false", false, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			true,
			token.AND,
			false,
		},
		{
			"or",
			[]token.Token{
				token.New(token.TRUE, "true", true, 1, nil),
				token.New(token.OR, "or", nil, 1, nil),
				token.New(token.FALSE, "false", false, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			true,
			token.OR,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, _, err := parseExpressionFromTokens(tt.tokens)
			if err != nil {
				t.Fatalf("parseExpr() error = %v", err)
			}
			_, left, right := assertLogical(t, expr, tt.operator)
			assertLiteral(t, left, tt.left)
			assertLiteral(t, right, tt.right)
		})
	}
}

func TestParseAssignmentExpression(t *testing.T) {
	// x = 42
	tokens := []token.Token{
		token.New(token.IDENTIFIER, "x", nil, 1, nil),
		token.New(token.EQUAL, "=", nil, 1, nil),
		token.New(token.NUMBER, "42", 42.0, 1, nil),
		token.New(token.EOF, "", nil, 1, nil),
	}
	expr, _, err := parseExpressionFromTokens(tokens)
	if err != nil {
		t.Fatalf("parseExpr() error = %v", err)
	}
	_, value := assertAssignment(t, expr, "x")
	assertLiteral(t, value, 42.0)
}

func TestParseCallExpression(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []token.Token
		argCount int
	}{
		{
			"no args",
			[]token.Token{
				token.New(token.IDENTIFIER, "foo", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			0,
		},
		{
			"one arg",
			[]token.Token{
				token.New(token.IDENTIFIER, "foo", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.NUMBER, "1", 1.0, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, _, err := parseExpressionFromTokens(tt.tokens)
			if err != nil {
				t.Fatalf("parseExpr() error = %v", err)
			}
			_, callee, args := assertCall(t, expr, tt.argCount)
			assertVariable(t, callee, "foo")
			if tt.argCount > 0 {
				assertLiteral(t, args[0], 1.0)
			}
		})
	}
}

func TestParseGetExpression(t *testing.T) {
	// obj.property
	tokens := []token.Token{
		token.New(token.IDENTIFIER, "obj", nil, 1, nil),
		token.New(token.DOT, ".", nil, 1, nil),
		token.New(token.IDENTIFIER, "property", nil, 1, nil),
		token.New(token.EOF, "", nil, 1, nil),
	}
	expr, _, err := parseExpressionFromTokens(tokens)
	if err != nil {
		t.Fatalf("parseExpr() error = %v", err)
	}
	_, obj := assertGet(t, expr, "property")
	assertVariable(t, obj, "obj")
}

func TestParseSetExpression(t *testing.T) {
	// obj.property = value
	tokens := []token.Token{
		token.New(token.IDENTIFIER, "obj", nil, 1, nil),
		token.New(token.DOT, ".", nil, 1, nil),
		token.New(token.IDENTIFIER, "property", nil, 1, nil),
		token.New(token.EQUAL, "=", nil, 1, nil),
		token.New(token.IDENTIFIER, "value", nil, 1, nil),
		token.New(token.EOF, "", nil, 1, nil),
	}
	expr, _, err := parseExpressionFromTokens(tokens)
	if err != nil {
		t.Fatalf("parseExpr() error = %v", err)
	}
	_, obj, value := assertSet(t, expr, "property")
	assertVariable(t, obj, "obj")
	assertVariable(t, value, "value")
}

func TestParseThisExpression(t *testing.T) {
	// this
	tokens := []token.Token{
		token.New(token.THIS, "this", nil, 1, nil),
		token.New(token.EOF, "", nil, 1, nil),
	}
	expr, _, err := parseExpressionFromTokens(tokens)
	if err != nil {
		t.Fatalf("parseExpr() error = %v", err)
	}
	assertThis(t, expr)
}

func TestParseSuperExpression(t *testing.T) {
	// super.method
	tokens := []token.Token{
		token.New(token.SUPER, "super", nil, 1, nil),
		token.New(token.DOT, ".", nil, 1, nil),
		token.New(token.IDENTIFIER, "method", nil, 1, nil),
		token.New(token.EOF, "", nil, 1, nil),
	}
	expr, _, err := parseExpressionFromTokens(tokens)
	if err != nil {
		t.Fatalf("parseExpr() error = %v", err)
	}
	assertSuper(t, expr, "method")
}

func TestParseArrayLiteral(t *testing.T) {
	tests := []struct {
		name   string
		tokens []token.Token
		count  int
	}{
		{
			"empty array",
			[]token.Token{
				token.New(token.LEFT_BRACKET, "[", nil, 1, nil),
				token.New(token.RIGHT_BRACKET, "]", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			0,
		},
		{
			"single element",
			[]token.Token{
				token.New(token.LEFT_BRACKET, "[", nil, 1, nil),
				token.New(token.NUMBER, "1", 1.0, 1, nil),
				token.New(token.RIGHT_BRACKET, "]", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, _, err := parseExpressionFromTokens(tt.tokens)
			if err != nil {
				t.Fatalf("parseExpr() error = %v", err)
			}
			_, elements := assertArray(t, expr, tt.count)
			if tt.count > 0 {
				assertLiteral(t, elements[0], 1.0)
			}
		})
	}
}

func TestParseHashLiteral(t *testing.T) {
	tests := []struct {
		name   string
		tokens []token.Token
		count  int
	}{
		{
			"empty hash",
			[]token.Token{
				token.New(token.LEFT_BRACE, "{", nil, 1, nil),
				token.New(token.RIGHT_BRACE, "}", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			0,
		},
		{
			"single pair",
			[]token.Token{
				token.New(token.LEFT_BRACE, "{", nil, 1, nil),
				token.New(token.STRING, `"key"`, "key", 1, nil),
				token.New(token.COLON, ":", nil, 1, nil),
				token.New(token.STRING, `"value"`, "value", 1, nil),
				token.New(token.RIGHT_BRACE, "}", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, _, err := parseExpressionFromTokens(tt.tokens)
			if err != nil {
				t.Fatalf("parseExpr() error = %v", err)
			}
			_, pairs := assertHash(t, expr, tt.count)
			if tt.count > 0 {
				assertLiteral(t, pairs[0].Key, "key")
				assertLiteral(t, pairs[0].Value, "value")
			}
		})
	}
}

func TestParseIndexExpression(t *testing.T) {
	// arr[0]
	tokens := []token.Token{
		token.New(token.IDENTIFIER, "arr", nil, 1, nil),
		token.New(token.LEFT_BRACKET, "[", nil, 1, nil),
		token.New(token.NUMBER, "0", 0.0, 1, nil),
		token.New(token.RIGHT_BRACKET, "]", nil, 1, nil),
		token.New(token.EOF, "", nil, 1, nil),
	}
	expr, _, err := parseExpressionFromTokens(tokens)
	if err != nil {
		t.Fatalf("parseExpr() error = %v", err)
	}
	_, obj, index := assertIndex(t, expr)
	assertVariable(t, obj, "arr")
	assertLiteral(t, index, 0.0)
}

func TestParseSetIndexExpression(t *testing.T) {
	// arr[0] = 42
	tokens := []token.Token{
		token.New(token.IDENTIFIER, "arr", nil, 1, nil),
		token.New(token.LEFT_BRACKET, "[", nil, 1, nil),
		token.New(token.NUMBER, "0", 0.0, 1, nil),
		token.New(token.RIGHT_BRACKET, "]", nil, 1, nil),
		token.New(token.EQUAL, "=", nil, 1, nil),
		token.New(token.NUMBER, "42", 42.0, 1, nil),
		token.New(token.EOF, "", nil, 1, nil),
	}
	expr, _, err := parseExpressionFromTokens(tokens)
	if err != nil {
		t.Fatalf("parseExpr() error = %v", err)
	}
	_, obj, index, value := assertSetIndex(t, expr)
	assertVariable(t, obj, "arr")
	assertLiteral(t, index, 0.0)
	assertLiteral(t, value, 42.0)
}

func TestParseRightAssociativeAssignment(t *testing.T) {
	// Assignment is right-associative: a = b = 1 should be a = (b = 1)
	tokens := []token.Token{
		token.New(token.IDENTIFIER, "a", nil, 1, nil),
		token.New(token.EQUAL, "=", nil, 1, nil),
		token.New(token.IDENTIFIER, "b", nil, 1, nil),
		token.New(token.EQUAL, "=", nil, 1, nil),
		token.New(token.NUMBER, "1", 1.0, 1, nil),
		token.New(token.EOF, "", nil, 1, nil),
	}
	expr, _, err := parseExpressionFromTokens(tokens)
	if err != nil {
		t.Fatalf("parseExpr() error = %v", err)
	}
	_, innerAssign := assertAssignment(t, expr, "a")
	_, value := assertAssignment(t, innerAssign, "b")
	assertLiteral(t, value, 1.0)
}

func TestParseInvalidExpressions(t *testing.T) {
	tests := []struct {
		name   string
		tokens []token.Token
	}{
		{
			"missing closing paren",
			// ( 1 + 2
			[]token.Token{
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.NUMBER, "1", 1.0, 1, nil),
				token.New(token.PLUS, "+", nil, 1, nil),
				token.New(token.NUMBER, "2", 2.0, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
		},
		{
			"just operator",
			// +
			[]token.Token{
				token.New(token.PLUS, "+", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, collector, _ := parseExpressionFromTokens(tt.tokens)
			if len(collector.Errors) == 0 {
				t.Error("expected diagnostic errors for invalid expression")
			}
		})
	}
}

func TestPrattParseFunctionExpr(t *testing.T) {
	tests := []struct {
		name   string
		tokens []token.Token
		check  func(*testing.T, ast.Expr)
	}{
		{
			"simple anonymous function",
			[]token.Token{
				token.New(token.FUN, "fun", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.LEFT_BRACE, "{", nil, 1, nil),
				token.New(token.RIGHT_BRACE, "}", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			func(t *testing.T, expr ast.Expr) {
				fn := assertFunctionExpr(t, expr)
				if len(fn.Params) != 0 {
					t.Errorf("expected 0 params, got %d", len(fn.Params))
				}
				if len(fn.Body.Statements) != 0 {
					t.Errorf("expected 0 statements, got %d", len(fn.Body.Statements))
				}
			},
		},
		{
			"anonymous function with params",
			[]token.Token{
				token.New(token.FUN, "fun", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.IDENTIFIER, "a", nil, 1, nil),
				token.New(token.COMMA, ",", nil, 1, nil),
				token.New(token.IDENTIFIER, "b", nil, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.LEFT_BRACE, "{", nil, 1, nil),
				token.New(token.RIGHT_BRACE, "}", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			func(t *testing.T, expr ast.Expr) {
				fn := assertFunctionExpr(t, expr)
				if len(fn.Params) != 2 {
					t.Errorf("expected 2 params, got %d", len(fn.Params))
				}
				if fn.Params[0].Lexeme != "a" || fn.Params[1].Lexeme != "b" {
					t.Errorf("expected params [a, b], got [%s, %s]", fn.Params[0].Lexeme, fn.Params[1].Lexeme)
				}
			},
		},
		{
			"anonymous function with body",
			[]token.Token{
				token.New(token.FUN, "fun", nil, 1, nil),
				token.New(token.LEFT_PAREN, "(", nil, 1, nil),
				token.New(token.RIGHT_PAREN, ")", nil, 1, nil),
				token.New(token.LEFT_BRACE, "{", nil, 1, nil),
				token.New(token.RETURN, "return", nil, 1, nil),
				token.New(token.NUMBER, "1", 1.0, 1, nil),
				token.New(token.SEMICOLON, ";", nil, 1, nil),
				token.New(token.RIGHT_BRACE, "}", nil, 1, nil),
				token.New(token.EOF, "", nil, 1, nil),
			},
			func(t *testing.T, expr ast.Expr) {
				fn := assertFunctionExpr(t, expr)
				if len(fn.Body.Statements) != 1 {
					t.Errorf("expected 1 statement, got %d", len(fn.Body.Statements))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, _, err := parseExpressionFromTokens(tt.tokens)
			if err != nil {
				t.Fatalf("parseExpr() error = %v", err)
			}
			tt.check(t, expr)
		})
	}
}

// Assertion Helpers

func assertLiteral(t *testing.T, expr ast.Expr, expected interface{}) {
	t.Helper()
	lit, ok := expr.(*ast.LiteralExpr)
	if !ok {
		t.Fatalf("expected LiteralExpr, got %T", expr)
	}
	if !reflect.DeepEqual(lit.Value, expected) {
		t.Fatalf("expected literal value %v, got %v", expected, lit.Value)
	}
}

func assertVariable(t *testing.T, expr ast.Expr, expected string) {
	t.Helper()
	varExpr, ok := expr.(*ast.VariableExpr)
	if !ok {
		t.Fatalf("expected VariableExpr, got %T", expr)
	}
	if varExpr.Name.Lexeme != expected {
		t.Fatalf("expected variable name %s, got %s", expected, varExpr.Name.Lexeme)
	}
}

func assertUnary(t *testing.T, expr ast.Expr, op token.Type) (*ast.UnaryExpr, ast.Expr) {
	t.Helper()
	unary, ok := expr.(*ast.UnaryExpr)
	if !ok {
		t.Fatalf("expected UnaryExpr, got %T", expr)
	}
	if unary.Operator.Type != op {
		t.Fatalf("expected operator %v, got %v", op, unary.Operator.Type)
	}
	return unary, unary.Expr
}

func assertBinary(t *testing.T, expr ast.Expr, op token.Type) (*ast.BinaryExpr, ast.Expr, ast.Expr) {
	t.Helper()
	binary, ok := expr.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", expr)
	}
	if binary.Operator.Type != op {
		t.Fatalf("expected operator %v, got %v", op, binary.Operator.Type)
	}
	return binary, binary.Left, binary.Right
}

func assertGrouping(t *testing.T, expr ast.Expr) ast.Expr {
	t.Helper()
	grouping, ok := expr.(*ast.GroupingExpr)
	if !ok {
		t.Fatalf("expected GroupingExpr, got %T", expr)
	}
	return grouping.Expr
}

func assertLogical(t *testing.T, expr ast.Expr, op token.Type) (*ast.LogicalExpr, ast.Expr, ast.Expr) {
	t.Helper()
	logical, ok := expr.(*ast.LogicalExpr)
	if !ok {
		t.Fatalf("expected LogicalExpr, got %T", expr)
	}
	if logical.Operator.Type != op {
		t.Fatalf("expected operator %v, got %v", op, logical.Operator.Type)
	}
	return logical, logical.Left, logical.Right
}

func assertAssignment(t *testing.T, expr ast.Expr, name string) (*ast.AssignExpr, ast.Expr) {
	t.Helper()
	assign, ok := expr.(*ast.AssignExpr)
	if !ok {
		t.Fatalf("expected AssignExpr, got %T", expr)
	}
	if assign.Name.Lexeme != name {
		t.Fatalf("expected assignment to %s, got %s", name, assign.Name.Lexeme)
	}
	return assign, assign.Value
}

func assertCall(t *testing.T, expr ast.Expr, argCount int) (*ast.CallExpr, ast.Expr, []ast.Expr) {
	t.Helper()
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", expr)
	}
	if len(call.Arguments) != argCount {
		t.Fatalf("expected %d arguments, got %d", argCount, len(call.Arguments))
	}
	return call, call.Callee, call.Arguments
}

func assertGet(t *testing.T, expr ast.Expr, name string) (*ast.GetExpr, ast.Expr) {
	t.Helper()
	get, ok := expr.(*ast.GetExpr)
	if !ok {
		t.Fatalf("expected GetExpr, got %T", expr)
	}
	if get.Name.Lexeme != name {
		t.Fatalf("expected property %s, got %s", name, get.Name.Lexeme)
	}
	return get, get.Object
}

func assertSet(t *testing.T, expr ast.Expr, name string) (*ast.SetExpr, ast.Expr, ast.Expr) {
	t.Helper()
	set, ok := expr.(*ast.SetExpr)
	if !ok {
		t.Fatalf("expected SetExpr, got %T", expr)
	}
	if set.Name.Lexeme != name {
		t.Fatalf("expected property %s, got %s", name, set.Name.Lexeme)
	}
	return set, set.Object, set.Value
}

func assertThis(t *testing.T, expr ast.Expr) {
	t.Helper()
	_, ok := expr.(*ast.ThisExpr)
	if !ok {
		t.Fatalf("expected ThisExpr, got %T", expr)
	}
}

func assertSuper(t *testing.T, expr ast.Expr, method string) {
	t.Helper()
	super, ok := expr.(*ast.SuperExpr)
	if !ok {
		t.Fatalf("expected SuperExpr, got %T", expr)
	}
	if super.Method.Lexeme != method {
		t.Fatalf("expected super method %s, got %s", method, super.Method.Lexeme)
	}
}

func assertArray(t *testing.T, expr ast.Expr, count int) (*ast.ArrayLiteralExpr, []ast.Expr) {
	t.Helper()
	arr, ok := expr.(*ast.ArrayLiteralExpr)
	if !ok {
		t.Fatalf("expected ArrayLiteralExpr, got %T", expr)
	}
	if len(arr.Elements) != count {
		t.Fatalf("expected %d elements, got %d", count, len(arr.Elements))
	}
	return arr, arr.Elements
}

func assertHash(t *testing.T, expr ast.Expr, count int) (*ast.HashLiteralExpr, []ast.HashPair) {
	t.Helper()
	hash, ok := expr.(*ast.HashLiteralExpr)
	if !ok {
		t.Fatalf("expected HashLiteralExpr, got %T", expr)
	}
	if len(hash.Pairs) != count {
		t.Fatalf("expected %d pairs, got %d", count, len(hash.Pairs))
	}
	return hash, hash.Pairs
}

func assertIndex(t *testing.T, expr ast.Expr) (*ast.IndexExpr, ast.Expr, ast.Expr) {
	t.Helper()
	idx, ok := expr.(*ast.IndexExpr)
	if !ok {
		t.Fatalf("expected IndexExpr, got %T", expr)
	}
	return idx, idx.Object, idx.Index
}

func assertSetIndex(t *testing.T, expr ast.Expr) (*ast.SetIndexExpr, ast.Expr, ast.Expr, ast.Expr) {
	t.Helper()
	sidx, ok := expr.(*ast.SetIndexExpr)
	if !ok {
		t.Fatalf("expected SetIndexExpr, got %T", expr)
	}
	return sidx, sidx.Object, sidx.Index, sidx.Value
}

func assertFunctionExpr(t *testing.T, expr ast.Expr) *ast.FunctionExpr {
	t.Helper()
	fn, ok := expr.(*ast.FunctionExpr)
	if !ok {
		t.Fatalf("expected FunctionExpr, got %T", expr)
	}
	return fn
}
