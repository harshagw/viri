package parser

import (
	"strings"
	"testing"

	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/token"
)

func createResolverFromAST() (*Resolver, *objects.DiagnosticCollector) {
	collector := &objects.DiagnosticCollector{}
	resolver := NewResolver(collector)
	return resolver, collector
}

func TestResolveLocalVariable(t *testing.T) {
	xTok := token.Token{Type: token.IDENTIFIER, Lexeme: "x", Literal: nil, Line: 1, FilePath: nil}
	yTok := token.Token{Type: token.IDENTIFIER, Lexeme: "y", Literal: nil, Line: 1, FilePath: nil}
	mod := &ast.Module{
		Statements: []ast.Stmt{
			&ast.VarDeclStmt{
				Name:        &xTok,
				Initializer: &ast.LiteralExpr{Value: 1.0},
			},
			&ast.VarDeclStmt{
				Name:        &yTok,
				Initializer: &ast.VariableExpr{Name: &xTok},
			},
		},
	}
	resolver, _ := createResolverFromAST()
	locals, err := resolver.Resolve(mod)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	// Verify x in y's initializer is resolved
	varExpr := mod.Statements[1].(*ast.VarDeclStmt).Initializer.(*ast.VariableExpr)
	assertResolved(t, locals, varExpr, "x", 0)
}

func TestResolveVariableInBlock(t *testing.T) {
	xTok := token.Token{Type: token.IDENTIFIER, Lexeme: "x", Literal: nil, Line: 1, FilePath: nil}
	mod := &ast.Module{
		Statements: []ast.Stmt{
			&ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.VarDeclStmt{
						Name:        &xTok,
						Initializer: &ast.LiteralExpr{Value: 1.0},
					},
					&ast.ExprStmt{
						Expr: &ast.VariableExpr{Name: &xTok},
					},
				},
			},
		},
	}
	resolver, _ := createResolverFromAST()
	locals, err := resolver.Resolve(mod)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	block := mod.Statements[0].(*ast.BlockStmt)
	varExpr := block.Statements[1].(*ast.ExprStmt).Expr.(*ast.VariableExpr)
	assertResolved(t, locals, varExpr, "x", 0)
}

func TestResolveVariableBeforeDeclaration(t *testing.T) {
	// Represents: var x = y; var y = 1;
	xTok := token.Token{Type: token.IDENTIFIER, Lexeme: "x", Literal: nil, Line: 1, FilePath: nil}
	yTok := token.Token{Type: token.IDENTIFIER, Lexeme: "y", Literal: nil, Line: 1, FilePath: nil}
	mod := &ast.Module{
		Statements: []ast.Stmt{
			&ast.VarDeclStmt{
				Name:        &xTok,
				Initializer: &ast.VariableExpr{Name: &yTok},
			},
			&ast.VarDeclStmt{
				Name:        &yTok,
				Initializer: &ast.LiteralExpr{Value: 1.0},
			},
		},
	}
	resolver, _ := createResolverFromAST()
	locals, err := resolver.Resolve(mod)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	// y should NOT be in locals because it is global
	varExpr := mod.Statements[0].(*ast.VarDeclStmt).Initializer.(*ast.VariableExpr)
	assertNotResolved(t, locals, varExpr, "y")
}

func TestResolveFunctionParameters(t *testing.T) {
	// Represents: fun foo(x) { return x; }
	fooTok := token.Token{Type: token.IDENTIFIER, Lexeme: "foo", Literal: nil, Line: 1, FilePath: nil}
	xTok := token.Token{Type: token.IDENTIFIER, Lexeme: "x", Literal: nil, Line: 1, FilePath: nil}
	mod := &ast.Module{
		Statements: []ast.Stmt{
			&ast.FunctionStmt{
				Name:   &fooTok,
				Params: []*token.Token{&xTok},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.ReturnStmt{
							Keyword: &token.Token{Type: token.RETURN, Lexeme: "return"},
							Value:   &ast.VariableExpr{Name: &xTok},
						},
					},
				},
			},
			&ast.ExprStmt{
				Expr: &ast.VariableExpr{Name: &fooTok},
			},
		},
	}
	resolver, _ := createResolverFromAST()
	locals, err := resolver.Resolve(mod)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	// x should be resolved
	retStmt := mod.Statements[0].(*ast.FunctionStmt).Body.Statements[0].(*ast.ReturnStmt)
	varExpr := retStmt.Value.(*ast.VariableExpr)
	assertResolved(t, locals, varExpr, "x", 0)
}

