package checker

import (
	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/types"
)

// atModuleLevel reports whether the walk is in the outermost scope, where
// collect has already declared every name. Declaring them again there would
// report a spurious duplicate.
func (c *Checker) atModuleLevel() bool {
	return len(c.scopes) == 1
}

func (c *Checker) checkStatements(statements []ast.Stmt) {
	for _, stmt := range statements {
		c.checkStatement(stmt)
	}
}

func (c *Checker) checkStatement(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case nil:
		return

	case *ast.ExprStmt:
		// The one position where a void call is legal: its result is
		// discarded rather than used.
		c.checkExpr(s.Expr)

	case *ast.PrintStmt:
		// print accepts any value. It is the single place where types are
		// erased, and it is safe because print is a statement, not a value
		// position — nothing downstream can consume what it printed.
		if t := c.checkExpr(s.Expr); t == types.Void {
			c.error(s.Expr, "Cannot print the result of a function that returns nothing.")
		}

	case *ast.VarDeclStmt:
		c.checkVarDecl(s)

	case *ast.BlockStmt:
		c.beginScope()
		c.checkStatements(s.Statements)
		c.endScope()

	case *ast.IfStmt:
		c.checkCondition(s.Condition, "if")
		c.checkStatement(s.ThenBranch)
		c.checkStatement(s.ElseBranch)

	case *ast.WhileStmt:
		c.checkCondition(s.Condition, "while")
		c.checkStatement(s.Body)

	case *ast.ForStmt:
		// The initialiser's scope is the loop, so a sibling loop can reuse
		// the same variable name.
		c.beginScope()
		c.checkStatement(s.Initializer)
		if s.Condition != nil {
			c.checkCondition(s.Condition, "for")
		}
		if s.Increment != nil {
			c.checkExpr(s.Increment)
		}
		c.checkStatement(s.Body)
		c.endScope()

	case *ast.FunctionStmt:
		c.checkFunctionStmt(s)

	case *ast.ReturnStmt:
		c.checkReturn(s)

	case *ast.ClassStmt:
		c.checkClassStmt(s)

	case *ast.BreakStmt, *ast.ContinueStmt, *ast.ImportStmt:
		// Nothing to type. Placement is checked by the compiler.

	default:
		c.error(stmt, "Unrecognised statement.")
	}
}

// checkCondition enforces that control flow branches on a real bool. There is
// no truthiness in Viri 2.0, which is what lets the VM drop its IsTruthy call
// in Phase 4.
func (c *Checker) checkCondition(expr ast.Expr, keyword string) {
	t := c.checkExpr(expr)
	if t == types.Invalid {
		return
	}
	if t != types.Bool {
		article := "A"
		if keyword == "if" {
			article = "An"
		}
		c.error(expr, article+" "+keyword+" condition must be bool, got "+t.String()+".")
	}
}

func (c *Checker) checkVarDecl(s *ast.VarDeclStmt) {
	// At module level collect already resolved this annotation and bound the
	// name. Resolving it again would report any problem with it twice.
	var declared types.Type
	if c.atModuleLevel() {
		if b, ok := c.lookup(s.Name.Lexeme); ok {
			declared = b.typ
		}
	}
	if declared == nil {
		declared = c.resolveType(s.Type)
		if !types.IsStorable(declared) {
			c.errorAt(s.Name, "Variable '"+s.Name.Lexeme+"' cannot be void.")
			declared = types.Invalid
		}
	}

	// Every type is non-nilable, so there is no correct value to default to.
	if s.Initializer == nil {
		c.errorAt(s.Name, "'"+s.Name.Lexeme+"' needs an initializer; "+
			declared.String()+" has no default value.")
	} else {
		actual := c.checkExprExpecting(s.Initializer, declared)
		if !types.AssignableTo(actual, declared) {
			c.error(s.Initializer, "Cannot assign "+actual.String()+" to '"+
				s.Name.Lexeme+"' of type "+declared.String()+".")
		}
	}

	if !c.atModuleLevel() {
		kind := "variable"
		if s.IsConst {
			kind = "constant"
		}
		c.declareName(s.Name, kind, declared, s.IsConst)
	}
}

func (c *Checker) checkFunctionStmt(s *ast.FunctionStmt) {
	// As with variables, a module-level signature was already resolved by
	// collect; reuse it rather than reporting its errors a second time.
	var sig *types.Function
	if c.atModuleLevel() {
		if b, ok := c.lookup(s.Name.Lexeme); ok {
			sig, _ = b.typ.(*types.Function)
		}
	}
	if sig == nil {
		sig = c.signatureOf(s.Params, s.ReturnType)
		if !c.atModuleLevel() {
			c.declareName(s.Name, "function", sig, true)
		}
	}
	c.checkFunctionBody(s.Params, sig, s.Body)
}

