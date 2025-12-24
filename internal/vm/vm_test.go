package vm

import (
	"fmt"
	"testing"

	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/compiler"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/token"
)

type vmTestCase struct {
	input    interface{} // Can be ast.Expr or ast.Stmt
	expected interface{}
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{&ast.ExprStmt{
			Expr: &ast.LiteralExpr{Value: 1},
		}, 1},
		{&ast.ExprStmt{
			Expr: &ast.LiteralExpr{Value: 2},
		}, 2},
		{&ast.ExprStmt{
			Expr: &ast.BinaryExpr{
				Left:     &ast.LiteralExpr{Value: 1},
				Right:    &ast.LiteralExpr{Value: 2},
				Operator: &token.Token{Type: token.PLUS},
			},
		}, 3},
		{&ast.ExprStmt{
			Expr: &ast.BinaryExpr{
				Left:     &ast.LiteralExpr{Value: 4},
				Right:    &ast.LiteralExpr{Value: 2},
				Operator: &token.Token{Type: token.SLASH},
			},
		}, 2}, // Floating point 4/2 = 2.0
		{&ast.ExprStmt{
			Expr: &ast.BinaryExpr{
				Left:     &ast.LiteralExpr{Value: 123},
				Right:    &ast.LiteralExpr{Value: 12312},
				Operator: &token.Token{Type: token.MINUS},
			},
		}, 123 - 12312},
		{&ast.ExprStmt{
			Expr: &ast.BinaryExpr{
				Left:     &ast.LiteralExpr{Value: 5},
				Right:    &ast.LiteralExpr{Value: 5},
				Operator: &token.Token{Type: token.STAR},
			},
		}, 25},
	}

	runVmTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTestCase{
		{&ast.ExprStmt{
			Expr: &ast.LiteralExpr{Value: true},
		}, true},
		{&ast.ExprStmt{
			Expr: &ast.LiteralExpr{Value: false},
		}, false},
		{&ast.ExprStmt{
			Expr: &ast.BinaryExpr{
				Left:     &ast.LiteralExpr{Value: 1},
				Right:    &ast.LiteralExpr{Value: 2},
				Operator: &token.Token{Type: token.LESS},
			},
		}, true},
		{&ast.ExprStmt{
			Expr: &ast.BinaryExpr{
				Left:     &ast.LiteralExpr{Value: 1},
				Right:    &ast.LiteralExpr{Value: 2},
				Operator: &token.Token{Type: token.GREATER},
			},
		}, false},
	}

	runVmTests(t, tests)
}

