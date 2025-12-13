package token

// keywordLookup maps reserved words to their token types.
var keywordLookup = map[string]Type{
	"and":      AND,
	"or":       OR,
	"var":      VAR,
	"print":    PRINT,
	"if":       IF,
	"else":     ELSE,
	"for":      FOR,
	"while":    WHILE,
	"true":     TRUE,
	"false":    FALSE,
	"nil":      NIL,
	"fun":      FUN,
	"class":    CLASS,
	"return":   RETURN,
	"super":    SUPER,
	"this":     THIS,
	"continue": CONTINUE,
	"break":    BREAK,
	"import":   IMPORT,
	"export":   EXPORT,
	"as":       AS,
}

// LookupKeyword returns the token type for a keyword or IDENTIFIER if not reserved.
func LookupKeyword(text string) Type {
	if tt, ok := keywordLookup[text]; ok {
		return tt
	}
	return IDENTIFIER
}
