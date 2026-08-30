package parser

import (
	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/token"
)

// parseType parses a type annotation:
//
//	type := IDENTIFIER ['.' IDENTIFIER]            a builtin, class, or module.Class
//	      | '[' ']' type                           []T
//	      | 'map' '[' type ']' type                map[K]V
//	      | 'fun' '(' [type {',' type}] ')' [':' type]
//
// Builtin type names are not special here. `number` and `Shape` are both plain
// identifiers producing a NamedType; which one names a builtin and which names
// a class is decided by the checker, against its scopes. That keeps a single
// resolution path for every name in type position, and means adding a builtin
// type later does not touch the parser.
func (p *Parser) parseType() (ast.TypeExpr, error) {
	switch {
	case p.match(token.IDENTIFIER):
		name := p.peekPrevious()
		// A type may be qualified by an import alias: `shapes.Circle`.
		if p.match(token.DOT) {
			qualified, err := p.consume(token.IDENTIFIER, "Expect type name after '.'.")
			if err != nil {
				return nil, err
			}
			return &ast.NamedType{Module: name, Name: qualified}, nil
		}
		return &ast.NamedType{Name: name}, nil

	case p.match(token.LEFT_BRACKET):
		bracket := p.peekPrevious()
		if _, err := p.consume(token.RIGHT_BRACKET, "Expect ']' after '[' in array type."); err != nil {
			return nil, err
		}
		elem, err := p.parseType()
		if err != nil {
			return nil, err
		}
		return &ast.ArrayType{Elem: elem, Bracket: bracket}, nil

	case p.match(token.MAP):
		keyword := p.peekPrevious()
		if _, err := p.consume(token.LEFT_BRACKET, "Expect '[' after 'map'."); err != nil {
			return nil, err
		}
		key, err := p.parseType()
		if err != nil {
			return nil, err
		}
		if _, err := p.consume(token.RIGHT_BRACKET, "Expect ']' after map key type."); err != nil {
			return nil, err
		}
		value, err := p.parseType()
		if err != nil {
			return nil, err
		}
		return &ast.MapType{Key: key, Value: value, Keyword: keyword}, nil

	case p.match(token.FUN):
		keyword := p.peekPrevious()
		if _, err := p.consume(token.LEFT_PAREN, "Expect '(' after 'fun' in function type."); err != nil {
			return nil, err
		}
		params := make([]ast.TypeExpr, 0)
		if !p.check(token.RIGHT_PAREN) {
			for {
				param, err := p.parseType()
				if err != nil {
					return nil, err
				}
				params = append(params, param)
				if !p.match(token.COMMA) {
					break
				}
			}
		}
		if _, err := p.consume(token.RIGHT_PAREN, "Expect ')' after function type parameters."); err != nil {
			return nil, err
		}
		// A return type is optional; omitting it means the function returns
		// nothing.
		var ret ast.TypeExpr
		if p.match(token.COLON) {
			var err error
			ret, err = p.parseType()
			if err != nil {
				return nil, err
			}
		}
		return &ast.FunType{Params: params, Return: ret, Keyword: keyword}, nil
	}

	return nil, p.error(p.peekCurrent(), "Expect a type.")
}

// parseTypeAnnotation consumes ': type'. Annotations are required by the
// grammar, so the colon is consumed rather than matched — the error names what
// was being declared.
func (p *Parser) parseTypeAnnotation(what string) (ast.TypeExpr, error) {
	if _, err := p.consume(token.COLON, "Expect ':' and a type after "+what+"."); err != nil {
		return nil, err
	}
	return p.parseType()
}
