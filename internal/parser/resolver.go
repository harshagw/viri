package parser

import (
	"errors"
	"fmt"

	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/token"
)

var ErrResolve = errors.New("resolve error")

type VariableInfo struct {
	defined bool
	used    bool
	token   token.Token
}

type FunctionType int

const (
	FunctionTypeNone FunctionType = iota
	FunctionTypeFunction
	FunctionTypeMethod
	FunctionTypeInitializer
)

type ClassType int

const (
	ClassTypeNone ClassType = iota
	ClassTypeClass
	ClassTypeSubclass
)

type Resolver struct {
	diagnosticHandler objects.DiagnosticHandler
	currentFunction   FunctionType
	currentClass      ClassType
	loopDepth         int
	scopes            []map[string]*VariableInfo
	locals            map[ast.Expr]int
	hadError          bool
}

func NewResolver(diagnosticHandler objects.DiagnosticHandler) *Resolver {
	return &Resolver{
		diagnosticHandler: diagnosticHandler,
		scopes:            []map[string]*VariableInfo{},
		currentFunction:   FunctionTypeNone,
		currentClass:      ClassTypeNone,
		locals:            make(map[ast.Expr]int),
	}
}

func (r *Resolver) Resolve(stmts []ast.Stmt) (map[ast.Expr]int, error) {
	r.beginScope()
	for _, stmt := range stmts {
		r.resolveStmt(stmt)
	}
	r.endScope()
	if r.hadError {
		return r.locals, ErrResolve
	}
	return r.locals, nil
}

func (r *Resolver) resolveStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		r.visitBlock(s)
	case *ast.VarDeclStmt:
		r.visitVarDeclStmt(s)
	case *ast.FunctionStmt:
		r.visitFunction(s)
	case *ast.ClassStmt:
		r.visitClass(s)
	case *ast.ExprStmt:
		r.resolveExpr(s.Expr)
	case *ast.PrintStmt:
		r.resolveExpr(s.Expr)
	case *ast.IfStmt:
		r.visitIfStmt(s)
	case *ast.WhileStmt:
		r.visitWhileStmt(s)
	case *ast.BreakStmt:
		if r.loopDepth == 0 {
			r.reportError(s.Keyword, "break statement must be inside a loop.")
		}
	case *ast.ReturnStmt:
		r.visitReturnStmt(s)
	default:
		panic(fmt.Sprintf("unknown statement type: %T", s))
	}
}

func (r *Resolver) resolveExpr(expr ast.Expr) {
	switch e := expr.(type) {
	case *ast.BinaryExpr:
		r.resolveExpr(e.Left)
		r.resolveExpr(e.Right)
	case *ast.GroupingExpr:
		r.resolveExpr(e.Expr)
	case *ast.LiteralExpr:
	case *ast.ArrayLiteralExpr:
		for _, el := range e.Elements {
			r.resolveExpr(el)
		}
	case *ast.HashLiteralExpr:
		for _, pair := range e.Pairs {
			r.resolveExpr(pair.Key)
			r.resolveExpr(pair.Value)
		}
	case *ast.UnaryExpr:
		r.resolveExpr(e.Expr)
	case *ast.VariableExpr:
		r.visitVariable(e)
	case *ast.AssignExpr:
		r.visitAssignment(e)
	case *ast.SetIndexExpr:
		r.resolveExpr(e.Object)
		r.resolveExpr(e.Index)
		r.resolveExpr(e.Value)
	case *ast.LogicalExpr:
		r.resolveExpr(e.Left)
		r.resolveExpr(e.Right)
	case *ast.CallExpr:
		r.visitCall(e)
	case *ast.GetExpr:
		r.resolveExpr(e.Object)
	case *ast.SetExpr:
		r.resolveExpr(e.Value)
		r.resolveExpr(e.Object)
	case *ast.ThisExpr:
		r.visitThisExpr(e)
	case *ast.SuperExpr:
		r.visitSuperExpr(e)
	case *ast.IndexExpr:
		r.resolveExpr(e.Object)
		r.resolveExpr(e.Index)
	default:
		panic(fmt.Sprintf("unknown expression type: %T", e))
	}
}

func (r *Resolver) visitBlock(stmt *ast.BlockStmt) {
	r.beginScope()
	for _, statement := range stmt.Statements {
		r.resolveStmt(statement)
	}
	r.endScope()
}

func (r *Resolver) visitVarDeclStmt(stmt *ast.VarDeclStmt) {
	r.declare(stmt.Name)
	if stmt.Initializer != nil {
		r.resolveExpr(stmt.Initializer)
	}
	r.define(stmt.Name)
}

func (r *Resolver) visitVariable(variable *ast.VariableExpr) {
	if len(r.scopes) > 0 {
		if info, ok := r.scopes[len(r.scopes)-1][variable.Name.Lexeme]; ok && !info.defined {
			r.reportError(variable.Name, "Can't read local variable in its own initializer.")
			return
		}
	}
	r.resolveLocal(variable, variable.Name)
	r.markVariableUsed(variable.Name)
}

func (r *Resolver) visitAssignment(assignment *ast.AssignExpr) {
	r.resolveExpr(assignment.Value)
	r.resolveLocal(assignment, assignment.Name)
	r.markVariableUsed(assignment.Name)
}

func (r *Resolver) resolveLocal(expr ast.Expr, name token.Token) {
	for i := len(r.scopes) - 1; i >= 0; i-- {
		if _, ok := r.scopes[i][name.Lexeme]; ok {
			r.locals[expr] = len(r.scopes) - i - 1
			return
		}
	}
}