func TestResolveClassMethod(t *testing.T) {
	// Represents: class Foo { bar() { return this; } }
	fooTok := token.Token{Type: token.IDENTIFIER, Lexeme: "Foo", Literal: nil, Line: 1, FilePath: nil}
	barTok := token.Token{Type: token.IDENTIFIER, Lexeme: "bar", Literal: nil, Line: 1, FilePath: nil}
	thisTok := token.Token{Type: token.THIS, Lexeme: "this", Literal: nil, Line: 1, FilePath: nil}
	mod := &ast.Module{
		Statements: []ast.Stmt{
			&ast.ClassStmt{
				Name: &fooTok,
				Methods: []*ast.FunctionStmt{
					{
						Name: &barTok,
						Body: &ast.BlockStmt{
							Statements: []ast.Stmt{
								&ast.ReturnStmt{
									Keyword: &token.Token{Type: token.RETURN, Lexeme: "return"},
									Value:   &ast.ThisExpr{Keyword: &thisTok},
								},
							},
						},
					},
				},
			},
			&ast.ExprStmt{
				Expr: &ast.VariableExpr{Name: &fooTok},
			},
		},
	}
	resolver, _ := createResolverFromAST()
	locals, err := resolver.Resolve(mod)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	// this should be resolved
	classStmt := mod.Statements[0].(*ast.ClassStmt)
	retStmt := classStmt.Methods[0].Body.Statements[0].(*ast.ReturnStmt)
	thisExpr := retStmt.Value.(*ast.ThisExpr)
	assertResolved(t, locals, thisExpr, "this", 1)
}

