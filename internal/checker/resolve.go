package checker

import (
	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/types"
)

// resolveType turns a type annotation into the type it names.
//
// This is where builtin type names are finally distinguished from class names.
// The parser produced a plain NamedType for both, so there is one resolution
// path: look in the builtin table, then at the module's classes, then at an
// import's exports. A name that matches nothing is an error here, with a
// position, rather than a silent success that fails later.
func (c *Checker) resolveType(t ast.TypeExpr) types.Type {
	switch n := t.(type) {
	case nil:
		// An omitted function return type means the function returns nothing.
		return types.Void

	case *ast.NamedType:
		if n.Module != nil {
			return c.resolveQualifiedType(n)
		}
		name := n.Name.Lexeme
		if builtin, ok := builtinTypes[name]; ok {
			return builtin
		}
		if class, ok := c.classes[name]; ok {
			return class
		}
		c.errorAt(n.Name, "Unknown type '"+name+"'.")
		return types.Invalid

	case *ast.ArrayType:
		elem := c.resolveType(n.Elem)
		if !types.IsStorable(elem) {
			c.error(n, "Array elements cannot be void.")
			return types.Invalid
		}
		return &types.Array{Elem: elem}

	case *ast.MapType:
		// The VM's hash accepts string keys only, so the type system does not
		// promise otherwise. The key node is kept in the AST precisely so this
		// error can point at the key rather than the whole declaration.
		if key := c.resolveType(n.Key); key != types.String && key != types.Invalid {
			c.error(n.Key, "Map keys must be string, got "+key.String()+".")
			return types.Invalid
		}
		value := c.resolveType(n.Value)
		if !types.IsStorable(value) {
			c.error(n, "Map values cannot be void.")
			return types.Invalid
		}
		return &types.Map{Value: value}

	case *ast.FunType:
		params := make([]types.Type, 0, len(n.Params))
		for _, p := range n.Params {
			pt := c.resolveType(p)
			if !types.IsStorable(pt) {
				c.error(p, "A parameter cannot be void.")
				pt = types.Invalid
			}
			params = append(params, pt)
		}
		return &types.Function{Params: params, Return: c.resolveType(n.Return)}
	}

	c.error(t, "Unrecognised type.")
	return types.Invalid
}

// resolveQualifiedType resolves `alias.Name` against an imported module.
func (c *Checker) resolveQualifiedType(n *ast.NamedType) types.Type {
	mod, ok := c.imports[n.Module.Lexeme]
	if !ok {
		c.errorAt(n.Module, "Unknown module '"+n.Module.Lexeme+"'.")
		return types.Invalid
	}
	if class, ok := mod.Classes[n.Name.Lexeme]; ok {
		return class
	}
	if _, exported := mod.Exports[n.Name.Lexeme]; exported {
		c.errorAt(n.Name, "'"+n.Module.Lexeme+"."+n.Name.Lexeme+"' is a value, not a type.")
		return types.Invalid
	}
	c.errorAt(n.Name, "'"+n.Name.Lexeme+"' is not exported from module '"+n.Module.Lexeme+"'.")
	return types.Invalid
}

// signatureOf builds a function type from a declaration's annotations.
func (c *Checker) signatureOf(params []ast.Param, returnType ast.TypeExpr) *types.Function {
	paramTypes := make([]types.Type, 0, len(params))
	for _, p := range params {
		pt := c.resolveType(p.Type)
		if !types.IsStorable(pt) {
			c.errorAt(p.Name, "Parameter '"+p.Name.Lexeme+"' cannot be void.")
			pt = types.Invalid
		}
		paramTypes = append(paramTypes, pt)
	}
	return &types.Function{Params: paramTypes, Return: c.resolveType(returnType)}
}
