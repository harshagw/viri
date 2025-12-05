package internal

import "errors"

type VariableInfo struct {
	defined bool
	used    bool
	token   Token
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
)

type Resolver struct {
	viri                *Viri
	interpreter         *Interpreter
	currentFunctionType FunctionType
	currentClassType    ClassType
	scopes              []map[string]*VariableInfo
}

func NewResolver(viri *Viri, interpreter *Interpreter) *Resolver {
	return &Resolver{
		viri:                viri,
		interpreter:         interpreter,
		scopes:              []map[string]*VariableInfo{},
		currentFunctionType: FunctionTypeNone,
		currentClassType:    ClassTypeNone,
	}
}

func (r *Resolver) Resolve(stmts []Stmt) {
	for _, stmt := range stmts {
		r.resolveStmt(stmt)
	}
}

func (r *Resolver) resolveStmt(stmt Stmt) {
	stmt.Accept(r)
}

func (r *Resolver) resolveExpr(expr Expr) {
	expr.Accept(r)
}

func (r *Resolver) visitBlock(stmt *Block) (interface{}, error) {
	r.beginScope()
	for _, statement := range stmt.statements {
		r.resolveStmt(statement)
	}
	r.endScope()
	return nil, nil
}

func (r *Resolver) visitVarDeclStmt(stmt *VarDeclStmt) (interface{}, error) {
	r.declare(stmt.token)
	if stmt.initializer != nil {
		r.resolveExpr(stmt.initializer)
	}
	r.define(stmt.token)
	return nil, nil
}

func (r *Resolver) visitVariable(variable *Variable) (interface{}, error) {
	if len(r.scopes) > 0 {
		if info, ok := r.scopes[len(r.scopes)-1][variable.Name.Lexeme]; ok && !info.defined {
			r.viri.Error(variable.Name, "Can't read local variable in its own initializer.")
			return nil, nil
		}
	}
	r.resolveLocal(variable, variable.Name.Lexeme)
	r.markVariableUsed(variable.Name.Lexeme)
	return nil, nil
}

func (r *Resolver) visitAssignment(assignment *Assignment) (interface{}, error) {
	r.resolveExpr(assignment.Value)
	r.resolveLocal(assignment, assignment.Name.Lexeme)
	r.markVariableUsed(assignment.Name.Lexeme)
	return nil, nil
}

func (r *Resolver) resolveLocal(expr Expr, name string) {
	for i := len(r.scopes) - 1; i >= 0; i-- {
		if _, ok := r.scopes[i][name]; ok {
			r.interpreter.resolve(expr, len(r.scopes)-i-1)
			return
		}
	}
}

func (r *Resolver) visitFunction(function *Function) (interface{}, error) {
	r.declare(function.token)
	r.define(function.token)

	newFunctionType := FunctionTypeFunction
	if function.token.Lexeme == "init" {
		newFunctionType = FunctionTypeInitializer
	}
	r.resolveFunction(function, newFunctionType)
	return nil, nil
}

func (r *Resolver) visitClass(class *Class) (interface{}, error) {
	r.declare(class.name)
	r.define(class.name)
	previousClassType := r.currentClassType
	r.currentClassType = ClassTypeClass
	r.beginScope()
	thisToken := NewToken(THIS, "this", nil, class.name.Line)
	r.declare(thisToken)
	r.define(thisToken)
	for _, method := range class.methods {
		r.resolveFunction(method, FunctionTypeMethod)
	}
	r.endScope()
	r.currentClassType = previousClassType
	return nil, nil
}

func (r *Resolver) resolveFunction(function *Function, functionType FunctionType) {
	previousFunctionType := r.currentFunctionType
	r.currentFunctionType = functionType
	r.beginScope()
	for _, param := range function.parameters {
		r.declare(param)
		r.define(param)
	}
	// Resolve block statements directly without creating a new scope
	// Function body executes in the same environment as parameters
	for _, statement := range function.body.statements {
		r.resolveStmt(statement)
	}
	r.endScope()
	r.currentFunctionType = previousFunctionType
}

func (r *Resolver) visitExprStmt(exprStmt *ExprStmt) (interface{}, error) {
	r.resolveExpr(exprStmt.Expr)
	return nil, nil
}

func (r *Resolver) visitPrintStmt(printStmt *PrintStmt) (interface{}, error) {
	r.resolveExpr(printStmt.Expr)
	return nil, nil
}

func (r *Resolver) visitIfStmt(ifStmt *IfStmt) (interface{}, error) {
	r.resolveExpr(ifStmt.condition)
	r.resolveStmt(ifStmt.ifBranch)
	if ifStmt.elseBranch != nil {
		r.resolveStmt(ifStmt.elseBranch)
	}
	return nil, nil
}

