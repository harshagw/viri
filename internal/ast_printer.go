package internal

import (
	"fmt"
	"strings"
)

type AstPrinter struct {
	builder *strings.Builder
	prefix  string
	isLast  bool
}

func NewAstPrinter() *AstPrinter {
	return &AstPrinter{}
}

func (astPrinter *AstPrinter) PrintStatements(statements []Stmt) string {
	var builder strings.Builder
	astPrinter.builder = &builder
	for i, stmt := range statements {
		astPrinter.prefix = ""
		astPrinter.isLast = i == len(statements)-1
		stmt.Accept(astPrinter)
	}
	return builder.String()
}

func (astPrinter *AstPrinter) PrintExpr(expr Expr) string {
	var builder strings.Builder
	astPrinter.builder = &builder
	astPrinter.prefix = ""
	astPrinter.isLast = true
	expr.Accept(astPrinter)
	return builder.String()
}

// ExprVisitor implementation

func (astPrinter *AstPrinter) visitBinaryExp(binaryExp *BinaryExp) (interface{}, error) {
	astPrinter.writeTreeNode(astPrinter.prefix, astPrinter.isLast, "BinaryExp ("+binaryExp.Operator.Lexeme+")")
	newPrefix := astPrinter.calculateNewPrefix(astPrinter.prefix, astPrinter.isLast)
	astPrinter.printExpr(binaryExp.Left, newPrefix, false)
	astPrinter.printExpr(binaryExp.Right, newPrefix, true)
	return nil, nil
}

func (astPrinter *AstPrinter) visitUnary(unary *Unary) (interface{}, error) {
	astPrinter.writeTreeNode(astPrinter.prefix, astPrinter.isLast, "Unary ("+unary.Operator.Lexeme+")")
	newPrefix := astPrinter.calculateNewPrefix(astPrinter.prefix, astPrinter.isLast)
	astPrinter.printExpr(unary.Expr, newPrefix, true)
	return nil, nil
}

func (astPrinter *AstPrinter) visitGrouping(grouping *Grouping) (interface{}, error) {
	astPrinter.writeTreeNode(astPrinter.prefix, astPrinter.isLast, "Grouping")
	newPrefix := astPrinter.calculateNewPrefix(astPrinter.prefix, astPrinter.isLast)
	astPrinter.printExpr(grouping.Expr, newPrefix, true)
	return nil, nil
}

func (astPrinter *AstPrinter) visitLiteral(literal *Literal) (interface{}, error) {
	astPrinter.writeTreeNode(astPrinter.prefix, astPrinter.isLast, "Literal ("+fmt.Sprintf("%v", literal.Value)+")")
	return nil, nil
}

func (astPrinter *AstPrinter) visitVariable(variable *Variable) (interface{}, error) {
	astPrinter.writeTreeNode(astPrinter.prefix, astPrinter.isLast, "Variable ("+variable.Name.Lexeme+")")
	return nil, nil
}

func (astPrinter *AstPrinter) visitAssignment(assignment *Assignment) (interface{}, error) {
	astPrinter.writeTreeNode(astPrinter.prefix, astPrinter.isLast, "Assignment (=) ")
	newPrefix := astPrinter.calculateNewPrefix(astPrinter.prefix, astPrinter.isLast)
	astPrinter.writeTreeNode(newPrefix, true, assignment.Name.Lexeme)
	astPrinter.printExpr(assignment.Value, newPrefix, true)
	return nil, nil
}

func (astPrinter *AstPrinter) visitLogical(logical *Logical) (interface{}, error) {
	astPrinter.writeTreeNode(astPrinter.prefix, astPrinter.isLast, "Logical ("+logical.Operator.Lexeme+")")
	newPrefix := astPrinter.calculateNewPrefix(astPrinter.prefix, astPrinter.isLast)
	astPrinter.printExpr(logical.Left, newPrefix, false)
	astPrinter.printExpr(logical.Right, newPrefix, true)
	return nil, nil
}

func (astPrinter *AstPrinter) visitCall(call *Call) (interface{}, error) {
	astPrinter.writeTreeNode(astPrinter.prefix, astPrinter.isLast, "Call")
	newPrefix := astPrinter.calculateNewPrefix(astPrinter.prefix, astPrinter.isLast)
	astPrinter.printExpr(call.callee, newPrefix, false)
	astPrinter.writeTreeNode(newPrefix, true, "arguments")
	argumentsPrefix := astPrinter.calculateNewPrefix(newPrefix, true)
	for i, argument := range call.arguments {
		argumentIsLast := i == len(call.arguments)-1
		astPrinter.printExpr(argument, argumentsPrefix, argumentIsLast)
	}
	if len(call.arguments) == 0{
		astPrinter.writeTreeNode(argumentsPrefix, true, "nil")
	}
	return nil, nil
}

// StmtVisitor implementation

func (astPrinter *AstPrinter) visitExprStmt(exprStmt *ExprStmt) (interface{}, error) {
	astPrinter.writeTreeNode(astPrinter.prefix, astPrinter.isLast, "ExprStmt")
	newPrefix := astPrinter.calculateNewPrefix(astPrinter.prefix, astPrinter.isLast)
	astPrinter.printExpr(exprStmt.Expr, newPrefix, true)
	return nil, nil
}

func (astPrinter *AstPrinter) visitPrintStmt(printStmt *PrintStmt) (interface{}, error) {
	astPrinter.writeTreeNode(astPrinter.prefix, astPrinter.isLast, "PrintStmt")
	newPrefix := astPrinter.calculateNewPrefix(astPrinter.prefix, astPrinter.isLast)
	astPrinter.printExpr(printStmt.Expr, newPrefix, true)
	return nil, nil
}

