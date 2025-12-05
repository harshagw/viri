package internal

import (
	"fmt"
)

type Stmt interface {
	Accept(visitor StmtVisitor) (interface{}, error)
}

type StmtVisitor interface {
	visitExprStmt(exprStmt *ExprStmt) (interface{}, error)
	visitPrintStmt(printStmt *PrintStmt) (interface{}, error)
	visitVarDeclStmt(varDeclStmt *VarDeclStmt) (interface{}, error)
	visitBlock(block *Block) (interface{}, error)
	visitIfStmt(ifStmt *IfStmt) (interface{}, error)
	visitWhileStmt(whileStmt *WhileStmt) (interface{}, error)
	visitBreakStmt(breakStmt *BreakStmt) (interface{}, error)
	visitFunction(function *Function) (interface{}, error)
	visitReturnStmt(returnStmt *ReturnStmt) (interface{}, error)
	visitClass(class *Class) (interface{}, error)
}

type ExprStmt struct {
	Expr Expr
}

func (exprStmt *ExprStmt) Accept(visitor StmtVisitor) (interface{}, error) {
	return visitor.visitExprStmt(exprStmt)
}

type PrintStmt struct {
	Expr Expr
}

type Printable interface {
	ToString() string
}

func (printStmt *PrintStmt) Accept(visitor StmtVisitor) (interface{}, error) {
	return visitor.visitPrintStmt(printStmt)
}

func (ps *PrintStmt) Print(value interface{}) error {
	// based on the type print the value
	switch value := value.(type) {
	case string, int, int64, bool, float64, Callable:
		fmt.Println(value)
	case Printable:
		fmt.Println(value.ToString())
	default:
		return fmt.Errorf("print doesn't support the expression - %T", value)
	}
	return nil
}

type VarDeclStmt struct {
	token       Token
	initializer Expr
}

func (varDeclStmt *VarDeclStmt) Accept(visitor StmtVisitor) (interface{}, error) {
	return visitor.visitVarDeclStmt(varDeclStmt)
}

type Block struct {
	statements []Stmt
}

func (block *Block) Accept(visitor StmtVisitor) (interface{}, error) {
	return visitor.visitBlock(block)
}

type IfStmt struct {
	condition  Expr
	ifBranch   Stmt
	elseBranch Stmt
}

func (ifStmt *IfStmt) Accept(visitor StmtVisitor) (interface{}, error) {
	return visitor.visitIfStmt(ifStmt)
}

type WhileStmt struct {
	condition Expr
	body      Stmt
}

func (whileStmt *WhileStmt) Accept(visitor StmtVisitor) (interface{}, error) {
	return visitor.visitWhileStmt(whileStmt)
}

type BreakStmt struct {
	keyword Token
}

func (breakStmt *BreakStmt) Accept(visitor StmtVisitor) (interface{}, error) {
	return visitor.visitBreakStmt(breakStmt)
}

type Function struct {
	token      Token
	parameters []Token
	body       *Block
}

func (function *Function) Accept(visitor StmtVisitor) (interface{}, error) {
	return visitor.visitFunction(function)
}

type ReturnStmt struct {
	keyword Token
	value   Expr
}

func (returnStmt *ReturnStmt) Accept(visitor StmtVisitor) (interface{}, error) {
	return visitor.visitReturnStmt(returnStmt)
}

type Class struct {
	name    Token
	methods []*Function
}

func (class *Class) Accept(visitor StmtVisitor) (interface{}, error) {
	return visitor.visitClass(class)
}
