package ast

import "github.com/harshagw/viri/internal/token"

// Expr is the root expression interface.
type Expr interface {
	exprNode()
}

type BinaryExpr struct {
	Left     Expr
	Right    Expr
	Operator *token.Token
}

func (*BinaryExpr) exprNode() {}

type GroupingExpr struct {
	Expr Expr
}

func (*GroupingExpr) exprNode() {}

type LiteralExpr struct {
	Value interface{}
}

func (*LiteralExpr) exprNode() {}

type UnaryExpr struct {
	Operator *token.Token
	Expr     Expr
}

func (*UnaryExpr) exprNode() {}

type VariableExpr struct {
	Name *token.Token
}

func (*VariableExpr) exprNode() {}

type AssignExpr struct {
	Name  *token.Token
	Value Expr
}

func (*AssignExpr) exprNode() {}

type LogicalExpr struct {
	Left     Expr
	Operator *token.Token
	Right    Expr
}

func (*LogicalExpr) exprNode() {}

type CallExpr struct {
	Callee       Expr
	Arguments    []Expr
	ClosingParen *token.Token
}

func (*CallExpr) exprNode() {}

type GetExpr struct {
	Object Expr
	Name   *token.Token
}

func (*GetExpr) exprNode() {}

type SetExpr struct {
	Object Expr
	Name   *token.Token
	Value  Expr
}

func (*SetExpr) exprNode() {}

type ThisExpr struct {
	Keyword *token.Token
}

func (*ThisExpr) exprNode() {}

type SuperExpr struct {
	Keyword *token.Token
	Method  *token.Token
}

func (*SuperExpr) exprNode() {}

type ArrayLiteralExpr struct {
	Elements []Expr
}

func (*ArrayLiteralExpr) exprNode() {}

type HashPair struct {
	Key   Expr
	Value Expr
}

type HashLiteralExpr struct {
	Pairs []HashPair
	Brace *token.Token
}

func (*HashLiteralExpr) exprNode() {}

type IndexExpr struct {
	Object  Expr
	Index   Expr
	Bracket *token.Token
}

func (*IndexExpr) exprNode() {}

type SetIndexExpr struct {
	Object  Expr
	Index   Expr
	Value   Expr
	Bracket *token.Token
}

func (*SetIndexExpr) exprNode() {}
