package scanner

import (
	"bytes"
	"testing"

	"github.com/harshagw/viri/internal/token"
)

func TestScanner(t *testing.T) {
	testFile := "test.viri"

	tests := []struct {
		name     string
		input    string
		path     *string
		expected []token.Token
		wantErr  bool
	}{
		{"left paren", "(", nil, []token.Token{{Type: token.LEFT_PAREN, Lexeme: "(", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"right paren", ")", nil, []token.Token{{Type: token.RIGHT_PAREN, Lexeme: ")", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"left brace", "{", nil, []token.Token{{Type: token.LEFT_BRACE, Lexeme: "{", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"right brace", "}", nil, []token.Token{{Type: token.RIGHT_BRACE, Lexeme: "}", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"left bracket", "[", nil, []token.Token{{Type: token.LEFT_BRACKET, Lexeme: "[", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"right bracket", "]", nil, []token.Token{{Type: token.RIGHT_BRACKET, Lexeme: "]", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"comma", ",", nil, []token.Token{{Type: token.COMMA, Lexeme: ",", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"colon", ":", nil, []token.Token{{Type: token.COLON, Lexeme: ":", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"dot", ".", nil, []token.Token{{Type: token.DOT, Lexeme: ".", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"minus", "-", nil, []token.Token{{Type: token.MINUS, Lexeme: "-", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"plus", "+", nil, []token.Token{{Type: token.PLUS, Lexeme: "+", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"semicolon", ";", nil, []token.Token{{Type: token.SEMICOLON, Lexeme: ";", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"star", "*", nil, []token.Token{{Type: token.STAR, Lexeme: "*", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"slash", "/", nil, []token.Token{{Type: token.SLASH, Lexeme: "/", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"bang equal", "!=", nil, []token.Token{{Type: token.BANG_EQUAL, Lexeme: "!=", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"bang", "!", nil, []token.Token{{Type: token.BANG, Lexeme: "!", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"equal equal", "==", nil, []token.Token{{Type: token.EQUAL_EQUAL, Lexeme: "==", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"equal", "=", nil, []token.Token{{Type: token.EQUAL, Lexeme: "=", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"less equal", "<=", nil, []token.Token{{Type: token.LESS_EQUAL, Lexeme: "<=", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"less", "<", nil, []token.Token{{Type: token.LESS, Lexeme: "<", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"greater equal", ">=", nil, []token.Token{{Type: token.GREATER_EQUAL, Lexeme: ">=", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"greater", ">", nil, []token.Token{{Type: token.GREATER, Lexeme: ">", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"simple string", `"hello"`, nil, []token.Token{{Type: token.STRING, Lexeme: `"hello"`, Literal: "hello", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"empty string", `""`, nil, []token.Token{{Type: token.STRING, Lexeme: `""`, Literal: "", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"string with spaces", `"hello world"`, nil, []token.Token{{Type: token.STRING, Lexeme: `"hello world"`, Literal: "hello world", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"multiline string", "\"hello\nworld\"", nil, []token.Token{{Type: token.STRING, Lexeme: "\"hello\nworld\"", Literal: "hello\nworld", Line: 2}, {Type: token.EOF, Lexeme: "", Line: 2}}, false},
		{"unterminated string", `"hello`, nil, nil, true},
		{"integer", "123", nil, []token.Token{{Type: token.NUMBER, Lexeme: "123", Literal: 123.0, Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"float", "123.456", nil, []token.Token{{Type: token.NUMBER, Lexeme: "123.456", Literal: 123.456, Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"identifier", "foo", nil, []token.Token{{Type: token.IDENTIFIER, Lexeme: "foo", Literal: "foo", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"keyword var", "var", nil, []token.Token{{Type: token.VAR, Lexeme: "var", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"keyword fun", "fun", nil, []token.Token{{Type: token.FUN, Lexeme: "fun", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"keyword class", "class", nil, []token.Token{{Type: token.CLASS, Lexeme: "class", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"keyword true", "true", nil, []token.Token{{Type: token.TRUE, Lexeme: "true", Literal: true, Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"keyword false", "false", nil, []token.Token{{Type: token.FALSE, Lexeme: "false", Literal: false, Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"keyword nil", "nil", nil, []token.Token{{Type: token.NIL, Lexeme: "nil", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"keyword and", "and", nil, []token.Token{{Type: token.AND, Lexeme: "and", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"keyword or", "or", nil, []token.Token{{Type: token.OR, Lexeme: "or", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"keyword if", "if", nil, []token.Token{{Type: token.IF, Lexeme: "if", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"keyword else", "else", nil, []token.Token{{Type: token.ELSE, Lexeme: "else", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"keyword for", "for", nil, []token.Token{{Type: token.FOR, Lexeme: "for", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"keyword while", "while", nil, []token.Token{{Type: token.WHILE, Lexeme: "while", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"keyword return", "return", nil, []token.Token{{Type: token.RETURN, Lexeme: "return", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"keyword super", "super", nil, []token.Token{{Type: token.SUPER, Lexeme: "super", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"keyword this", "this", nil, []token.Token{{Type: token.THIS, Lexeme: "this", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"keyword continue", "continue", nil, []token.Token{{Type: token.CONTINUE, Lexeme: "continue", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"keyword break", "break", nil, []token.Token{{Type: token.BREAK, Lexeme: "break", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"keyword import", "import", nil, []token.Token{{Type: token.IMPORT, Lexeme: "import", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"keyword export", "export", nil, []token.Token{{Type: token.EXPORT, Lexeme: "export", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"keyword as", "as", nil, []token.Token{{Type: token.AS, Lexeme: "as", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"identifier with number", "foo123", nil, []token.Token{{Type: token.IDENTIFIER, Lexeme: "foo123", Literal: "foo123", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"keyword as prefix", "variable", nil, []token.Token{{Type: token.IDENTIFIER, Lexeme: "variable", Literal: "variable", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"comment", "// this is a comment\nvar x;", nil, []token.Token{{Type: token.VAR, Lexeme: "var", Line: 2}, {Type: token.IDENTIFIER, Lexeme: "x", Literal: "x", Line: 2}, {Type: token.SEMICOLON, Lexeme: ";", Line: 2}, {Type: token.EOF, Lexeme: "", Line: 2}}, false},
		{"whitespace", " \t\r\n var x;", nil, []token.Token{{Type: token.VAR, Lexeme: "var", Line: 2}, {Type: token.IDENTIFIER, Lexeme: "x", Literal: "x", Line: 2}, {Type: token.SEMICOLON, Lexeme: ";", Line: 2}, {Type: token.EOF, Lexeme: "", Line: 2}}, false},
		{"line numbers", "var x;\nvar y;\nvar z;", nil, []token.Token{
			{Type: token.VAR, Lexeme: "var", Line: 1}, {Type: token.IDENTIFIER, Lexeme: "x", Literal: "x", Line: 1}, {Type: token.SEMICOLON, Lexeme: ";", Line: 1},
			{Type: token.VAR, Lexeme: "var", Line: 2}, {Type: token.IDENTIFIER, Lexeme: "y", Literal: "y", Line: 2}, {Type: token.SEMICOLON, Lexeme: ";", Line: 2},
			{Type: token.VAR, Lexeme: "var", Line: 3}, {Type: token.IDENTIFIER, Lexeme: "z", Literal: "z", Line: 3}, {Type: token.SEMICOLON, Lexeme: ";", Line: 3},
			{Type: token.EOF, Lexeme: "", Line: 3},
		}, false},
		{"unexpected character", "@", nil, nil, true},
		{"complex expression", "var x = 10 + 20 * 30;", nil, []token.Token{
			{Type: token.VAR, Lexeme: "var", Line: 1},
			{Type: token.IDENTIFIER, Lexeme: "x", Literal: "x", Line: 1},
			{Type: token.EQUAL, Lexeme: "=", Line: 1},
			{Type: token.NUMBER, Lexeme: "10", Literal: 10.0, Line: 1},
			{Type: token.PLUS, Lexeme: "+", Line: 1},
			{Type: token.NUMBER, Lexeme: "20", Literal: 20.0, Line: 1},
			{Type: token.STAR, Lexeme: "*", Line: 1},
			{Type: token.NUMBER, Lexeme: "30", Literal: 30.0, Line: 1},
			{Type: token.SEMICOLON, Lexeme: ";", Line: 1},
			{Type: token.EOF, Lexeme: "", Line: 1},
		}, false},
		{"with file path", "var x;", &testFile, []token.Token{
			{Type: token.VAR, Lexeme: "var", Line: 1, FilePath: &testFile},
			{Type: token.IDENTIFIER, Lexeme: "x", Literal: "x", Line: 1, FilePath: &testFile},
			{Type: token.SEMICOLON, Lexeme: ";", Line: 1, FilePath: &testFile},
			{Type: token.EOF, Lexeme: "", Line: 1, FilePath: &testFile},
		}, false},
		{"number with trailing dot", "123.", nil, []token.Token{{Type: token.NUMBER, Lexeme: "123", Literal: 123.0, Line: 1}, {Type: token.DOT, Lexeme: ".", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"multiple dots", "12.34.56", nil, []token.Token{{Type: token.NUMBER, Lexeme: "12.34", Literal: 12.34, Line: 1}, {Type: token.DOT, Lexeme: ".", Line: 1}, {Type: token.NUMBER, Lexeme: "56", Literal: 56.0, Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"dot without digits", ".", nil, []token.Token{{Type: token.DOT, Lexeme: ".", Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
		{"dot then digits", ".5", nil, []token.Token{{Type: token.DOT, Lexeme: ".", Line: 1}, {Type: token.NUMBER, Lexeme: "5", Literal: 5.0, Line: 1}, {Type: token.EOF, Lexeme: "", Line: 1}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := New(bytes.NewBufferString(tt.input), tt.path)
			tokens, err := sc.Scan()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if len(tokens) != len(tt.expected) {
				t.Fatalf("expected %d tokens, got %d", len(tt.expected), len(tokens))
			}

			for i, exp := range tt.expected {
				if tokens[i].Type != exp.Type {
					t.Errorf("token[%d] type = %v, want %v", i, tokens[i].Type, exp.Type)
				}
				if tokens[i].Lexeme != exp.Lexeme {
					t.Errorf("token[%d] lexeme = %q, want %q", i, tokens[i].Lexeme, exp.Lexeme)
				}
				// Only check literal if it's expected
				if exp.Literal != nil {
					if tokens[i].Literal != exp.Literal {
						t.Errorf("token[%d] literal = %v, want %v", i, tokens[i].Literal, exp.Literal)
					}
				}
				if tokens[i].Line != exp.Line {
					t.Errorf("token[%d] line = %d, want %d", i, tokens[i].Line, exp.Line)
				}
				if exp.FilePath != nil {
					if tokens[i].FilePath != exp.FilePath {
						t.Errorf("token[%d] path = %v, want %v", i, tokens[i].FilePath, exp.FilePath)
					}
				}
			}
		})
	}
}