func (r *Resolver) visitFunction(function *ast.FunctionStmt) {
	r.declare(function.Name)
	r.define(function.Name)

	newFunctionType := FunctionTypeFunction
	if r.currentClass != ClassTypeNone && function.Name.Lexeme == "init" {
		newFunctionType = FunctionTypeInitializer
	}
	r.resolveFunction(function, newFunctionType)
}

func (r *Resolver) visitClass(class *ast.ClassStmt) {
	r.declare(class.Name)
	r.define(class.Name)

	previousClassType := r.currentClass
	r.currentClass = ClassTypeClass

	if class.SuperClass != nil {
		if class.SuperClass.Name.Lexeme == class.Name.Lexeme {
			r.reportError(class.SuperClass.Name, "A class cannot inherit from itself.")
			return
		}

		r.resolveExpr(class.SuperClass)

		r.beginScope()

		r.currentClass = ClassTypeSubclass

		superToken := token.New(token.SUPER, "super", nil, class.SuperClass.Name.Line)
		r.declare(superToken)
		r.define(superToken)
	}

	r.beginScope()
	thisToken := token.New(token.THIS, "this", nil, class.Name.Line)
	r.declare(thisToken)
	r.define(thisToken)

	for _, method := range class.Methods {
		r.resolveFunction(method, FunctionTypeMethod)
	}

	r.endScope()

	if class.SuperClass != nil {
		r.endScope()
	}

	r.currentClass = previousClassType
}

func (r *Resolver) resolveFunction(function *ast.FunctionStmt, functionType FunctionType) {
	previousFunctionType := r.currentFunction
	r.currentFunction = functionType
	r.beginScope()
	for _, param := range function.Params {
		r.declare(param)
		r.define(param)
	}
	for _, statement := range function.Body.Statements {
		r.resolveStmt(statement)
	}
	r.endScope()
	r.currentFunction = previousFunctionType
}

func (r *Resolver) visitIfStmt(ifStmt *ast.IfStmt) {
	r.resolveExpr(ifStmt.Condition)
	r.resolveStmt(ifStmt.ThenBranch)
	if ifStmt.ElseBranch != nil {
		r.resolveStmt(ifStmt.ElseBranch)
	}
}

func (r *Resolver) visitWhileStmt(whileStmt *ast.WhileStmt) {
	r.resolveExpr(whileStmt.Condition)
	r.loopDepth++
	r.resolveStmt(whileStmt.Body)
	r.loopDepth--
}

func (r *Resolver) visitReturnStmt(returnStmt *ast.ReturnStmt) {
	if r.currentFunction == FunctionTypeNone {
		r.reportError(returnStmt.Keyword, "Can't return from top-level code.")
	}

	if r.currentFunction == FunctionTypeInitializer && returnStmt.Value != nil {
		r.reportError(returnStmt.Keyword, "Can't return a value from an initializer.")
	}

	if returnStmt.Value != nil {
		r.resolveExpr(returnStmt.Value)
	}
}

func (r *Resolver) visitCall(call *ast.CallExpr) {
	r.resolveExpr(call.Callee)
	for _, argument := range call.Arguments {
		r.resolveExpr(argument)
	}
}

func (r *Resolver) visitThisExpr(this *ast.ThisExpr) {
	if r.currentClass == ClassTypeNone {
		r.reportError(this.Keyword, "Can't use 'this' outside of a class.")
		return
	}
	r.resolveLocal(this, this.Keyword)
}

func (r *Resolver) visitSuperExpr(super *ast.SuperExpr) {
	if r.currentClass == ClassTypeNone {
		r.reportError(super.Keyword, "Can't use 'super' outside of a class.")
		return
	}
	if r.currentClass != ClassTypeSubclass {
		r.reportError(super.Keyword, "Can't use 'super' in a class with no superclass.")
		return
	}
	r.resolveLocal(super, super.Keyword)
}

// Scope utilities

func (r *Resolver) beginScope() {
	r.scopes = append(r.scopes, make(map[string]*VariableInfo))
}

func (r *Resolver) endScope() {
	if len(r.scopes) == 0 {
		return
	}

	scope := r.scopes[len(r.scopes)-1]
	for name, info := range scope {
		if info.defined && !info.used && name != "this" && name != "super" {
			r.reportWarn(info.token, "Local variable '"+name+"' is declared but never used.")
		}
	}

	r.scopes = r.scopes[:len(r.scopes)-1]
}

func (r *Resolver) declare(tok token.Token) {
	if len(r.scopes) == 0 {
		return
	}

	currentScope := r.scopes[len(r.scopes)-1]

	if _, ok := currentScope[tok.Lexeme]; ok {
		r.reportError(tok, "Cannot declare variable with this name again.")
	}

	currentScope[tok.Lexeme] = &VariableInfo{
		defined: false,
		used:    false,
		token:   tok,
	}
}

func (r *Resolver) define(tok token.Token) {
	if len(r.scopes) == 0 {
		return
	}

	if info, ok := r.scopes[len(r.scopes)-1][tok.Lexeme]; ok {
		info.defined = true
	} else {
		r.scopes[len(r.scopes)-1][tok.Lexeme] = &VariableInfo{
			defined: true,
			used:    false,
			token:   tok,
		}
	}
}

func (r *Resolver) markVariableUsed(tok token.Token) {
	for i := len(r.scopes) - 1; i >= 0; i-- {
		if info, ok := r.scopes[i][tok.Lexeme]; ok {
			info.used = true
			return
		}
	}
}

func (r *Resolver) reportError(tok token.Token, message string) {
	r.hadError = true
	if r.diagnosticHandler != nil {
		r.diagnosticHandler.Error(tok, message)
	}
}

func (r *Resolver) reportWarn(tok token.Token, message string) {
	if r.diagnosticHandler != nil {
		r.diagnosticHandler.Warn(tok, message)
	}
}
