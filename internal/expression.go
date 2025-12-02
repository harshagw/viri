package internal

type Expr interface{
	Accept(visitor ExprVisitor) (interface{}, error)
}

type ExprVisitor interface{
	visitBinaryExp(binaryExp *BinaryExp) (interface{}, error)
	visitGrouping(grouping *Grouping) (interface{}, error)
	visitLiteral(literal *Literal) (interface{}, error)
	visitUnary(unary *Unary) (interface{}, error)
}


type BinaryExp struct {
	Left Expr
	Right Expr
	Operator Token
}

func (binaryExp *BinaryExp) Accept(visitor ExprVisitor) (interface{}, error) {
	return visitor.visitBinaryExp(binaryExp)
}

type Grouping struct {
	Expr Expr
}


func (grouping *Grouping) Accept(visitor ExprVisitor) (interface{}, error) {
	return visitor.visitGrouping(grouping)
}

type Literal struct {
	Value interface{}
}

func (literal *Literal) Accept(visitor ExprVisitor) (interface{}, error) {
	return visitor.visitLiteral(literal)
}

type Unary struct {
	Operator Token
	Expr Expr
}

func (unary *Unary) Accept(visitor ExprVisitor) (interface{}, error) {
	return visitor.visitUnary(unary)
}