func TestGroupingExpressions(t *testing.T) {
	tests := []vmTestCase{
		// (5)
		{&ast.ExprStmt{
			Expr: &ast.GroupingExpr{Expr: &ast.LiteralExpr{Value: 5}},
		}, 5},
		// (1 + 2)
		{&ast.ExprStmt{
			Expr: &ast.GroupingExpr{
				Expr: &ast.BinaryExpr{
					Left:     &ast.LiteralExpr{Value: 1},
					Right:    &ast.LiteralExpr{Value: 2},
					Operator: &token.Token{Type: token.PLUS},
				},
			},
		}, 3},
		// (1 + 2) * 3
		{&ast.ExprStmt{
			Expr: &ast.BinaryExpr{
				Left: &ast.GroupingExpr{
					Expr: &ast.BinaryExpr{
						Left:     &ast.LiteralExpr{Value: 1},
						Right:    &ast.LiteralExpr{Value: 2},
						Operator: &token.Token{Type: token.PLUS},
					},
				},
				Right:    &ast.LiteralExpr{Value: 3},
				Operator: &token.Token{Type: token.STAR},
			},
		}, 9},
		// 3 * (1 + 2)
		{&ast.ExprStmt{
			Expr: &ast.BinaryExpr{
				Left: &ast.LiteralExpr{Value: 3},
				Right: &ast.GroupingExpr{
					Expr: &ast.BinaryExpr{
						Left:     &ast.LiteralExpr{Value: 1},
						Right:    &ast.LiteralExpr{Value: 2},
						Operator: &token.Token{Type: token.PLUS},
					},
				},
				Operator: &token.Token{Type: token.STAR},
			},
		}, 9},
		// ((5))
		{&ast.ExprStmt{
			Expr: &ast.GroupingExpr{
				Expr: &ast.GroupingExpr{
					Expr: &ast.LiteralExpr{Value: 5},
				},
			},
		}, 5},
		// (2 + 3) * (4 + 5)
		{&ast.ExprStmt{
			Expr: &ast.BinaryExpr{
				Left: &ast.GroupingExpr{
					Expr: &ast.BinaryExpr{
						Left:     &ast.LiteralExpr{Value: 2},
						Right:    &ast.LiteralExpr{Value: 3},
						Operator: &token.Token{Type: token.PLUS},
					},
				},
				Right: &ast.GroupingExpr{
					Expr: &ast.BinaryExpr{
						Left:     &ast.LiteralExpr{Value: 4},
						Right:    &ast.LiteralExpr{Value: 5},
						Operator: &token.Token{Type: token.PLUS},
					},
				},
				Operator: &token.Token{Type: token.STAR},
			},
		}, 45},
		// (true)
		{&ast.ExprStmt{
			Expr: &ast.GroupingExpr{Expr: &ast.LiteralExpr{Value: true}},
		}, true},
		// (!false)
		{&ast.ExprStmt{
			Expr: &ast.GroupingExpr{
				Expr: &ast.UnaryExpr{
					Operator: &token.Token{Type: token.BANG},
					Expr:     &ast.LiteralExpr{Value: false},
				},
			},
		}, true},
		// (1 + 2) * (1 * (2 + 3)) - 12 = 3 * 5 - 12 = 15 - 12 = 3
		{&ast.ExprStmt{
			Expr: &ast.BinaryExpr{
				Left: &ast.BinaryExpr{
					Left: &ast.GroupingExpr{
						Expr: &ast.BinaryExpr{
							Left:     &ast.LiteralExpr{Value: 1},
							Right:    &ast.LiteralExpr{Value: 2},
							Operator: &token.Token{Type: token.PLUS},
						},
					},
					Right: &ast.GroupingExpr{
						Expr: &ast.BinaryExpr{
							Left: &ast.LiteralExpr{Value: 1},
							Right: &ast.GroupingExpr{
								Expr: &ast.BinaryExpr{
									Left:     &ast.LiteralExpr{Value: 2},
									Right:    &ast.LiteralExpr{Value: 3},
									Operator: &token.Token{Type: token.PLUS},
								},
							},
							Operator: &token.Token{Type: token.STAR},
						},
					},
					Operator: &token.Token{Type: token.STAR},
				},
				Right:    &ast.LiteralExpr{Value: 12},
				Operator: &token.Token{Type: token.MINUS},
			},
		}, 3},
	}

	runVmTests(t, tests)
}

func TestUnaryExpressions(t *testing.T) {
	tests := []vmTestCase{
		{&ast.ExprStmt{
			Expr: &ast.UnaryExpr{
				Operator: &token.Token{Type: token.MINUS},
				Expr:     &ast.LiteralExpr{Value: 5},
			},
		}, -5},
		{&ast.ExprStmt{
			Expr: &ast.UnaryExpr{
				Operator: &token.Token{Type: token.BANG},
				Expr:     &ast.LiteralExpr{Value: true},
			},
		}, false},
		{&ast.ExprStmt{
			Expr: &ast.UnaryExpr{
				Operator: &token.Token{Type: token.BANG},
				Expr:     &ast.LiteralExpr{Value: false},
			},
		}, true},
		{&ast.ExprStmt{
			Expr: &ast.UnaryExpr{
				Operator: &token.Token{Type: token.BANG},
				Expr:     &ast.LiteralExpr{Value: 5},
			},
		}, false},
		{&ast.ExprStmt{
			Expr: &ast.UnaryExpr{
				Operator: &token.Token{Type: token.MINUS},
				Expr:     &ast.LiteralExpr{Value: 10},
			},
		}, -10},
	}

	runVmTests(t, tests)
}

