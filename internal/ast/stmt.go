package ast

import "github.com/harshagw/viri/internal/token"

// Stmt is the root statement interface.
type Stmt interface {
	stmtNode()
}

type ExprStmt struct {
	Expr Expr
}

func (*ExprStmt) stmtNode() {}

type PrintStmt struct {
	Expr Expr
}

func (*PrintStmt) stmtNode() {}

type VarDeclStmt struct {
	Name        token.Token
	Initializer Expr
}

func (*VarDeclStmt) stmtNode() {}

type BlockStmt struct {
	Statements []Stmt
}

func (*BlockStmt) stmtNode() {}

type IfStmt struct {
	Condition  Expr
	ThenBranch Stmt
	ElseBranch Stmt
}

func (*IfStmt) stmtNode() {}

type WhileStmt struct {
	Condition Expr
	Body      Stmt
}

func (*WhileStmt) stmtNode() {}

type BreakStmt struct {
	Keyword token.Token
}

func (*BreakStmt) stmtNode() {}

type FunctionStmt struct {
	Name   token.Token
	Params []token.Token
	Body   *BlockStmt
}

func (*FunctionStmt) stmtNode() {}

type ReturnStmt struct {
	Keyword token.Token
	Value   Expr
}

func (*ReturnStmt) stmtNode() {}

type ClassStmt struct {
	Name    token.Token
	SuperClass *VariableExpr
	Methods []*FunctionStmt
}

func (*ClassStmt) stmtNode() {}