func (astPrinter *AstPrinter) visitVarDeclStmt(varDeclStmt *VarDeclStmt) (interface{}, error) {
	astPrinter.writeTreeNode(astPrinter.prefix, astPrinter.isLast, fmt.Sprintf("VarDeclStmt(%s)", varDeclStmt.tokenName))
	if varDeclStmt.initializer != nil {
		newPrefix := astPrinter.calculateNewPrefix(astPrinter.prefix, astPrinter.isLast)
		astPrinter.printExpr(varDeclStmt.initializer, newPrefix, true)
	}
	return nil, nil
}

func (astPrinter *AstPrinter) visitBlock(block *Block) (interface{}, error) {
	astPrinter.writeTreeNode(astPrinter.prefix, astPrinter.isLast, "Block")
	newPrefix := astPrinter.calculateNewPrefix(astPrinter.prefix, astPrinter.isLast)
	for i, stmt := range block.statements {
		stmtIsLast := i == len(block.statements)-1
		astPrinter.printStmt(stmt, newPrefix, stmtIsLast)
	}
	return nil, nil
}

func (astPrinter *AstPrinter) visitIfStmt(ifStmt *IfStmt) (interface{}, error) {
	astPrinter.writeTreeNode(astPrinter.prefix, astPrinter.isLast, "IfStmt")
	newPrefix := astPrinter.calculateNewPrefix(astPrinter.prefix, astPrinter.isLast)
	astPrinter.writeTreeNode(newPrefix, false, "condition")
	conditionPrefix := astPrinter.calculateNewPrefix(newPrefix, false)
	astPrinter.printExpr(ifStmt.condition, conditionPrefix, true)
	astPrinter.writeTreeNode(newPrefix, ifStmt.elseBranch == nil, "ifBranch")
	astPrinter.printStmt(ifStmt.ifBranch, conditionPrefix, ifStmt.elseBranch == nil)
	if ifStmt.elseBranch != nil {
		astPrinter.writeTreeNode(newPrefix, true, "elseBranch")
		astPrinter.printStmt(ifStmt.elseBranch, conditionPrefix, true)
	}
	return nil, nil
}

func (astPrinter *AstPrinter) visitWhileStmt(whileStmt *WhileStmt) (interface{}, error) {
	astPrinter.writeTreeNode(astPrinter.prefix, astPrinter.isLast, "WhileStmt")
	newPrefix := astPrinter.calculateNewPrefix(astPrinter.prefix, astPrinter.isLast)
	astPrinter.writeTreeNode(newPrefix, false, "condition")
	conditionPrefix := astPrinter.calculateNewPrefix(newPrefix, false)
	astPrinter.printExpr(whileStmt.condition, conditionPrefix, true)
	astPrinter.writeTreeNode(newPrefix, true, "body")
	astPrinter.printStmt(whileStmt.body, conditionPrefix, true)
	return nil, nil
}

func (astPrinter *AstPrinter) visitBreakStmt(breakStmt *BreakStmt) (interface{}, error) {
	astPrinter.writeTreeNode(astPrinter.prefix, astPrinter.isLast, "BreakStmt")
	return nil, nil
}

func (astPrinter *AstPrinter) visitFunction(function *Function) (interface{}, error) {
	astPrinter.writeTreeNode(astPrinter.prefix, astPrinter.isLast, "Function ("+function.token.Lexeme+")")
	newPrefix := astPrinter.calculateNewPrefix(astPrinter.prefix, astPrinter.isLast)
	astPrinter.writeTreeNode(newPrefix, false, "parameters")
	parametersPrefix := astPrinter.calculateNewPrefix(newPrefix, false)
	for i, parameter := range function.parameters {
		astPrinter.writeTreeNode(parametersPrefix, i == len(function.parameters)-1, parameter.Lexeme)
	}
	astPrinter.writeTreeNode(newPrefix, false, "body")
	astPrinter.printStmt(function.body, parametersPrefix, true)
	return nil, nil
}

func (astPrinter *AstPrinter) visitReturnStmt(returnStmt *ReturnStmt) (interface{}, error) {
	astPrinter.writeTreeNode(astPrinter.prefix, astPrinter.isLast, "ReturnStmt")
	newPrefix := astPrinter.calculateNewPrefix(astPrinter.prefix, astPrinter.isLast)
	astPrinter.printExpr(returnStmt.value, newPrefix, true)
	return nil, nil
}

// Helper methods for tree printing

func (astPrinter *AstPrinter) printExpr(expr Expr, prefix string, isLast bool) {
	oldPrefix := astPrinter.prefix
	oldIsLast := astPrinter.isLast
	astPrinter.prefix = prefix
	astPrinter.isLast = isLast
	expr.Accept(astPrinter)
	astPrinter.prefix = oldPrefix
	astPrinter.isLast = oldIsLast
}

func (astPrinter *AstPrinter) printStmt(stmt Stmt, prefix string, isLast bool) {
	oldPrefix := astPrinter.prefix
	oldIsLast := astPrinter.isLast
	astPrinter.prefix = prefix
	astPrinter.isLast = isLast
	stmt.Accept(astPrinter)
	astPrinter.prefix = oldPrefix
	astPrinter.isLast = oldIsLast
}

func (astPrinter *AstPrinter) writeTreeNode(prefix string, isLast bool, label string) {
	astPrinter.builder.WriteString(prefix)
	if isLast {
		astPrinter.builder.WriteString("└── ")
	} else {
		astPrinter.builder.WriteString("├── ")
	}
	astPrinter.builder.WriteString(label)
	astPrinter.builder.WriteString("\n")
}

func (astPrinter *AstPrinter) calculateNewPrefix(prefix string, isLast bool) string {
	if isLast {
		return prefix + "    "
	}
	return prefix + "│   "
}