package checker

import (
	"strconv"

	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/token"
	"github.com/harshagw/viri/internal/types"
)

// checkExpr gives an expression a type and records it.
func (c *Checker) checkExpr(expr ast.Expr) types.Type {
	return c.checkExprExpecting(expr, nil)
}

// checkExprExpecting types an expression, passing down the type the context
// wants where that matters. Only empty collection literals need it: `[]` has
// no element to infer from, and `var xs: []number = [];` is exactly the case
// the mandatory annotation exists to serve.
func (c *Checker) checkExprExpecting(expr ast.Expr, want types.Type) types.Type {
	if expr == nil {
		return types.Invalid
	}
	t := c.synthesize(expr, want)
	c.exprTypes[expr] = t
	return t
}

func (c *Checker) synthesize(expr ast.Expr, want types.Type) types.Type {
	switch e := expr.(type) {
	case *ast.LiteralExpr:
		return literalType(e)

	case *ast.GroupingExpr:
		return c.checkExprExpecting(e.Expr, want)

	case *ast.VariableExpr:
		return c.checkVariable(e)

	case *ast.ThisExpr:
		if c.currentClass == nil {
			c.error(e, "'this' can only be used inside a class.")
			return types.Invalid
		}
		return c.currentClass

	case *ast.AssignExpr:
		return c.checkAssign(e)

	case *ast.UnaryExpr:
		return c.checkUnary(e)

	case *ast.BinaryExpr:
		return c.checkBinary(e)

	case *ast.LogicalExpr:
		return c.checkLogical(e)

	case *ast.CallExpr:
		return c.checkCall(e)

	case *ast.GetExpr:
		return c.checkGet(e)

	case *ast.SetExpr:
		return c.checkSet(e)

	case *ast.SuperExpr:
		return c.checkSuper(e)

	case *ast.IndexExpr:
		return c.checkIndex(e)

	case *ast.SetIndexExpr:
		return c.checkSetIndex(e)

	case *ast.ArrayLiteralExpr:
		return c.checkArrayLiteral(e, want)

	case *ast.HashLiteralExpr:
		return c.checkHashLiteral(e, want)

	case *ast.FunctionExpr:
		return c.checkFunctionExpr(e)
	}

	c.error(expr, "Unrecognised expression.")
	return types.Invalid
}

func literalType(e *ast.LiteralExpr) types.Type {
	switch e.Value.(type) {
	case float64, int:
		return types.Number
	case string:
		return types.String
	case bool:
		return types.Bool
	}
	return types.Invalid
}

func (c *Checker) checkVariable(e *ast.VariableExpr) types.Type {
	name := e.Name.Lexeme
	if b, ok := c.lookup(name); ok {
		return b.typ
	}
	// An import alias is a name too, but it is only meaningful qualified.
	if mod, ok := c.imports[name]; ok {
		return mod
	}
	if IsBuiltinTypeName(name) {
		c.errorAt(e.Name, "'"+name+"' is a type, not a value.")
		return types.Invalid
	}
	c.errorAt(e.Name, "Undefined variable '"+name+"'.")
	return types.Invalid
}

func (c *Checker) checkAssign(e *ast.AssignExpr) types.Type {
	b, ok := c.lookup(e.Name.Lexeme)
	if !ok {
		c.errorAt(e.Name, "Undefined variable '"+e.Name.Lexeme+"'.")
		c.checkExpr(e.Value)
		return types.Invalid
	}
	if b.isConst {
		c.errorAt(e.Name, "Cannot assign to constant '"+e.Name.Lexeme+"'.")
	}
	got := c.checkExprExpecting(e.Value, b.typ)
	if !types.AssignableTo(got, b.typ) {
		c.error(e.Value, "Cannot assign "+got.String()+" to '"+e.Name.Lexeme+
			"' of type "+b.typ.String()+".")
	}
	return b.typ
}

func (c *Checker) checkUnary(e *ast.UnaryExpr) types.Type {
	operand := c.checkExpr(e.Expr)
	if operand == types.Invalid {
		return types.Invalid
	}
	switch e.Operator.Type {
	case token.BANG:
		if operand != types.Bool {
			c.error(e, "'!' needs a bool, got "+operand.String()+".")
			return types.Invalid
		}
		return types.Bool
	case token.MINUS:
		if operand != types.Number {
			c.error(e, "Unary '-' needs a number, got "+operand.String()+".")
			return types.Invalid
		}
		return types.Number
	}
	c.error(e, "Unrecognised unary operator '"+e.Operator.Lexeme+"'.")
	return types.Invalid
}

