package internal

import (
	"fmt"
)

type Stmt interface {
	Accept(visitor StmtVisitor) (interface{}, error)
}

type StmtVisitor interface{
	visitExprStmt(exprStmt *ExprStmt) (interface{}, error)
	visitPrintStmt(printStmt *PrintStmt) (interface{}, error)
	visitVarDeclStmt(varDeclStmt *VarDeclStmt) (interface{}, error)
	visitBlock(block *Block) (interface{}, error)
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

func (printStmt *PrintStmt) Accept(visitor StmtVisitor) (interface{}, error) {
	return visitor.visitPrintStmt(printStmt)
}

func (ps *PrintStmt) Print(value interface{}) error {
	// based on the type print the value
	switch value.(type) {
	case string:
		fmt.Println(value)
	case int:
		fmt.Println(value)
	case bool:
		fmt.Println(value)
	case float64:
		fmt.Println(value)	
	default:
		return fmt.Errorf("print doesn't support the expression - %T", value)
	}
	return nil
}

type VarDeclStmt struct {
	tokenName string
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