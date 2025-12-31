package ast

import "github.com/harshagw/viri/internal/token"

// Expr is the expression interface.
type Expr interface {
	Node
	exprNode()
}

type BinaryExpr struct {
	Left     Expr
	Right    Expr
	Operator *token.Token
}

func (*BinaryExpr) exprNode()                       {}
func (e *BinaryExpr) GetPrimaryToken() *token.Token { return e.Operator }

type GroupingExpr struct {
	Expr Expr
}

func (*GroupingExpr) exprNode()                       {}
func (e *GroupingExpr) GetPrimaryToken() *token.Token { return e.Expr.GetPrimaryToken() }

type LiteralExpr struct {
	Value interface{}
}

func (*LiteralExpr) exprNode()                       {}
func (e *LiteralExpr) GetPrimaryToken() *token.Token { return nil }

type UnaryExpr struct {
	Operator *token.Token
	Expr     Expr
}

func (*UnaryExpr) exprNode()                       {}
func (e *UnaryExpr) GetPrimaryToken() *token.Token { return e.Operator }

type VariableExpr struct {
	Name *token.Token
}

func (*VariableExpr) exprNode()                       {}
func (e *VariableExpr) GetPrimaryToken() *token.Token { return e.Name }

type AssignExpr struct {
	Name  *token.Token
	Value Expr
}

func (*AssignExpr) exprNode()                       {}
func (e *AssignExpr) GetPrimaryToken() *token.Token { return e.Name }

type LogicalExpr struct {
	Left     Expr
	Operator *token.Token
	Right    Expr
}

func (*LogicalExpr) exprNode()                       {}
func (e *LogicalExpr) GetPrimaryToken() *token.Token { return e.Operator }

type CallExpr struct {
	Callee       Expr
	Arguments    []Expr
	ClosingParen *token.Token
}

func (*CallExpr) exprNode()                       {}
func (e *CallExpr) GetPrimaryToken() *token.Token { return e.ClosingParen }

type GetExpr struct {
	Object Expr
	Name   *token.Token
}

func (*GetExpr) exprNode()                       {}
func (e *GetExpr) GetPrimaryToken() *token.Token { return e.Name }

type SetExpr struct {
	Object Expr
	Name   *token.Token
	Value  Expr
}

func (*SetExpr) exprNode()                       {}
func (e *SetExpr) GetPrimaryToken() *token.Token { return e.Name }

type ThisExpr struct {
	Keyword *token.Token
}

func (*ThisExpr) exprNode()                       {}
func (e *ThisExpr) GetPrimaryToken() *token.Token { return e.Keyword }

type SuperExpr struct {
	Keyword *token.Token
	Method  *token.Token
}

func (*SuperExpr) exprNode()                       {}
func (e *SuperExpr) GetPrimaryToken() *token.Token { return e.Keyword }

type ArrayLiteralExpr struct {
	Elements []Expr
}

func (*ArrayLiteralExpr) exprNode() {}
func (e *ArrayLiteralExpr) GetPrimaryToken() *token.Token {
	if len(e.Elements) > 0 {
		return e.Elements[0].GetPrimaryToken()
	}
	return nil
}

type HashPair struct {
	Key   Expr
	Value Expr
}

type HashLiteralExpr struct {
	Pairs []HashPair
	Brace *token.Token
}

func (*HashLiteralExpr) exprNode()                       {}
func (e *HashLiteralExpr) GetPrimaryToken() *token.Token { return e.Brace }

type IndexExpr struct {
	Object  Expr
	Index   Expr
	Bracket *token.Token
}

func (*IndexExpr) exprNode()                       {}
func (e *IndexExpr) GetPrimaryToken() *token.Token { return e.Bracket }

type SetIndexExpr struct {
	Object  Expr
	Index   Expr
	Value   Expr
	Bracket *token.Token
}

func (*SetIndexExpr) exprNode()                       {}
func (e *SetIndexExpr) GetPrimaryToken() *token.Token { return e.Bracket }

type FunctionExpr struct {
	Params []*token.Token
	Body   *BlockStmt
}

func (*FunctionExpr) exprNode() {}
func (e *FunctionExpr) GetPrimaryToken() *token.Token {
	if len(e.Params) > 0 {
		return e.Params[0]
	}
	if e.Body != nil {
		return e.Body.GetPrimaryToken()
	}
	return nil
}