// checkFunctionBody walks a function body with its parameters in scope and its
// signature as the return contract.
func (c *Checker) checkFunctionBody(params []ast.Param, sig *types.Function, body *ast.BlockStmt) {
	previousFn, previousInit := c.currentFn, c.inInit
	c.currentFn, c.inInit = sig, false
	defer func() { c.currentFn, c.inInit = previousFn, previousInit }()

	c.beginScope()
	for i, p := range params {
		var t types.Type = types.Invalid
		if i < len(sig.Params) {
			t = sig.Params[i]
		}
		c.declareName(p.Name, "parameter", t, false)
	}
	if body != nil {
		c.checkStatements(body.Statements)
	}
	c.endScope()
}

func (c *Checker) checkReturn(s *ast.ReturnStmt) {
	if c.currentFn == nil {
		// Placement is the compiler's error to report; nothing to type.
		if s.Value != nil {
			c.checkExpr(s.Value)
		}
		return
	}

	want := c.currentFn.Return

	if s.Value == nil {
		if want != types.Void {
			c.error(s, "This function must return "+want.String()+".")
		}
		return
	}

	if c.inInit {
		c.error(s, "'init' cannot return a value.")
		c.checkExpr(s.Value)
		return
	}

	got := c.checkExprExpecting(s.Value, want)
	if want == types.Void {
		c.error(s.Value, "This function returns nothing, so it cannot return a value.")
		return
	}
	if !types.AssignableTo(got, want) {
		c.error(s.Value, "Cannot return "+got.String()+" from a function returning "+want.String()+".")
	}
}

func (c *Checker) checkClassStmt(s *ast.ClassStmt) {
	class := c.classes[s.Name.Lexeme]
	if class == nil {
		return // already reported during collect
	}

	c.checkInheritedFields(s, class)
	c.checkOverrides(s, class)

	previousClass := c.currentClass
	c.currentClass = class
	defer func() { c.currentClass = previousClass }()

	for _, method := range s.Methods {
		if method.Name.Lexeme == "init" {
			c.checkInit(method, class)
			continue
		}
		sig, ok := class.LookupMethod(method.Name.Lexeme)
		if !ok {
			continue
		}
		c.checkFunctionBody(method.Params, sig, method.Body)
	}
}

// checkInheritedFields rejects a subclass redeclaring a field it already has.
// Allowing it would give one instance two fields of the same name, and Phase 3
// lays out slots by name.
func (c *Checker) checkInheritedFields(s *ast.ClassStmt, class *types.Class) {
	if class.Super == nil {
		return
	}
	for _, field := range s.Fields {
		if _, inherited := class.Super.LookupField(field.Name.Lexeme); inherited {
			c.errorAt(field.Name, "Field '"+field.Name.Lexeme+
				"' is already declared on superclass '"+class.Super.Name+"'.")
		}
	}
}

// checkOverrides requires an overriding method to have exactly the signature it
// replaces. Invariance is the simple sound rule; loosening it is additive.
// init is exempt: constructors are not virtual.
func (c *Checker) checkOverrides(s *ast.ClassStmt, class *types.Class) {
	if class.Super == nil {
		return
	}
	for _, method := range s.Methods {
		name := method.Name.Lexeme
		if name == "init" {
			continue
		}
		inherited, ok := class.Super.LookupMethod(name)
		if !ok {
			continue
		}
		own, ok := class.Methods[name]
		if !ok {
			continue
		}
		if !types.Equal(own, inherited) {
			c.errorAt(method.Name, "Method '"+name+"' does not match the one it overrides: "+
				"declared "+own.String()+", superclass declares "+inherited.String()+".")
		}
	}
}

func (c *Checker) checkInit(method *ast.FunctionStmt, class *types.Class) {
	sig := &types.Function{Params: class.Constructor().Params, Return: types.Void}

	previousFn, previousInit := c.currentFn, c.inInit
	c.currentFn, c.inInit = sig, true
	defer func() { c.currentFn, c.inInit = previousFn, previousInit }()

	c.beginScope()
	for i, p := range method.Params {
		var t types.Type = types.Invalid
		if i < len(sig.Params) {
			t = sig.Params[i]
		}
		c.declareName(p.Name, "parameter", t, false)
	}
	if method.Body != nil {
		c.checkStatements(method.Body.Statements)
	}
	c.endScope()
}
