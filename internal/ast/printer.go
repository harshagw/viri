package ast

import (
	"fmt"
	"strings"
)

// Printer renders AST nodes in a tree form for debugging.
type Printer struct {
	builder *strings.Builder
	prefix  string
	isLast  bool
}

func NewPrinter() *Printer {
	return &Printer{}
}

func (p *Printer) PrintStatements(statements []Stmt) string {
	var b strings.Builder
	p.builder = &b
	for i, stmt := range statements {
		p.prefix = ""
		p.isLast = i == len(statements)-1
		p.printStmt(stmt)
	}
	return b.String()
}

func (p *Printer) PrintExpr(expr Expr) string {
	var b strings.Builder
	p.builder = &b
	p.prefix = ""
	p.isLast = true
	p.printExpr(expr)
	return b.String()
}

// Internal helpers.

func (p *Printer) printExpr(expr Expr) {
	switch n := expr.(type) {
	case *BinaryExpr:
		p.writeNode(fmt.Sprintf("Binary (%s)", n.Operator.Lexeme))
		newPrefix := p.childPrefix()
		p.withPrefix(newPrefix, false, func() { p.printExpr(n.Left) })
		p.withPrefix(newPrefix, true, func() { p.printExpr(n.Right) })
	case *UnaryExpr:
		p.writeNode(fmt.Sprintf("Unary (%s)", n.Operator.Lexeme))
		p.withPrefix(p.childPrefix(), true, func() { p.printExpr(n.Expr) })
	case *GroupingExpr:
		p.writeNode("Grouping")
		p.withPrefix(p.childPrefix(), true, func() { p.printExpr(n.Expr) })
	case *LiteralExpr:
		p.writeNode(fmt.Sprintf("Literal (%v)", n.Value))
	case *VariableExpr:
		p.writeNode("Variable (" + n.Name.Lexeme + ")")
	case *AssignExpr:
		p.writeNode("Assign")
		newPrefix := p.childPrefix()
		p.withPrefix(newPrefix, false, func() { p.writeNode(n.Name.Lexeme) })
		p.withPrefix(newPrefix, true, func() { p.printExpr(n.Value) })
	case *LogicalExpr:
		p.writeNode(fmt.Sprintf("Logical (%s)", n.Operator.Lexeme))
		newPrefix := p.childPrefix()
		p.withPrefix(newPrefix, false, func() { p.printExpr(n.Left) })
		p.withPrefix(newPrefix, true, func() { p.printExpr(n.Right) })
	case *CallExpr:
		p.writeNode("Call")
		newPrefix := p.childPrefix()
		p.withPrefix(newPrefix, false, func() { p.printExpr(n.Callee) })
		p.withPrefix(newPrefix, true, func() {
			p.writeNode("arguments")
			argsPrefix := p.childPrefix()
			if len(n.Arguments) == 0 {
				p.withPrefix(argsPrefix, true, func() { p.writeNode("nil") })
				return
			}
			for i, arg := range n.Arguments {
				p.withPrefix(argsPrefix, i == len(n.Arguments)-1, func() { p.printExpr(arg) })
			}
		})
	case *GetExpr:
		p.writeNode("Get")
		newPrefix := p.childPrefix()
		p.withPrefix(newPrefix, false, func() { p.printExpr(n.Object) })
		p.withPrefix(newPrefix, true, func() { p.writeNode("field (" + n.Name.Lexeme + ")") })
	case *SetExpr:
		p.writeNode("Set")
		newPrefix := p.childPrefix()
		p.withPrefix(newPrefix, false, func() { p.printExpr(n.Object) })
		p.withPrefix(newPrefix, true, func() {
			p.writeNode("field (" + n.Name.Lexeme + ")")
			p.withPrefix(p.childPrefix(), true, func() { p.printExpr(n.Value) })
		})
	case *ThisExpr:
		p.writeNode("This")
	case *ArrayLiteralExpr:
		p.writeNode("ArrayLiteral")
		newPrefix := p.childPrefix()
		p.withPrefix(newPrefix, true, func() {
			p.writeNode("elements")
			elementsPrefix := p.childPrefix()
			for i, element := range n.Elements {
				p.withPrefix(elementsPrefix, i == len(n.Elements)-1, func() { p.printExpr(element) })
			}
		})
	case *IndexExpr:
		p.writeNode("Index")
		newPrefix := p.childPrefix()
		p.withPrefix(newPrefix, false, func() { p.printExpr(n.Object) })
		p.withPrefix(newPrefix, true, func() { 
			p.writeNode("index")
			p.withPrefix(p.childPrefix(), true, func() { p.printExpr(n.Index) })
		})
	case *SetIndexExpr:
		p.writeNode("SetIndex")
		newPrefix := p.childPrefix()
		p.withPrefix(newPrefix, false, func() { 
			p.writeNode("object")
			p.withPrefix(p.childPrefix(), true, func() { p.printExpr(n.Object) })
		})
		p.withPrefix(newPrefix, false, func() {
			p.writeNode("index")
			p.withPrefix(p.childPrefix(), true, func() { p.printExpr(n.Index) })
		})
		p.withPrefix(newPrefix, true, func() { 
			p.writeNode("value")
			p.withPrefix(p.childPrefix(), true, func() { p.printExpr(n.Value) })
		 })
	case *HashLiteralExpr:
		p.writeNode("HashLiteral")
		newPrefix := p.childPrefix()
		p.withPrefix(newPrefix, true, func() {
			p.writeNode("pairs")
			pairsPrefix := p.childPrefix()
			for i, pair := range n.Pairs {
				p.withPrefix(pairsPrefix, i == len(n.Pairs)-1, func() { p.printExpr(pair.Key) })
				p.withPrefix(pairsPrefix, true, func() { p.printExpr(pair.Value) })
			}
		})
	default:
		p.writeNode("Unknown Expr")
	}
}

