package parser

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/scanner"
)

// Every node must be able to name a source position. The checker reports
// through objects.DiagnosticHandler, which takes a token.Token by value, so a
// node whose GetPrimaryToken is nil is a diagnostic that cannot be pointed
// anywhere. These tests cover the shapes where that used to happen: literals,
// empty collection literals, empty blocks, and a for loop with no clauses.

func parseSource(t *testing.T, src string) *ast.Module {
	t.Helper()
	tokens, err := scanner.New(bytes.NewBufferString(src), nil).Scan()
	if err != nil {
		t.Fatalf("scan %q: %v", src, err)
	}
	p, collector := createParserFromTokens(tokens)
	mod, err := p.Parse()
	if err != nil {
		t.Fatalf("parse %q: %v (diagnostics: %v)", src, err, collector.Errors)
	}
	if len(collector.Errors) > 0 {
		t.Fatalf("parse %q reported %v", src, collector.Errors)
	}
	return mod
}

// parseSourceExpectingErrors parses source that is meant to be rejected and
// returns the diagnostics it produced.
func parseSourceExpectingErrors(t *testing.T, src string) (*ast.Module, []objects.Diagnostic) {
	t.Helper()
	tokens, err := scanner.New(bytes.NewBufferString(src), nil).Scan()
	if err != nil {
		t.Fatalf("scan %q: %v", src, err)
	}
	p, collector := createParserFromTokens(tokens)
	mod, _ := p.Parse()
	return mod, collector.Errors
}

func TestEveryNodeHasAPosition(t *testing.T) {
	sources := []struct {
		name string
		src  string
	}{
		{"number literal", `var x: number = 1;`},
		{"string literal", `var x: string = "hi";`},
		{"bool literal", `var x: bool = true;`},
		{"empty array literal", `var xs: []number = [];`},
		{"empty hash literal", `var m: map[string]number = {};`},
		{"empty block", `{ }`},
		{"empty function body", `fun f() { }`},
		{"empty lambda body", `var f: fun() = fun() { };`},
		{"for with no clauses", `for (;;) { break; }`},
		{"nested empties", `var xs: []map[string]number = [{}, {}];`},
		{"class with empty method", `class C { m() { } }`},
		{"mixed program", `
			var a: number = 1;
			fun add(x: number, y: number): number { return x + y; }
			class Shape { area(): number { return 0; } }
			class Sized { size: number; init(size: number) { this.size = size; } }
			if (a > 0) { print add(a, 2); } else { print "no"; }
			while (false) { continue; }
			var arr: []number = [1, 2];
			arr[0] = 3;
			print arr[0];
			var m: map[string]bool = {"k": true};
			var f: fun(number): number = fun(n: number): number { return n; };
		`},
	}

	for _, tc := range sources {
		t.Run(tc.name, func(t *testing.T) {
			mod := parseSource(t, tc.src)
			for _, stmt := range mod.GetAllStatements() {
				walkStmt(t, stmt)
			}
		})
	}
}

func requirePosition(t *testing.T, node ast.Node, what string) {
	t.Helper()
	if node == nil {
		return
	}
	if node.GetPrimaryToken() == nil {
		t.Errorf("%s (%T) has no primary token; diagnostics about it cannot be located", what, node)
	}
}

