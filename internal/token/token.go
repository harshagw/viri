package token

import "fmt"

// Type represents a lexical token type.
type Type int

const (
	// Single-character tokens.
	LEFT_PAREN Type = iota
	RIGHT_PAREN
	LEFT_BRACE
	RIGHT_BRACE
	LEFT_BRACKET
	RIGHT_BRACKET
	COMMA
	DOT
	MINUS
	PLUS
	SEMICOLON
	SLASH
	STAR
	COLON

	// One or two character tokens.
	BANG
	BANG_EQUAL
	EQUAL
	EQUAL_EQUAL
	GREATER
	GREATER_EQUAL
	LESS
	LESS_EQUAL

	// Literals.
	IDENTIFIER
	STRING
	NUMBER

	// Keywords.
	AND
	OR
	VAR
	PRINT

	IF
	ELSE
	FOR
	WHILE

	TRUE
	FALSE
	NIL

	FUN
	CLASS
	RETURN
	SUPER
	THIS
	CONTINUE
	BREAK

	IMPORT
	EXPORT
	AS

	EOF
)

func (tt Type) String() string {
	switch tt {
	case LEFT_PAREN:
		return "LEFT_PAREN"
	case RIGHT_PAREN:
		return "RIGHT_PAREN"
	case LEFT_BRACE:
		return "LEFT_BRACE"
	case RIGHT_BRACE:
		return "RIGHT_BRACE"
	case LEFT_BRACKET:
		return "LEFT_BRACKET"
	case RIGHT_BRACKET:
		return "RIGHT_BRACKET"
	case COMMA:
		return "COMMA"
	case DOT:
		return "DOT"
	case MINUS:
		return "MINUS"
	case PLUS:
		return "PLUS"
	case SEMICOLON:
		return "SEMICOLON"
	case SLASH:
		return "SLASH"
	case STAR:
		return "STAR"
	case COLON:
		return "COLON"
	case BANG:
		return "BANG"
	case BANG_EQUAL:
		return "BANG_EQUAL"
	case EQUAL:
		return "EQUAL"
	case EQUAL_EQUAL:
		return "EQUAL_EQUAL"
	case GREATER:
		return "GREATER"
	case GREATER_EQUAL:
		return "GREATER_EQUAL"
	case LESS:
		return "LESS"
	case LESS_EQUAL:
		return "LESS_EQUAL"
	case IDENTIFIER:
		return "IDENTIFIER"
	case STRING:
		return "STRING"
	case NUMBER:
		return "NUMBER"
	case AND:
		return "AND"
	case OR:
		return "OR"
	case VAR:
		return "VAR"
	case PRINT:
		return "PRINT"
	case IF:
		return "IF"
	case ELSE:
		return "ELSE"
	case FOR:
		return "FOR"
	case WHILE:
		return "WHILE"
	case TRUE:
		return "TRUE"
	case FALSE:
		return "FALSE"
	case NIL:
		return "NIL"
	case FUN:
		return "FUN"
	case CLASS:
		return "CLASS"
	case RETURN:
		return "RETURN"
	case SUPER:
		return "SUPER"
	case THIS:
		return "THIS"
	case CONTINUE:
		return "CONTINUE"
	case BREAK:
		return "BREAK"
	case IMPORT:
		return "IMPORT"
	case EXPORT:
		return "EXPORT"
	case AS:
		return "AS"
	case EOF:
		return "EOF"
	default:
		return "UNKNOWN"
	}
}

// Token is a lexical token produced by the scanner.
type Token struct {
	Type     Type
	Lexeme   string
	Literal  interface{}
	Line     int
	FilePath *string 
}

func New(tt Type, lexeme string, literal interface{}, line int, filePath *string) Token {
	return Token{
		Type:     tt,
		Lexeme:   lexeme,
		Literal:  literal,
		Line:     line,
		FilePath: filePath,
	}
}

func (t *Token) String() string {
	return fmt.Sprintf("%s %s %v %v", t.Type.String(), t.Lexeme, t.Literal, t.Line)
}