func (c *Checker) checkBinary(e *ast.BinaryExpr) types.Type {
	left := c.checkExpr(e.Left)
	right := c.checkExpr(e.Right)
	op := e.Operator.Lexeme

	if left == types.Invalid || right == types.Invalid {
		return binaryResult(e.Operator.Type)
	}

	switch e.Operator.Type {
	case token.PLUS:
		// The one overloaded operator: numbers add, strings concatenate.
		if left == types.Number && right == types.Number {
			return types.Number
		}
		if left == types.String && right == types.String {
			return types.String
		}
		c.error(e, "'+' needs two numbers or two strings, got "+
			left.String()+" and "+right.String()+".")
		return types.Invalid

	case token.MINUS, token.STAR, token.SLASH:
		if left != types.Number || right != types.Number {
			c.error(e, "'"+op+"' needs two numbers, got "+
				left.String()+" and "+right.String()+".")
			return types.Invalid
		}
		return types.Number

	case token.LESS, token.LESS_EQUAL, token.GREATER, token.GREATER_EQUAL:
		if left != types.Number || right != types.Number {
			c.error(e, "'"+op+"' needs two numbers, got "+
				left.String()+" and "+right.String()+".")
			return types.Bool
		}
		return types.Bool

	case token.EQUAL_EQUAL, token.BANG_EQUAL:
		// Comparing unrelated types is a mistake, not a false result.
		if !types.Comparable(left, right) {
			c.error(e, "'"+op+"' needs operands of the same type, got "+
				left.String()+" and "+right.String()+".")
		}
		return types.Bool
	}

	c.error(e, "Unrecognised operator '"+op+"'.")
	return types.Invalid
}

// binaryResult is the type an operator yields once its operands are known bad,
// so one error does not cascade into the enclosing expression.
func binaryResult(op token.Type) types.Type {
	switch op {
	case token.LESS, token.LESS_EQUAL, token.GREATER, token.GREATER_EQUAL,
		token.EQUAL_EQUAL, token.BANG_EQUAL:
		return types.Bool
	}
	return types.Invalid
}

func (c *Checker) checkLogical(e *ast.LogicalExpr) types.Type {
	left := c.checkExpr(e.Left)
	right := c.checkExpr(e.Right)
	// 'and' and 'or' produce bool rather than one of their operands, so the
	// result is usable as a condition without truthiness.
	if left != types.Bool && left != types.Invalid {
		c.error(e.Left, "'"+e.Operator.Lexeme+"' needs bool operands, got "+left.String()+".")
	}
	if right != types.Bool && right != types.Invalid {
		c.error(e.Right, "'"+e.Operator.Lexeme+"' needs bool operands, got "+right.String()+".")
	}
	return types.Bool
}

func (c *Checker) checkFunctionExpr(e *ast.FunctionExpr) types.Type {
	sig := c.signatureOf(e.Params, e.ReturnType)
	c.checkFunctionBody(e.Params, sig, e.Body)
	return sig
}

// --- calls -----------------------------------------------------------------

func (c *Checker) checkCall(e *ast.CallExpr) types.Type {
	callee := c.checkExpr(e.Callee)

	if callee == types.Invalid {
		for _, arg := range e.Arguments {
			c.checkExpr(arg)
		}
		return types.Invalid
	}

	fn, ok := callee.(*types.Function)
	if !ok {
		c.error(e, "Cannot call "+callee.String()+"; it is not a function.")
		for _, arg := range e.Arguments {
			c.checkExpr(arg)
		}
		return types.Invalid
	}

	// len is variadic in its parameter type in a way Viri cannot spell.
	if fn == nativeLen {
		return c.checkLenCall(e)
	}

	if len(e.Arguments) != len(fn.Params) {
		c.error(e, "Expected "+strconv.Itoa(len(fn.Params))+" arguments but got "+
			strconv.Itoa(len(e.Arguments))+".")
	}

	for i, arg := range e.Arguments {
		if i >= len(fn.Params) {
			c.checkExpr(arg)
			continue
		}
		want := fn.Params[i]
		got := c.checkExprExpecting(arg, want)
		if !types.AssignableTo(got, want) {
			c.error(arg, "Argument "+strconv.Itoa(i+1)+" is "+got.String()+
				", expected "+want.String()+".")
		}
	}
	return fn.Return
}