func walkStmt(t *testing.T, stmt ast.Stmt) {
	t.Helper()
	if stmt == nil {
		return
	}
	requirePosition(t, stmt, "statement")

	switch s := stmt.(type) {
	case *ast.ExprStmt:
		walkExpr(t, s.Expr)
	case *ast.PrintStmt:
		walkExpr(t, s.Expr)
	case *ast.VarDeclStmt:
		walkExpr(t, s.Initializer)
	case *ast.BlockStmt:
		for _, inner := range s.Statements {
			walkStmt(t, inner)
		}
	case *ast.IfStmt:
		walkExpr(t, s.Condition)
		walkStmt(t, s.ThenBranch)
		walkStmt(t, s.ElseBranch)
	case *ast.WhileStmt:
		walkExpr(t, s.Condition)
		walkStmt(t, s.Body)
	case *ast.ForStmt:
		walkStmt(t, s.Initializer)
		walkExpr(t, s.Condition)
		walkExpr(t, s.Increment)
		walkStmt(t, s.Body)
	case *ast.FunctionStmt:
		walkStmt(t, s.Body)
	case *ast.ReturnStmt:
		walkExpr(t, s.Value)
	case *ast.ClassStmt:
		for _, method := range s.Methods {
			walkStmt(t, method)
		}
	case *ast.BreakStmt, *ast.ContinueStmt, *ast.ImportStmt:
		// leaves
	default:
		t.Errorf("walkStmt does not handle %T — add it so the invariant stays covered", s)
	}
}

func walkExpr(t *testing.T, expr ast.Expr) {
	t.Helper()
	if expr == nil {
		return
	}
	requirePosition(t, expr, "expression")

	switch e := expr.(type) {
	case *ast.BinaryExpr:
		walkExpr(t, e.Left)
		walkExpr(t, e.Right)
	case *ast.LogicalExpr:
		walkExpr(t, e.Left)
		walkExpr(t, e.Right)
	case *ast.GroupingExpr:
		walkExpr(t, e.Expr)
	case *ast.UnaryExpr:
		walkExpr(t, e.Expr)
	case *ast.AssignExpr:
		walkExpr(t, e.Value)
	case *ast.CallExpr:
		walkExpr(t, e.Callee)
		for _, arg := range e.Arguments {
			walkExpr(t, arg)
		}
	case *ast.GetExpr:
		walkExpr(t, e.Object)
	case *ast.SetExpr:
		walkExpr(t, e.Object)
		walkExpr(t, e.Value)
	case *ast.IndexExpr:
		walkExpr(t, e.Object)
		walkExpr(t, e.Index)
	case *ast.SetIndexExpr:
		walkExpr(t, e.Object)
		walkExpr(t, e.Index)
		walkExpr(t, e.Value)
	case *ast.ArrayLiteralExpr:
		for _, el := range e.Elements {
			walkExpr(t, el)
		}
	case *ast.HashLiteralExpr:
		for _, pair := range e.Pairs {
			walkExpr(t, pair.Key)
			walkExpr(t, pair.Value)
		}
	case *ast.FunctionExpr:
		walkStmt(t, e.Body)
	case *ast.LiteralExpr, *ast.VariableExpr, *ast.ThisExpr, *ast.SuperExpr:
		// leaves
	default:
		t.Errorf("walkExpr does not handle %T — add it so the invariant stays covered", e)
	}
}

// The position must be the real one, not a zero value.
func TestEmptyLiteralPositionsAreAccurate(t *testing.T) {
	mod := parseSource(t, "var xs: []number = [];\nvar m: map[string]number = {};\nfun f() {\n}\n")

	for i := range mod.Statements {
		if mod.Statements[i].GetPrimaryToken() == nil {
			t.Fatalf("statement %d has no token", i)
		}
	}

	xs := mod.Statements[0].(*ast.VarDeclStmt).Initializer
	if got := xs.GetPrimaryToken(); got == nil || got.Line != 1 {
		t.Errorf("empty array literal: got %v, want a token on line 1", fmt.Sprint(got))
	}

	m := mod.Statements[1].(*ast.VarDeclStmt).Initializer
	if got := m.GetPrimaryToken(); got == nil || got.Line != 2 {
		t.Errorf("empty hash literal: got %v, want a token on line 2", fmt.Sprint(got))
	}

	body := mod.Statements[2].(*ast.FunctionStmt).Body
	if got := body.GetPrimaryToken(); got == nil || got.Line != 3 {
		t.Errorf("empty function body: got %v, want the '{' on line 3", fmt.Sprint(got))
	}
}