func (p *Printer) printStmt(stmt Stmt) {
	switch n := stmt.(type) {
	case *ExprStmt:
		p.writeNode("ExprStmt")
		p.withPrefix(p.childPrefix(), true, func() { p.printExpr(n.Expr) })
	case *PrintStmt:
		p.writeNode("PrintStmt")
		p.withPrefix(p.childPrefix(), true, func() { p.printExpr(n.Expr) })
	case *VarDeclStmt:
		p.writeNode("VarDecl (" + n.Name.Lexeme + ")")
		if n.Initializer != nil {
			p.withPrefix(p.childPrefix(), true, func() { p.printExpr(n.Initializer) })
		}
	case *BlockStmt:
		p.writeNode("Block")
		newPrefix := p.childPrefix()
		for i, s := range n.Statements {
			p.withPrefix(newPrefix, i == len(n.Statements)-1, func() { p.printStmt(s) })
		}
	case *IfStmt:
		p.writeNode("If")
		newPrefix := p.childPrefix()
		p.withPrefix(newPrefix, false, func() {
			p.writeNode("condition")
			p.withPrefix(p.childPrefix(), true, func() { p.printExpr(n.Condition) })
		})
		p.withPrefix(newPrefix, n.ElseBranch == nil, func() {
			p.writeNode("then")
			p.withPrefix(p.childPrefix(), true, func() { p.printStmt(n.ThenBranch) })
		})
		if n.ElseBranch != nil {
			p.withPrefix(newPrefix, true, func() {
				p.writeNode("else")
				p.withPrefix(p.childPrefix(), true, func() { p.printStmt(n.ElseBranch) })
			})
		}
	case *WhileStmt:
		p.writeNode("While")
		newPrefix := p.childPrefix()
		p.withPrefix(newPrefix, false, func() {
			p.writeNode("condition")
			p.withPrefix(p.childPrefix(), true, func() { p.printExpr(n.Condition) })
		})
		p.withPrefix(newPrefix, true, func() {
			p.writeNode("body")
			p.withPrefix(p.childPrefix(), true, func() { p.printStmt(n.Body) })
		})
	case *BreakStmt:
		p.writeNode("Break")
	case *FunctionStmt:
		p.writeNode("Function (" + n.Name.Lexeme + ")")
		newPrefix := p.childPrefix()
		p.withPrefix(newPrefix, false, func() {
			p.writeNode("params")
			paramsPrefix := p.childPrefix()
			for i, param := range n.Params {
				p.withPrefix(paramsPrefix, i == len(n.Params)-1, func() { p.writeNode(param.Lexeme) })
			}
		})
		p.withPrefix(newPrefix, true, func() {
			p.writeNode("body")
			p.withPrefix(p.childPrefix(), true, func() { p.printStmt(n.Body) })
		})
	case *ReturnStmt:
		p.writeNode("Return")
		if n.Value != nil {
			p.withPrefix(p.childPrefix(), true, func() { p.printExpr(n.Value) })
		}
	case *ClassStmt:
		p.writeNode("Class (" + n.Name.Lexeme + ")")
		newPrefix := p.childPrefix()
		p.withPrefix(newPrefix, true, func() {
			p.writeNode("methods")
			methodsPrefix := p.childPrefix()
			for i, m := range n.Methods {
				p.withPrefix(methodsPrefix, i == len(n.Methods)-1, func() {
					p.writeNode(m.Name.Lexeme)
				})
			}
		})
	default:
		p.writeNode("Unknown Stmt")
	}
}

func (p *Printer) writeNode(label string) {
	p.builder.WriteString(p.prefix)
	if p.isLast {
		p.builder.WriteString("└── ")
	} else {
		p.builder.WriteString("├── ")
	}
	p.builder.WriteString(label)
	p.builder.WriteString("\n")
}

func (p *Printer) withPrefix(prefix string, isLast bool, fn func()) {
	oldPrefix, oldLast := p.prefix, p.isLast
	p.prefix, p.isLast = prefix, isLast
	fn()
	p.prefix, p.isLast = oldPrefix, oldLast
}

func (p *Printer) childPrefix() string {
	if p.isLast {
		return p.prefix + "    "
	}
	return p.prefix + "│   "
}