// checkLenCall hand-checks len(x), which accepts a string, an array, or a map.
func (c *Checker) checkLenCall(e *ast.CallExpr) types.Type {
	if len(e.Arguments) != 1 {
		c.error(e, "Expected 1 argument but got "+strconv.Itoa(len(e.Arguments))+".")
		for _, arg := range e.Arguments {
			c.checkExpr(arg)
		}
		return types.Number
	}
	arg := c.checkExpr(e.Arguments[0])
	switch arg.(type) {
	case *types.Array, *types.Map:
		return types.Number
	}
	if arg == types.String || arg == types.Invalid {
		return types.Number
	}
	c.error(e.Arguments[0], "len needs a string, array or map, got "+arg.String()+".")
	return types.Number
}

// --- properties ------------------------------------------------------------

func (c *Checker) checkGet(e *ast.GetExpr) types.Type {
	object := c.checkExpr(e.Object)
	if object == types.Invalid {
		return types.Invalid
	}

	// `alias.name` on an import reads that module's export.
	if mod, ok := object.(*types.Module); ok {
		if mod == nil {
			return types.Invalid
		}
		exported, ok := mod.Exports[e.Name.Lexeme]
		if !ok {
			c.errorAt(e.Name, "'"+e.Name.Lexeme+"' is not exported from module '"+mod.Path+"'.")
			return types.Invalid
		}
		return exported
	}

	class, ok := object.(*types.Class)
	if !ok {
		c.errorAt(e.Name, "Cannot read property '"+e.Name.Lexeme+"' of "+object.String()+".")
		return types.Invalid
	}

	if field, ok := class.LookupField(e.Name.Lexeme); ok {
		return field.Type
	}
	if method, ok := class.LookupMethod(e.Name.Lexeme); ok {
		return method
	}
	c.errorAt(e.Name, "Class '"+class.Name+"' has no field or method '"+e.Name.Lexeme+"'.")
	return types.Invalid
}

func (c *Checker) checkSet(e *ast.SetExpr) types.Type {
	object := c.checkExpr(e.Object)
	if object == types.Invalid {
		c.checkExpr(e.Value)
		return types.Invalid
	}

	class, ok := object.(*types.Class)
	if !ok {
		c.errorAt(e.Name, "Cannot assign to property '"+e.Name.Lexeme+"' of "+object.String()+".")
		c.checkExpr(e.Value)
		return types.Invalid
	}

	field, ok := class.LookupField(e.Name.Lexeme)
	if !ok {
		// Dynamic field creation is gone: a field must be declared on the
		// class before it can be assigned.
		if _, isMethod := class.LookupMethod(e.Name.Lexeme); isMethod {
			c.errorAt(e.Name, "Cannot assign to method '"+e.Name.Lexeme+"' of class '"+class.Name+"'.")
		} else {
			c.errorAt(e.Name, "Class '"+class.Name+"' has no field '"+e.Name.Lexeme+
				"'. Declare it on the class to assign it.")
		}
		c.checkExpr(e.Value)
		return types.Invalid
	}

	got := c.checkExprExpecting(e.Value, field.Type)
	if !types.AssignableTo(got, field.Type) {
		c.error(e.Value, "Cannot assign "+got.String()+" to field '"+e.Name.Lexeme+
			"' of type "+field.Type.String()+".")
	}
	return field.Type
}

func (c *Checker) checkSuper(e *ast.SuperExpr) types.Type {
	if c.currentClass == nil {
		c.error(e, "'super' can only be used inside a class.")
		return types.Invalid
	}
	if c.currentClass.Super == nil {
		c.error(e, "'super' can only be used in a class with a superclass.")
		return types.Invalid
	}

	super := c.currentClass.Super
	if e.Method.Lexeme == "init" {
		return super.Constructor()
	}
	method, ok := super.LookupMethod(e.Method.Lexeme)
	if !ok {
		c.errorAt(e.Method, "Superclass '"+super.Name+"' has no method '"+e.Method.Lexeme+"'.")
		return types.Invalid
	}
	return method
}

// --- indexing --------------------------------------------------------------