func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, tt := range tests {
		comp := compiler.New(nil)
		err := comp.Compile(tt.input)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := New(comp.Bytecode())
		err = vm.Run()
		if err != nil {
			t.Fatalf("vm error: %s", err)
		}

		stackElem := vm.LastPoppedStackElem()
		testExpectedObject(t, tt.expected, stackElem)
	}
}

func testExpectedObject(t *testing.T, expected interface{}, actual objects.Object) {
	t.Helper()

	if expected == nil {
		if actual != nil {
			t.Errorf("expected nil, got=%T (%+v)", actual, actual)
		}
		return
	}

	switch expected := expected.(type) {
	case int:
		err := testIntegerObject(int64(expected), actual)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	case bool:
		err := testBooleanObject(bool(expected), actual)
		if err != nil {
			t.Errorf("testBooleanObject failed: %s", err)
		}
	}
}

func testIntegerObject(expected int64, actual objects.Object) error {
	result, ok := actual.(*objects.Number)
	if !ok {
		return fmt.Errorf("object is not Integer. got=%T (%+v)",
			actual, actual)
	}

	if result.Value != float64(expected) {
		return fmt.Errorf("object has wrong value. want=%d, got=%f",
			expected, result.Value)
	}

	return nil
}

func testBooleanObject(expected bool, actual objects.Object) error {
	result, ok := actual.(*objects.Bool)
	if !ok {
		return fmt.Errorf("object is not Boolean. got=%T (%+v)",
			actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. want=%t, got=%t",
			expected, result.Value)
	}

	return nil
}

func TestConditionals(t *testing.T) {
	// Since if-else is a statement (not expression) in Viri,
	// we test by checking the last popped value from ExprStmt
	tests := []vmTestCase{
		// if (true) { 10; }
		{&ast.IfStmt{
			Condition: &ast.LiteralExpr{Value: true},
			ThenBranch: &ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.ExprStmt{Expr: &ast.LiteralExpr{Value: 10}},
				},
			},
			ElseBranch: nil,
		}, 10},
		// if (true) { 10; } else { 20; }
		{&ast.IfStmt{
			Condition: &ast.LiteralExpr{Value: true},
			ThenBranch: &ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.ExprStmt{Expr: &ast.LiteralExpr{Value: 10}},
				},
			},
			ElseBranch: &ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.ExprStmt{Expr: &ast.LiteralExpr{Value: 20}},
				},
			},
		}, 10},
		// if (false) { 10; } else { 20; }
		{&ast.IfStmt{
			Condition: &ast.LiteralExpr{Value: false},
			ThenBranch: &ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.ExprStmt{Expr: &ast.LiteralExpr{Value: 10}},
				},
			},
			ElseBranch: &ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.ExprStmt{Expr: &ast.LiteralExpr{Value: 20}},
				},
			},
		}, 20},
		// if (1) { 10; }
		{&ast.IfStmt{
			Condition: &ast.LiteralExpr{Value: 1},
			ThenBranch: &ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.ExprStmt{Expr: &ast.LiteralExpr{Value: 10}},
				},
			},
			ElseBranch: nil,
		}, 10},
		// if (1 < 2) { 10; }
		{&ast.IfStmt{
			Condition: &ast.BinaryExpr{
				Left:     &ast.LiteralExpr{Value: 1},
				Right:    &ast.LiteralExpr{Value: 2},
				Operator: &token.Token{Type: token.LESS},
			},
			ThenBranch: &ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.ExprStmt{Expr: &ast.LiteralExpr{Value: 10}},
				},
			},
			ElseBranch: nil,
		}, 10},
		// if (1 > 2) { 10; } else { 20; }
		{&ast.IfStmt{
			Condition: &ast.BinaryExpr{
				Left:     &ast.LiteralExpr{Value: 1},
				Right:    &ast.LiteralExpr{Value: 2},
				Operator: &token.Token{Type: token.GREATER},
			},
			ThenBranch: &ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.ExprStmt{Expr: &ast.LiteralExpr{Value: 10}},
				},
			},
			ElseBranch: &ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.ExprStmt{Expr: &ast.LiteralExpr{Value: 20}},
				},
			},
		}, 20},
	}

	runVmTests(t, tests)
}