func TestResolveSuperExpression(t *testing.T) {
	// Represents: class Foo < Bar { baz() { return super.method(); } }
	fooTok := token.Token{Type: token.IDENTIFIER, Lexeme: "Foo", Literal: nil, Line: 1, FilePath: nil}
	barTok := token.Token{Type: token.IDENTIFIER, Lexeme: "Bar", Literal: nil, Line: 1, FilePath: nil}
	bazTok := token.Token{Type: token.IDENTIFIER, Lexeme: "baz", Literal: nil, Line: 1, FilePath: nil}
	superTok := token.Token{Type: token.SUPER, Lexeme: "super", Literal: nil, Line: 1, FilePath: nil}
	methodTok := token.Token{Type: token.IDENTIFIER, Lexeme: "method", Literal: nil, Line: 1, FilePath: nil}
	mod := &ast.Module{
		Statements: []ast.Stmt{
			&ast.ClassStmt{
				Name: &fooTok,
				SuperClass: &ast.VariableExpr{
					Name: &barTok,
				},
				Methods: []*ast.FunctionStmt{
					{
						Name: &bazTok,
						Body: &ast.BlockStmt{
							Statements: []ast.Stmt{
								&ast.ReturnStmt{
									Keyword: &token.Token{Type: token.RETURN, Lexeme: "return"},
									Value: &ast.CallExpr{
										Callee: &ast.SuperExpr{
											Keyword: &superTok,
											Method:  &methodTok,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	resolver, _ := createResolverFromAST()
	locals, err := resolver.Resolve(mod)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	// super should be resolved
	classStmt := mod.Statements[0].(*ast.ClassStmt)
	retStmt := classStmt.Methods[0].Body.Statements[0].(*ast.ReturnStmt)
	superExpr := retStmt.Value.(*ast.CallExpr).Callee.(*ast.SuperExpr)
	assertResolved(t, locals, superExpr, "super", 2)
}

func TestResolveThisOutsideClass(t *testing.T) {
	// Represents: var x = this;
	thisTok := token.Token{Type: token.THIS, Lexeme: "this", Literal: nil, Line: 1, FilePath: nil}
	mod := &ast.Module{
		Statements: []ast.Stmt{
			&ast.VarDeclStmt{
				Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "x"},
				Initializer: &ast.ThisExpr{Keyword: &thisTok},
			},
		},
	}
	resolver, collector := createResolverFromAST()
	_, _ = resolver.Resolve(mod)
	if len(collector.Errors) == 0 {
		t.Error("expected error for using 'this' outside class")
	}
}

func TestResolveSuperOutsideClass(t *testing.T) {
	// Represents: var x = super.method;
	superTok := token.Token{Type: token.SUPER, Lexeme: "super", Literal: nil, Line: 1, FilePath: nil}
	mod := &ast.Module{
		Statements: []ast.Stmt{
			&ast.VarDeclStmt{
				Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "x"},
				Initializer: &ast.SuperExpr{
					Keyword: &superTok,
					Method:  &token.Token{Type: token.IDENTIFIER, Lexeme: "method"},
				},
			},
		},
	}
	resolver, collector := createResolverFromAST()
	_, _ = resolver.Resolve(mod)
	if len(collector.Errors) == 0 {
		t.Error("expected error for using 'super' outside class")
	}
}

func TestResolveSuperInNonSubclass(t *testing.T) {
	// Represents: class Foo { baz() { return super.method(); } }
	fooTok := token.Token{Type: token.IDENTIFIER, Lexeme: "Foo"}
	bazTok := token.Token{Type: token.IDENTIFIER, Lexeme: "baz"}
	superTok := token.Token{Type: token.SUPER, Lexeme: "super"}
	methodTok := token.Token{Type: token.IDENTIFIER, Lexeme: "method"}
	mod := &ast.Module{
		Statements: []ast.Stmt{
			&ast.ClassStmt{
				Name: &fooTok,
				Methods: []*ast.FunctionStmt{
					{
						Name: &bazTok,
						Body: &ast.BlockStmt{
							Statements: []ast.Stmt{
								&ast.ReturnStmt{
									Keyword: &token.Token{Type: token.RETURN, Lexeme: "return"},
									Value: &ast.CallExpr{
										Callee: &ast.SuperExpr{
											Keyword: &superTok,
											Method:  &methodTok,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	resolver, collector := createResolverFromAST()
	_, _ = resolver.Resolve(mod)
	found := false
	for _, err := range collector.Errors {
		if strings.Contains(err.Message, "no superclass") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected error for using 'super' in a class with no superclass")
	}
}

func TestResolveReturnOutsideFunction(t *testing.T) {
	// Represents: return 1;
	mod := &ast.Module{
		Statements: []ast.Stmt{
			&ast.ReturnStmt{
				Keyword: &token.Token{Type: token.RETURN, Lexeme: "return"},
				Value:   &ast.LiteralExpr{Value: 1.0},
			},
		},
	}
	resolver, collector := createResolverFromAST()
	_, _ = resolver.Resolve(mod)
	if len(collector.Errors) == 0 {
		t.Error("expected error for return outside function")
	}
}

func TestResolveBreakOutsideLoop(t *testing.T) {
	// Represents: break;
	mod := &ast.Module{
		Statements: []ast.Stmt{
			&ast.BreakStmt{
				Keyword: &token.Token{Type: token.BREAK, Lexeme: "break"},
			},
		},
	}
	resolver, collector := createResolverFromAST()
	_, _ = resolver.Resolve(mod)
	if len(collector.Errors) == 0 {
		t.Error("expected error for break outside loop")
	}
}

func TestResolveBreakInsideLoop(t *testing.T) {
	// Represents: while (true) { break; }
	mod := &ast.Module{
		Statements: []ast.Stmt{
			&ast.WhileStmt{
				Condition: &ast.LiteralExpr{Value: true},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.BreakStmt{
							Keyword: &token.Token{Type: token.BREAK, Lexeme: "break"},
						},
					},
				},
			},
		},
	}
	resolver, collector := createResolverFromAST()
	_, err := resolver.Resolve(mod)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(collector.Errors) > 0 {
		t.Errorf("got %d unexpected errors", len(collector.Errors))
	}
}

func TestResolveDuplicateVariableDeclaration(t *testing.T) {
	// Represents: { var x = 1; var x = 2; }
	xTok := token.Token{Type: token.IDENTIFIER, Lexeme: "x", Literal: nil, Line: 1, FilePath: nil}
	mod := &ast.Module{
		Statements: []ast.Stmt{
			&ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.VarDeclStmt{
						Name:        &xTok,
						Initializer: &ast.LiteralExpr{Value: 1.0},
					},
					&ast.VarDeclStmt{
						Name:        &xTok,
						Initializer: &ast.LiteralExpr{Value: 2.0},
					},
				},
			},
		},
	}
	resolver, collector := createResolverFromAST()
	_, _ = resolver.Resolve(mod)
	if len(collector.Errors) == 0 {
		t.Error("expected error for duplicate variable declaration in same scope")
	}
}

func TestResolveUnusedVariableWarning(t *testing.T) {
	// Represents: { var x = 1; }
	xTok := token.Token{Type: token.IDENTIFIER, Lexeme: "x", Literal: nil, Line: 1, FilePath: nil}
	mod := &ast.Module{
		Statements: []ast.Stmt{
			&ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.VarDeclStmt{
						Name:        &xTok,
						Initializer: &ast.LiteralExpr{Value: 1.0},
					},
				},
			},
		},
	}
	resolver, collector := createResolverFromAST()
	_, err := resolver.Resolve(mod)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(collector.Warnings) == 0 {
		t.Error("expected warning for unused variable")
	}
}

func TestResolveUnusedGlobalVariable(t *testing.T) {
	// Represents: var x = 1;
	xTok := token.Token{Type: token.IDENTIFIER, Lexeme: "x"}
	mod := &ast.Module{
		Statements: []ast.Stmt{
			&ast.VarDeclStmt{
				Name:        &xTok,
				Initializer: &ast.LiteralExpr{Value: 1.0},
			},
		},
	}
	resolver, collector := createResolverFromAST()
	_, err := resolver.Resolve(mod)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(collector.Warnings) == 0 {
		t.Error("expected warning for unused global variable")
	}
}

func TestResolveClassInheritanceCycle(t *testing.T) {
	// Represents: class Foo < Foo {}
	fooTok := token.Token{Type: token.IDENTIFIER, Lexeme: "Foo", Literal: nil, Line: 1, FilePath: nil}
	mod := &ast.Module{
		Statements: []ast.Stmt{
			&ast.ClassStmt{
				Name: &fooTok,
				SuperClass: &ast.VariableExpr{
					Name: &fooTok,
				},
			},
		},
	}
	resolver, collector := createResolverFromAST()
	_, _ = resolver.Resolve(mod)
	if len(collector.Errors) == 0 {
		t.Error("expected error for class inheriting from itself")
	}
}

func TestResolveVariableDepth(t *testing.T) {
	// Represents:
	// var x = 1;
	// {
	//   var y = x;
	//   {
	//     var z = y;
	//     print x;
	//   }
	// }
	xTok := token.Token{Type: token.IDENTIFIER, Lexeme: "x"}
	yTok := token.Token{Type: token.IDENTIFIER, Lexeme: "y"}
	zTok := token.Token{Type: token.IDENTIFIER, Lexeme: "z"}
	mod := &ast.Module{
		Statements: []ast.Stmt{
			&ast.VarDeclStmt{
				Name:        &xTok,
				Initializer: &ast.LiteralExpr{Value: 1.0},
			},
			&ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.VarDeclStmt{
						Name:        &yTok,
						Initializer: &ast.VariableExpr{Name: &xTok}, // x should be at depth 1
					},
					&ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.VarDeclStmt{
								Name:        &zTok,
								Initializer: &ast.VariableExpr{Name: &yTok}, // y should be at depth 1
							},
							&ast.ExprStmt{
								Expr: &ast.VariableExpr{Name: &zTok},
							},
							&ast.ExprStmt{
								Expr: &ast.VariableExpr{Name: &xTok}, // x should be at depth 2
							},
						},
					},
				},
			},
		},
	}
	resolver, _ := createResolverFromAST()
	locals, err := resolver.Resolve(mod)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	block1 := mod.Statements[1].(*ast.BlockStmt)
	xExpr1 := block1.Statements[0].(*ast.VarDeclStmt).Initializer.(*ast.VariableExpr)
	assertResolved(t, locals, xExpr1, "x", 1)

	block2 := block1.Statements[1].(*ast.BlockStmt)
	yExpr := block2.Statements[0].(*ast.VarDeclStmt).Initializer.(*ast.VariableExpr)
	assertResolved(t, locals, yExpr, "y", 1)

	xExpr2 := block2.Statements[2].(*ast.ExprStmt).Expr.(*ast.VariableExpr)
	assertResolved(t, locals, xExpr2, "x", 2)
}

func TestResolveStaticScoping(t *testing.T) {
	// Represents:
	// var a = "global";
	// {
	//   fun showA() { print a; }
	//   showA();
	//   var a = "block";
	//   showA();
	//   print a;
	// }
	aTok := token.Token{Type: token.IDENTIFIER, Lexeme: "a"}
	showATok := token.Token{Type: token.IDENTIFIER, Lexeme: "showA"}
	mod := &ast.Module{
		Statements: []ast.Stmt{
			&ast.VarDeclStmt{
				Name:        &aTok,
				Initializer: &ast.LiteralExpr{Value: "global"},
			},
			&ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.FunctionStmt{
						Name: &showATok,
						Body: &ast.BlockStmt{
							Statements: []ast.Stmt{
								&ast.PrintStmt{
									Expr: &ast.VariableExpr{Name: &aTok}, // Usage 1: inside function
								},
							},
						},
					},
					&ast.ExprStmt{
						Expr: &ast.CallExpr{Callee: &ast.VariableExpr{Name: &showATok}},
					},
					&ast.VarDeclStmt{
						Name:        &aTok,
						Initializer: &ast.LiteralExpr{Value: "block"},
					},
					&ast.ExprStmt{
						Expr: &ast.CallExpr{Callee: &ast.VariableExpr{Name: &showATok}},
					},
					&ast.PrintStmt{
						Expr: &ast.VariableExpr{Name: &aTok}, // Usage 2: direct block usage
					},
				},
			},
		},
	}
	resolver, _ := createResolverFromAST()
	locals, err := resolver.Resolve(mod)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	block := mod.Statements[1].(*ast.BlockStmt)
	showAFn := block.Statements[0].(*ast.FunctionStmt)
	printInsideFn := showAFn.Body.Statements[0].(*ast.PrintStmt).Expr.(*ast.VariableExpr)
	// Inside showA, 'a' was resolved before the local 'a' was declared, so it should be global.
	// Scopes: [Global, Block, showA-body]. Depth = 3 - 0 - 1 = 2.
	assertResolved(t, locals, printInsideFn, "a (global)", 2)

	printAtEnd := block.Statements[4].(*ast.PrintStmt).Expr.(*ast.VariableExpr)
	// At the end of the block, 'a' refers to the local 'a' declared in that block.
	// Scopes: [Global, Block]. Depth = 2 - 1 - 1 = 0.
	assertResolved(t, locals, printAtEnd, "a (local)", 0)
}

func assertResolved(t *testing.T, locals map[ast.Expr]int, expr ast.Expr, name string, expectedDepth int) {
	t.Helper()
	depth, found := locals[expr]
	if !found {
		t.Errorf("expected '%s' to be resolved", name)
		return
	}
	if depth != expectedDepth {
		t.Errorf("expected '%s' to have depth %d, got %d", name, expectedDepth, depth)
	}
}

func assertNotResolved(t *testing.T, locals map[ast.Expr]int, expr ast.Expr, name string) {
	t.Helper()
	if _, found := locals[expr]; found {
		t.Errorf("expected '%s' NOT to be resolved", name)
	}
}
