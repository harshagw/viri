package internal

type Expr interface{
	Accept(visitor Visitor) interface{}
}

type Visitor interface{
	visitBinaryExp(binaryExp *BinaryExp) interface{}
	visitGrouping(grouping *Grouping) interface{}
	visitLiteral(literal *Literal) interface{}
	visitUnary(unary *Unary) interface{}
}

type BinaryExp struct {
	Left Expr
	Right Expr
	Operator Token
}

func (binaryExp *BinaryExp) Accept(visitor Visitor) interface{} {
	return visitor.visitBinaryExp(binaryExp)
}

type Grouping struct {
	Expr Expr
}


func (grouping *Grouping) Accept(visitor Visitor) interface{} {
	return visitor.visitGrouping(grouping)
}

type Literal struct {
	Value interface{}
}

func (literal *Literal) Accept(visitor Visitor) interface{} {
	return visitor.visitLiteral(literal)
}

type Unary struct {
	Operator Token
	Expr Expr
}

func (unary *Unary) Accept(visitor Visitor) interface{} {
	return visitor.visitUnary(unary)
}