func TestBlockStatements(t *testing.T) {
	tests := []vmTestCase{
		// { 10 }
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.ExprStmt{Expr: &ast.LiteralExpr{Value: 10}},
			},
		}, 10},
		// { 10; 20 }
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.ExprStmt{Expr: &ast.LiteralExpr{Value: 10}},
				&ast.ExprStmt{Expr: &ast.LiteralExpr{Value: 20}},
			},
		}, 20},
		// { 1 + 2 }
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.ExprStmt{
					Expr: &ast.BinaryExpr{
						Left:     &ast.LiteralExpr{Value: 1},
						Right:    &ast.LiteralExpr{Value: 2},
						Operator: &token.Token{Type: token.PLUS},
					},
				},
			},
		}, 3},
	}

	runVmTests(t, tests)
}

func TestGlobalVarStatements(t *testing.T) {
	tests := []vmTestCase{
		// var one = 1; one;
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.VarDeclStmt{
					Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "one"},
					Initializer: &ast.LiteralExpr{Value: 1},
					IsConst:     false,
				},
				&ast.ExprStmt{
					Expr: &ast.VariableExpr{
						Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "one"},
					},
				},
			},
		}, 1},
		// var one = 1; var two = 2; one + two;
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.VarDeclStmt{
					Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "one"},
					Initializer: &ast.LiteralExpr{Value: 1},
					IsConst:     false,
				},
				&ast.VarDeclStmt{
					Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "two"},
					Initializer: &ast.LiteralExpr{Value: 2},
					IsConst:     false,
				},
				&ast.ExprStmt{
					Expr: &ast.BinaryExpr{
						Left: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "one"},
						},
						Right: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "two"},
						},
						Operator: &token.Token{Type: token.PLUS},
					},
				},
			},
		}, 3},
		// var one = 1; var two = one + one; one + two;
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.VarDeclStmt{
					Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "one"},
					Initializer: &ast.LiteralExpr{Value: 1},
					IsConst:     false,
				},
				&ast.VarDeclStmt{
					Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "two"},
					Initializer: &ast.BinaryExpr{
						Left: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "one"},
						},
						Right: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "one"},
						},
						Operator: &token.Token{Type: token.PLUS},
					},
					IsConst: false,
				},
				&ast.ExprStmt{
					Expr: &ast.BinaryExpr{
						Left: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "one"},
						},
						Right: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "two"},
						},
						Operator: &token.Token{Type: token.PLUS},
					},
				},
			},
		}, 3},
	}

	runVmTests(t, tests)
}

func TestGlobalAssignment(t *testing.T) {
	tests := []vmTestCase{
		// var x = 1; x = 5; x;
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.VarDeclStmt{
					Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "x"},
					Initializer: &ast.LiteralExpr{Value: 1},
					IsConst:     false,
				},
				&ast.ExprStmt{
					Expr: &ast.AssignExpr{
						Name:  &token.Token{Type: token.IDENTIFIER, Lexeme: "x"},
						Value: &ast.LiteralExpr{Value: 5},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.VariableExpr{
						Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "x"},
					},
				},
			},
		}, 5},
		// var x = 1; var y = 2; x = y; x;
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.VarDeclStmt{
					Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "x"},
					Initializer: &ast.LiteralExpr{Value: 1},
					IsConst:     false,
				},
				&ast.VarDeclStmt{
					Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "y"},
					Initializer: &ast.LiteralExpr{Value: 2},
					IsConst:     false,
				},
				&ast.ExprStmt{
					Expr: &ast.AssignExpr{
						Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "x"},
						Value: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "y"},
						},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.VariableExpr{
						Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "x"},
					},
				},
			},
		}, 2},
	}

	runVmTests(t, tests)
}
