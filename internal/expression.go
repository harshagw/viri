package internal


type ExprType string

const (
	BINARY_EXP ExprType = "BINARY_EXP"
	GROUPING ExprType = "GROUPING"
	LITERAL ExprType = "LITERAL"
	UNARY ExprType = "UNARY"
	VARIABLE ExprType = "VARIABLE"
	ASSIGNMENT ExprType = "ASSIGNMENT"
)

type Expr interface{
	Accept(visitor ExprVisitor) (interface{}, error)
	Type() ExprType
}

type ExprVisitor interface{
	visitBinaryExp(binaryExp *BinaryExp) (interface{}, error)
	visitGrouping(grouping *Grouping) (interface{}, error)
	visitLiteral(literal *Literal) (interface{}, error)
	visitUnary(unary *Unary) (interface{}, error)
	visitVariable(variable *Variable) (interface{}, error)
	visitAssignment(assignment *Assignment) (interface{}, error)
}

type BinaryExp struct {
	Left Expr
	Right Expr
	Operator Token
}

func (binaryExp *BinaryExp) Accept(visitor ExprVisitor) (interface{}, error) {
	return visitor.visitBinaryExp(binaryExp)
}

func (binaryExp *BinaryExp) Type() ExprType {
	return BINARY_EXP
}

type Grouping struct {
	Expr Expr
}

func (grouping *Grouping) Type() ExprType {
	return GROUPING
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

func (literal *Literal) Type() ExprType {
	return LITERAL
}

type Unary struct {
	Operator Token
	Expr Expr
}

func (unary *Unary) Accept(visitor ExprVisitor) (interface{}, error) {
	return visitor.visitUnary(unary)
}

func (unary *Unary) Type() ExprType {
	return UNARY
}

type Variable struct {
	Name Token
}

func (variable *Variable) Accept(visitor ExprVisitor) (interface{}, error) {
	return visitor.visitVariable(variable)
}

func (variable *Variable) Type() ExprType {
	return VARIABLE
}

type Assignment struct {
	Name Token
	Value Expr
}

func (assignment *Assignment) Accept(visitor ExprVisitor) (interface{}, error) {
	return visitor.visitAssignment(assignment)
}

func (assignment *Assignment) Type() ExprType {
	return ASSIGNMENT
}