func (c *Checker) checkIndex(e *ast.IndexExpr) types.Type {
	object := c.checkExpr(e.Object)
	index := c.checkExpr(e.Index)
	if object == types.Invalid {
		return types.Invalid
	}

	switch o := object.(type) {
	case *types.Array:
		c.requireIndex(e.Index, index, types.Number, "Array")
		return o.Elem
	case *types.Map:
		c.requireIndex(e.Index, index, types.String, "Map")
		return o.Value
	}
	c.error(e, "Cannot index "+object.String()+"; only arrays and maps can be indexed.")
	return types.Invalid
}

func (c *Checker) checkSetIndex(e *ast.SetIndexExpr) types.Type {
	object := c.checkExpr(e.Object)
	index := c.checkExpr(e.Index)
	if object == types.Invalid {
		c.checkExpr(e.Value)
		return types.Invalid
	}

	var elem types.Type
	switch o := object.(type) {
	case *types.Array:
		c.requireIndex(e.Index, index, types.Number, "Array")
		elem = o.Elem
	case *types.Map:
		c.requireIndex(e.Index, index, types.String, "Map")
		elem = o.Value
	default:
		c.error(e, "Cannot index "+object.String()+"; only arrays and maps can be indexed.")
		c.checkExpr(e.Value)
		return types.Invalid
	}

	got := c.checkExprExpecting(e.Value, elem)
	if !types.AssignableTo(got, elem) {
		c.error(e.Value, "Cannot store "+got.String()+" in a collection of "+elem.String()+".")
	}
	return elem
}

func (c *Checker) requireIndex(node ast.Expr, got, want types.Type, kind string) {
	if got == types.Invalid || got == want {
		return
	}
	c.error(node, kind+" indices must be "+want.String()+", got "+got.String()+".")
}

// --- collection literals ---------------------------------------------------

// checkArrayLiteral types `[a, b, c]`.
//
// With an expected type from the context, elements are checked against its
// element type — which is what makes `var xs: []number = [];` work. Without
// one, the elements must agree with each other exactly, since there is nothing
// to widen towards.
func (c *Checker) checkArrayLiteral(e *ast.ArrayLiteralExpr, want types.Type) types.Type {
	if wanted, ok := want.(*types.Array); ok {
		for _, element := range e.Elements {
			got := c.checkExprExpecting(element, wanted.Elem)
			if !types.AssignableTo(got, wanted.Elem) {
				c.error(element, "Cannot store "+got.String()+" in "+wanted.String()+".")
			}
		}
		return wanted
	}

	if len(e.Elements) == 0 {
		c.error(e, "Cannot infer the type of an empty array; annotate the declaration.")
		return types.Invalid
	}

	elem := c.checkExpr(e.Elements[0])
	if !types.IsStorable(elem) {
		c.error(e.Elements[0], "Array elements cannot be void.")
		return types.Invalid
	}
	for _, element := range e.Elements[1:] {
		got := c.checkExpr(element)
		if !types.Equal(got, elem) && got != types.Invalid {
			c.error(element, "Array elements must all be "+elem.String()+", got "+got.String()+".")
		}
	}
	return &types.Array{Elem: elem}
}

// checkHashLiteral types `{"k": v}`. Keys are always strings.
func (c *Checker) checkHashLiteral(e *ast.HashLiteralExpr, want types.Type) types.Type {
	wanted, hasWanted := want.(*types.Map)

	for _, pair := range e.Pairs {
		if key := c.checkExpr(pair.Key); key != types.String && key != types.Invalid {
			c.error(pair.Key, "Map keys must be string, got "+key.String()+".")
		}
	}

	if hasWanted {
		for _, pair := range e.Pairs {
			got := c.checkExprExpecting(pair.Value, wanted.Value)
			if !types.AssignableTo(got, wanted.Value) {
				c.error(pair.Value, "Cannot store "+got.String()+" in "+wanted.String()+".")
			}
		}
		return wanted
	}

	if len(e.Pairs) == 0 {
		c.error(e, "Cannot infer the type of an empty map; annotate the declaration.")
		return types.Invalid
	}

	value := c.checkExpr(e.Pairs[0].Value)
	if !types.IsStorable(value) {
		c.error(e.Pairs[0].Value, "Map values cannot be void.")
		return types.Invalid
	}
	for _, pair := range e.Pairs[1:] {
		got := c.checkExpr(pair.Value)
		if !types.Equal(got, value) && got != types.Invalid {
			c.error(pair.Value, "Map values must all be "+value.String()+", got "+got.String()+".")
		}
	}
	return &types.Map{Value: value}
}
