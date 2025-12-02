package internal

import (
	"fmt"
	"strings"
)

type AstPrinter struct {
	
}

func NewAstPrinter() *AstPrinter{
	return &AstPrinter{}
}

func (astPrinter *AstPrinter) Print(expr Expr) string {
	result := expr.Accept(astPrinter).(string)
	return result
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

func (astPrinter *AstPrinter) visitBinaryExp(binaryExp *BinaryExp) interface{} {
	return fmt.Sprintf("(%s %s %s)",
		binaryExp.Left.Accept(astPrinter),
		binaryExp.Operator.Lexeme,
		binaryExp.Right.Accept(astPrinter),
	)
}

func (astPrinter *AstPrinter) visitGrouping(grouping *Grouping) interface{} {
	return fmt.Sprintf("(%s)", grouping.Expr.Accept(astPrinter))
}

func (astPrinter *AstPrinter) visitLiteral(literal *Literal) interface{} {
	return fmt.Sprintf("%v", literal.Value)
}

func (astPrinter *AstPrinter) visitUnary(unary *Unary) interface{} {
	return fmt.Sprintf("(%s %s)", unary.Operator.Lexeme, unary.Expr.Accept(astPrinter))
}