func (r *Resolver) visitWhileStmt(whileStmt *WhileStmt) (interface{}, error) {
	r.resolveExpr(whileStmt.condition)
	r.resolveStmt(whileStmt.body)
	return nil, nil
}

func (r *Resolver) visitBreakStmt(breakStmt *BreakStmt) (interface{}, error) {
	return nil, nil
}

func (r *Resolver) visitReturnStmt(returnStmt *ReturnStmt) (interface{}, error) {
	if r.currentFunctionType == FunctionTypeNone {
		r.viri.Error(returnStmt.keyword, "Can't return from top-level code.")
	}

	if r.currentFunctionType == FunctionTypeInitializer && returnStmt.value != nil {
		r.viri.Error(returnStmt.keyword, "Can't return a value from an initializer.")
	}

	if returnStmt.value != nil {
		r.resolveExpr(returnStmt.value)
	}
	return nil, nil
}

func (r *Resolver) visitBinaryExp(binaryExp *BinaryExp) (interface{}, error) {
	r.resolveExpr(binaryExp.Left)
	r.resolveExpr(binaryExp.Right)
	return nil, nil
}

func (r *Resolver) visitGrouping(grouping *Grouping) (interface{}, error) {
	r.resolveExpr(grouping.Expr)
	return nil, nil
}

func (r *Resolver) visitLiteral(literal *Literal) (interface{}, error) {
	return nil, nil
}

func (r *Resolver) visitUnary(unary *Unary) (interface{}, error) {
	r.resolveExpr(unary.Expr)
	return nil, nil
}

func (r *Resolver) visitCall(call *Call) (interface{}, error) {
	r.resolveExpr(call.callee)
	for _, argument := range call.arguments {
		r.resolveExpr(argument)
	}
	return nil, nil
}

func (r *Resolver) visitGet(get *Get) (interface{}, error) {
	r.resolveExpr(get.object)
	return nil, nil
}

func (r *Resolver) visitSet(set *Set) (interface{}, error) {
	r.resolveExpr(set.value)
	r.resolveExpr(set.object)
	return nil, nil
}

func (r *Resolver) visitLogical(logical *Logical) (interface{}, error) {
	r.resolveExpr(logical.Left)
	r.resolveExpr(logical.Right)
	return nil, nil
}

func (r *Resolver) visitThisExpr(this *This) (interface{}, error) {
	if r.currentClassType == ClassTypeNone {
		r.viri.Error(this.keyword, "Can't use 'this' outside of a class.")
		return nil, errors.New("can't use 'this' outside of a class")
	}
	r.resolveLocal(this, this.keyword.Lexeme)
	return nil, nil
}

// Scope utility functions

// Adds a new scope to the stack
func (r *Resolver) beginScope() {
	r.scopes = append(r.scopes, make(map[string]*VariableInfo))
}

// Remove the top most scope from the stack
func (r *Resolver) endScope() {
	if len(r.scopes) == 0 {
		return
	}

	// Check for unused variables before removing the scope
	scope := r.scopes[len(r.scopes)-1]
	for name, info := range scope {
		if info.defined && !info.used && name != "this" {
			r.viri.Warn(info.token, "Local variable '"+name+"' is declared but never used.")
		}
	}

	r.scopes = r.scopes[:len(r.scopes)-1]
}

// Declare a variable in the current scope
// It means the variable is present but not yet initialzed
func (r *Resolver) declare(token Token) {
	if len(r.scopes) == 0 {
		return
	}

	currentScope := r.scopes[len(r.scopes)-1]

	if _, ok := currentScope[token.Lexeme]; ok {
		r.viri.Error(token, "Cannot declare variable with this name again.")
	}

	currentScope[token.Lexeme] = &VariableInfo{
		defined: false,
		used:    false,
		token:   token,
	}
}

// Define a variable in the current scope
// It means variables is initialized in the current scope
func (r *Resolver) define(token Token) {
	if len(r.scopes) == 0 {
		return
	}

	if info, ok := r.scopes[len(r.scopes)-1][token.Lexeme]; ok {
		info.defined = true
	} else {
		// This shouldn't happen, but handle it gracefully
		r.scopes[len(r.scopes)-1][token.Lexeme] = &VariableInfo{
			defined: true,
			used:    false,
			token:   token,
		}
	}
}

// Mark a variable as used in the current scope or any enclosing scope
func (r *Resolver) markVariableUsed(name string) {
	// Search from innermost to outermost scope
	for i := len(r.scopes) - 1; i >= 0; i-- {
		if info, ok := r.scopes[i][name]; ok {
			info.used = true
			return
		}
	}
}
