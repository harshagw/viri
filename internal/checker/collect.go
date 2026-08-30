package checker

import (
	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/token"
	"github.com/harshagw/viri/internal/types"
)

// collect reads every top-level signature before any body is checked, so
// declarations can refer to one another in any order — mutual recursion, a
// class whose method returns a class declared later, and so on. It mirrors the
// hoisting loop the compiler already does in compileModule.
//
// It runs in three sweeps, because a signature can mention a class and a class
// can mention a signature:
//
//  1. create an empty shell for every class, so its name resolves
//  2. fill the shells in — fields, methods, superclass
//  3. declare top-level functions and variables
func (c *Checker) collect(statements []ast.Stmt) {
	c.collectClassNames(statements)
	c.fillClasses(statements)
	c.collectValues(statements)
}

// declareName binds a top-level name, rejecting the two ways it can be wrong:
// reusing a builtin type name, or declaring the same name twice.
func (c *Checker) declareName(tok *token.Token, kind string, t types.Type, isConst bool) {
	if IsBuiltinTypeName(tok.Lexeme) {
		c.errorAt(tok, "Cannot use builtin type name '"+tok.Lexeme+"' as a "+kind+" name.")
		return
	}
	if !c.declare(tok.Lexeme, t, isConst) {
		c.errorAt(tok, "'"+tok.Lexeme+"' is already declared in this scope.")
	}
}

func (c *Checker) collectClassNames(statements []ast.Stmt) {
	for _, stmt := range statements {
		class, ok := stmt.(*ast.ClassStmt)
		if !ok {
			continue
		}
		if IsBuiltinTypeName(class.Name.Lexeme) {
			c.errorAt(class.Name, "Cannot use builtin type name '"+class.Name.Lexeme+"' as a class name.")
			continue
		}
		if _, exists := c.classes[class.Name.Lexeme]; exists {
			c.errorAt(class.Name, "'"+class.Name.Lexeme+"' is already declared in this scope.")
			continue
		}
		built := &types.Class{
			Name:    class.Name.Lexeme,
			Methods: make(map[string]*types.Function),
		}
		c.classes[class.Name.Lexeme] = built
		c.classTypes[class] = built
	}
}

func (c *Checker) fillClasses(statements []ast.Stmt) {
	for _, stmt := range statements {
		decl, ok := stmt.(*ast.ClassStmt)
		if !ok {
			continue
		}
		class := c.classes[decl.Name.Lexeme]
		if class == nil {
			continue // already reported
		}
		c.resolveSuperclass(decl, class)
		c.collectFields(decl, class)
		c.collectMethods(decl, class)
	}
}

func (c *Checker) resolveSuperclass(decl *ast.ClassStmt, class *types.Class) {
	if decl.SuperClass == nil {
		return
	}
	name := decl.SuperClass.Name.Lexeme
	if name == decl.Name.Lexeme {
		// The compiler rejects this too, but the checker runs first and would
		// otherwise loop walking the Super chain.
		c.errorAt(decl.SuperClass.Name, "A class cannot inherit from itself.")
		return
	}
	super, ok := c.classes[name]
	if !ok {
		c.errorAt(decl.SuperClass.Name, "Unknown superclass '"+name+"'.")
		return
	}
	class.Super = super
}

func (c *Checker) collectFields(decl *ast.ClassStmt, class *types.Class) {
	seen := make(map[string]bool, len(decl.Fields))
	for _, field := range decl.Fields {
		name := field.Name.Lexeme
		if seen[name] {
			c.errorAt(field.Name, "Field '"+name+"' is already declared on class '"+class.Name+"'.")
			continue
		}
		seen[name] = true

		fieldType := c.resolveType(field.Type)
		if !types.IsStorable(fieldType) {
			c.errorAt(field.Name, "Field '"+name+"' cannot be void.")
			fieldType = types.Invalid
		}
		class.Fields = append(class.Fields, types.Field{Name: name, Type: fieldType})
	}
}

func (c *Checker) collectMethods(decl *ast.ClassStmt, class *types.Class) {
	for _, method := range decl.Methods {
		name := method.Name.Lexeme
		if _, exists := class.Methods[name]; exists {
			c.errorAt(method.Name, "Method '"+name+"' is already declared on class '"+class.Name+"'.")
			continue
		}

		sig := c.signatureOf(method.Params, method.ReturnType)
		if name == "init" {
			if method.ReturnType != nil {
				c.error(method.ReturnType, "'init' cannot declare a return type.")
			}
			// Calling the class produces an instance, whatever init writes.
			class.Init = &types.Function{Params: sig.Params, Return: class}
			continue
		}
		class.Methods[name] = sig
	}
}

func (c *Checker) collectValues(statements []ast.Stmt) {
	for _, stmt := range statements {
		switch s := stmt.(type) {
		case *ast.ClassStmt:
			if class, ok := c.classes[s.Name.Lexeme]; ok {
				// In value position a class name IS its constructor — a
				// function returning an instance. Binding it to the class
				// type instead would make `var q: Point = Point;` check,
				// because the class itself and an instance of it would be
				// the same type.
				c.declareName(s.Name, "class", class.Constructor(), true)
			}

		case *ast.FunctionStmt:
			c.declareName(s.Name, "function", c.signatureOf(s.Params, s.ReturnType), true)

		case *ast.VarDeclStmt:
			declared := c.resolveType(s.Type)
			if !types.IsStorable(declared) {
				c.errorAt(s.Name, "Variable '"+s.Name.Lexeme+"' cannot be void.")
				declared = types.Invalid
			}
			kind := "variable"
			if s.IsConst {
				kind = "constant"
			}
			c.declareName(s.Name, kind, declared, s.IsConst)
		}
	}
}
