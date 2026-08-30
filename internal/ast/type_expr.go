package ast

import "github.com/harshagw/viri/internal/token"

// TypeExpr is a type as it appears in source. It is syntax, not a type: a
// NamedType records only that the programmer wrote a particular word in a type
// position. Resolving that word — to a builtin, a class, or nothing at all —
// is the checker's job, which is why there is no reference to internal/types
// here.
//
// TypeExpr is kept separate from Expr because the two never appear in the same
// position and share no meaning. `x[y]` is an index operation as an Expr and
// part of a map type as a TypeExpr; `[]number` and `fun(number): number` are
// not expressions at all. Keeping them apart means a declaration cannot hold
// `3 + 4` where a type belongs.
type TypeExpr interface {
	Node
	typeExprNode()
}

// NamedType is a name in type position: a builtin (`number`, `string`,
// `bool`), a class, or a class imported from another module (`shapes.Circle`).
// The parser does not distinguish them — all are plain identifiers, resolved
// later against the checker's scopes.
type NamedType struct {
	// Module is the import alias qualifying the name, or nil when the type
	// is named directly.
	Module *token.Token
	Name   *token.Token
}

func (*NamedType) typeExprNode() {}
func (t *NamedType) GetPrimaryToken() *token.Token {
	if t.Module != nil {
		return t.Module
	}
	return t.Name
}

// ArrayType is `[]T`.
type ArrayType struct {
	Elem TypeExpr
	// Bracket is the opening '[', which is the whole type's position.
	Bracket *token.Token
}

func (*ArrayType) typeExprNode()                   {}
func (t *ArrayType) GetPrimaryToken() *token.Token { return t.Bracket }

// MapType is `map[K]V`. Key is retained even though the checker requires it to
// be `string`, so the error can point at the offending key type rather than at
// the declaration as a whole.
type MapType struct {
	Key     TypeExpr
	Value   TypeExpr
	Keyword *token.Token // the 'map'
}

func (*MapType) typeExprNode()                   {}
func (t *MapType) GetPrimaryToken() *token.Token { return t.Keyword }

// FunType is `fun(T, ...): R`. Return is nil for a function returning nothing.
// Unlike FunctionExpr it has no parameter names and no body — a function type
// describes a signature, not a function.
type FunType struct {
	Params  []TypeExpr
	Return  TypeExpr
	Keyword *token.Token // the 'fun'
}

func (*FunType) typeExprNode()                   {}
func (t *FunType) GetPrimaryToken() *token.Token { return t.Keyword }

// String renders a TypeExpr back to source form, for diagnostics.
func TypeExprString(t TypeExpr) string {
	switch n := t.(type) {
	case nil:
		return "nil"
	case *NamedType:
		if n.Module != nil {
			return n.Module.Lexeme + "." + n.Name.Lexeme
		}
		return n.Name.Lexeme
	case *ArrayType:
		return "[]" + TypeExprString(n.Elem)
	case *MapType:
		return "map[" + TypeExprString(n.Key) + "]" + TypeExprString(n.Value)
	case *FunType:
		out := "fun("
		for i, p := range n.Params {
			if i > 0 {
				out += ", "
			}
			out += TypeExprString(p)
		}
		out += ")"
		if n.Return != nil {
			out += ": " + TypeExprString(n.Return)
		}
		return out
	}
	return "<unknown type>"
}

// Param is one declared function parameter. Type is never nil: annotations are
// required by the grammar, so the parser rejects a parameter without one.
type Param struct {
	Name *token.Token
	Type TypeExpr
}

// FieldDecl is one declared class field, e.g. `name: string;`.
type FieldDecl struct {
	Name *token.Token
	Type TypeExpr
}

func (f *FieldDecl) GetPrimaryToken() *token.Token { return f.Name }
