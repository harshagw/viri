package internal

import (
	"fmt"
	"strings"
)

type AstPrinter struct {
	viri *Viri
}

func NewAstPrinter(viri *Viri) *AstPrinter{
	return &AstPrinter{viri: viri}
}

func (astPrinter *AstPrinter) Print(expr Expr) (string, error) {
	result, err := expr.Accept(astPrinter)
	if err != nil {
		return "", err
	}
	return result.(string), nil
	
}

func (astPrinter *AstPrinter) PrintTree(expr Expr) string {
	var builder strings.Builder
	astPrinter.printTree(expr, &builder, "", true)
	return builder.String()
}

func (astPrinter *AstPrinter) printTree(expr Expr, builder *strings.Builder, prefix string, isLast bool) {
	switch e := expr.(type) {
	case *BinaryExp:
		builder.WriteString(prefix)
		if isLast {
			builder.WriteString("└── ")
		} else {
			builder.WriteString("├── ")
		}
		builder.WriteString(fmt.Sprintf("%s\n", e.Operator.Lexeme))
		
		newPrefix := prefix
		if isLast {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}
		
		astPrinter.printTree(e.Left, builder, newPrefix, false)
		astPrinter.printTree(e.Right, builder, newPrefix, true)
		
	case *Unary:
		builder.WriteString(prefix)
		if isLast {
			builder.WriteString("└── ")
		} else {
			builder.WriteString("├── ")
		}
		builder.WriteString(fmt.Sprintf("%s\n", e.Operator.Lexeme))
		
		newPrefix := prefix
		if isLast {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}
		
		astPrinter.printTree(e.Expr, builder, newPrefix, true)
		
	case *Grouping:
		builder.WriteString(prefix)
		if isLast {
			builder.WriteString("└── (group)\n")
		} else {
			builder.WriteString("├── (group)\n")
		}
		
		newPrefix := prefix
		if isLast {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}
		
		astPrinter.printTree(e.Expr, builder, newPrefix, true)
		
	case *Literal:
		builder.WriteString(prefix)
		if isLast {
			builder.WriteString(fmt.Sprintf("└── %v\n", e.Value))
		} else {
			builder.WriteString(fmt.Sprintf("├── %v\n", e.Value))
		}
	}
}

func (astPrinter *AstPrinter) visitBinaryExp(binaryExp *BinaryExp) (interface{}, error) {
	left, err := binaryExp.Left.Accept(astPrinter)
	if err != nil {
		return nil, err
	}
	right, err := binaryExp.Right.Accept(astPrinter)
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("(%s %s %s)",
		left,
		binaryExp.Operator.Lexeme,
		right,
	), nil
}

func (astPrinter *AstPrinter) visitGrouping(grouping *Grouping) (interface{}, error) {
	expr, err := grouping.Expr.Accept(astPrinter)
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("(%s)", expr), nil
}

func (astPrinter *AstPrinter) visitLiteral(literal *Literal) (interface{}, error) {
	return fmt.Sprintf("%v", literal.Value), nil
}

func (astPrinter *AstPrinter) visitUnary(unary *Unary) (interface{}, error) {
	expr, err := unary.Expr.Accept(astPrinter)
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("(%s %s)", unary.Operator.Lexeme, expr), nil
}

// Fix me - use the variable name to access the value stored

func (astPrinter *AstPrinter) visitVariable(variable *Variable) (interface{}, error) {
	return fmt.Sprintf("Assignment(%s)", variable.Name.Lexeme), nil
}

func (astPrinter *AstPrinter) visitAssignment(assignment *Assignment) (interface{}, error) {
	value, err := assignment.Value.Accept(astPrinter)
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("%s = %s", assignment.Name.Lexeme, value), nil
}

func (astPrinter *AstPrinter) visitLogical(logical *Logical) (interface{}, error) {
	left, err := logical.Left.Accept(astPrinter)
	if err != nil {
		return nil, err
	}
	right, err := logical.Right.Accept(astPrinter)
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("(%s %s %s)", left, logical.Operator.Lexeme, right), nil
}