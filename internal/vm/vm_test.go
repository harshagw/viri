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
		if actual != nil && actual.Type() != objects.TypeNil {
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
	case string:
		err := testStringObject(expected, actual)
		if err != nil {
			t.Errorf("testStringObject failed: %s", err)
		}
	case []int:
		array, ok := actual.(*objects.Array)
		if !ok {
			t.Errorf("object is not Array. got=%T (%+v)", actual, actual)
			return
		}
		if len(array.Elements) != len(expected) {
			t.Errorf("wrong number of elements. want=%d, got=%d",
				len(expected), len(array.Elements))
			return
		}
		for i, expectedElem := range expected {
			err := testIntegerObject(int64(expectedElem), array.Elements[i])
			if err != nil {
				t.Errorf("testIntegerObject failed for element %d: %s", i, err)
			}
		}
	case map[string]int:
		hash, ok := actual.(*objects.Hash)
		if !ok {
			t.Errorf("object is not Hash. got=%T (%+v)", actual, actual)
			return
		}
		if len(hash.Pairs) != len(expected) {
			t.Errorf("wrong number of pairs. want=%d, got=%d",
				len(expected), len(hash.Pairs))
			return
		}
		for key, expectedVal := range expected {
			val, ok := hash.Pairs[key]
			if !ok {
				t.Errorf("no value for key %q in hash", key)
				continue
			}
			err := testIntegerObject(int64(expectedVal), val)
			if err != nil {
				t.Errorf("testIntegerObject failed for key %q: %s", key, err)
			}
		}
	}
}

func testStringObject(expected string, actual objects.Object) error {
	result, ok := actual.(*objects.String)
	if !ok {
		return fmt.Errorf("object is not String. got=%T (%+v)",
			actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. want=%q, got=%q",
			expected, result.Value)
	}

	return nil
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

func TestStringExpressions(t *testing.T) {
	tests := []vmTestCase{
		{&ast.ExprStmt{
			Expr: &ast.LiteralExpr{Value: "hello"},
		}, "hello"},
		// String concatenation
		{&ast.ExprStmt{
			Expr: &ast.BinaryExpr{
				Left:     &ast.LiteralExpr{Value: "hello"},
				Right:    &ast.LiteralExpr{Value: " world"},
				Operator: &token.Token{Type: token.PLUS},
			},
		}, "hello world"},
		// Number + String
		{&ast.ExprStmt{
			Expr: &ast.BinaryExpr{
				Left:     &ast.LiteralExpr{Value: 42},
				Right:    &ast.LiteralExpr{Value: " is the answer"},
				Operator: &token.Token{Type: token.PLUS},
			},
		}, "42 is the answer"},
		// String + Number
		{&ast.ExprStmt{
			Expr: &ast.BinaryExpr{
				Left:     &ast.LiteralExpr{Value: "The answer is "},
				Right:    &ast.LiteralExpr{Value: 42},
				Operator: &token.Token{Type: token.PLUS},
			},
		}, "The answer is 42"},
		// Float number + String
		{&ast.ExprStmt{
			Expr: &ast.BinaryExpr{
				Left:     &ast.LiteralExpr{Value: 3.14},
				Right:    &ast.LiteralExpr{Value: " is pi"},
				Operator: &token.Token{Type: token.PLUS},
			},
		}, "3.14 is pi"},
	}

	runVmTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []vmTestCase{
		// []
		{&ast.ExprStmt{
			Expr: &ast.ArrayLiteralExpr{Elements: []ast.Expr{}},
		}, []int{}},
		// [1, 2, 3]
		{&ast.ExprStmt{
			Expr: &ast.ArrayLiteralExpr{
				Elements: []ast.Expr{
					&ast.LiteralExpr{Value: 1},
					&ast.LiteralExpr{Value: 2},
					&ast.LiteralExpr{Value: 3},
				},
			},
		}, []int{1, 2, 3}},
		// [1 + 2, 3 * 4, 5 + 6]
		{&ast.ExprStmt{
			Expr: &ast.ArrayLiteralExpr{
				Elements: []ast.Expr{
					&ast.BinaryExpr{
						Left:     &ast.LiteralExpr{Value: 1},
						Right:    &ast.LiteralExpr{Value: 2},
						Operator: &token.Token{Type: token.PLUS},
					},
					&ast.BinaryExpr{
						Left:     &ast.LiteralExpr{Value: 3},
						Right:    &ast.LiteralExpr{Value: 4},
						Operator: &token.Token{Type: token.STAR},
					},
					&ast.BinaryExpr{
						Left:     &ast.LiteralExpr{Value: 5},
						Right:    &ast.LiteralExpr{Value: 6},
						Operator: &token.Token{Type: token.PLUS},
					},
				},
			},
		}, []int{3, 12, 11}},
	}

	runVmTests(t, tests)
}

func TestHashLiterals(t *testing.T) {
	tests := []vmTestCase{
		// {}
		{&ast.ExprStmt{
			Expr: &ast.HashLiteralExpr{Pairs: []ast.HashPair{}},
		}, map[string]int{}},
		// {"one": 1, "two": 2}
		{&ast.ExprStmt{
			Expr: &ast.HashLiteralExpr{
				Pairs: []ast.HashPair{
					{Key: &ast.LiteralExpr{Value: "one"}, Value: &ast.LiteralExpr{Value: 1}},
					{Key: &ast.LiteralExpr{Value: "two"}, Value: &ast.LiteralExpr{Value: 2}},
				},
			},
		}, map[string]int{"one": 1, "two": 2}},
		// {"one": 1 + 1, "two": 2 * 2}
		{&ast.ExprStmt{
			Expr: &ast.HashLiteralExpr{
				Pairs: []ast.HashPair{
					{
						Key: &ast.LiteralExpr{Value: "one"},
						Value: &ast.BinaryExpr{
							Left:     &ast.LiteralExpr{Value: 1},
							Right:    &ast.LiteralExpr{Value: 1},
							Operator: &token.Token{Type: token.PLUS},
						},
					},
					{
						Key: &ast.LiteralExpr{Value: "two"},
						Value: &ast.BinaryExpr{
							Left:     &ast.LiteralExpr{Value: 2},
							Right:    &ast.LiteralExpr{Value: 2},
							Operator: &token.Token{Type: token.STAR},
						},
					},
				},
			},
		}, map[string]int{"one": 2, "two": 4}},
	}

	runVmTests(t, tests)
}

func TestIndexExpressions(t *testing.T) {
	tests := []vmTestCase{
		// [1, 2, 3][1]
		{&ast.ExprStmt{
			Expr: &ast.IndexExpr{
				Object: &ast.ArrayLiteralExpr{
					Elements: []ast.Expr{
						&ast.LiteralExpr{Value: 1},
						&ast.LiteralExpr{Value: 2},
						&ast.LiteralExpr{Value: 3},
					},
				},
				Index: &ast.LiteralExpr{Value: 1},
			},
		}, 2},
		// [1, 2, 3][0 + 2]
		{&ast.ExprStmt{
			Expr: &ast.IndexExpr{
				Object: &ast.ArrayLiteralExpr{
					Elements: []ast.Expr{
						&ast.LiteralExpr{Value: 1},
						&ast.LiteralExpr{Value: 2},
						&ast.LiteralExpr{Value: 3},
					},
				},
				Index: &ast.BinaryExpr{
					Left:     &ast.LiteralExpr{Value: 0},
					Right:    &ast.LiteralExpr{Value: 2},
					Operator: &token.Token{Type: token.PLUS},
				},
			},
		}, 3},
		// [[1, 2], [3, 4]][0][0]
		{&ast.ExprStmt{
			Expr: &ast.IndexExpr{
				Object: &ast.IndexExpr{
					Object: &ast.ArrayLiteralExpr{
						Elements: []ast.Expr{
							&ast.ArrayLiteralExpr{
								Elements: []ast.Expr{
									&ast.LiteralExpr{Value: 1},
									&ast.LiteralExpr{Value: 2},
								},
							},
							&ast.ArrayLiteralExpr{
								Elements: []ast.Expr{
									&ast.LiteralExpr{Value: 3},
									&ast.LiteralExpr{Value: 4},
								},
							},
						},
					},
					Index: &ast.LiteralExpr{Value: 0},
				},
				Index: &ast.LiteralExpr{Value: 0},
			},
		}, 1},
		// {"one": 1}["one"]
		{&ast.ExprStmt{
			Expr: &ast.IndexExpr{
				Object: &ast.HashLiteralExpr{
					Pairs: []ast.HashPair{
						{Key: &ast.LiteralExpr{Value: "one"}, Value: &ast.LiteralExpr{Value: 1}},
					},
				},
				Index: &ast.LiteralExpr{Value: "one"},
			},
		}, 1},
		// {"one": 1, "two": 2}["two"]
		{&ast.ExprStmt{
			Expr: &ast.IndexExpr{
				Object: &ast.HashLiteralExpr{
					Pairs: []ast.HashPair{
						{Key: &ast.LiteralExpr{Value: "one"}, Value: &ast.LiteralExpr{Value: 1}},
						{Key: &ast.LiteralExpr{Value: "two"}, Value: &ast.LiteralExpr{Value: 2}},
					},
				},
				Index: &ast.LiteralExpr{Value: "two"},
			},
		}, 2},
	}

	runVmTests(t, tests)
}

func TestIndexOutOfBoundsError(t *testing.T) {
	// [1, 2, 3][99] - out of bounds should error
	input := &ast.ExprStmt{
		Expr: &ast.IndexExpr{
			Object: &ast.ArrayLiteralExpr{
				Elements: []ast.Expr{
					&ast.LiteralExpr{Value: 1},
					&ast.LiteralExpr{Value: 2},
					&ast.LiteralExpr{Value: 3},
				},
			},
			Index: &ast.LiteralExpr{Value: 99},
		},
	}

	comp := compiler.New(nil)
	err := comp.Compile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(comp.Bytecode())
	err = vm.Run()
	if err == nil {
		t.Fatalf("expected error for out of bounds index, got none")
	}

	expected := "index out of bounds"
	if err.Error() != expected {
		t.Fatalf("wrong error. want=%q, got=%q", expected, err.Error())
	}
}

func TestHashKeyNotFoundError(t *testing.T) {
	// {"one": 1}["missing"] - missing key should error
	input := &ast.ExprStmt{
		Expr: &ast.IndexExpr{
			Object: &ast.HashLiteralExpr{
				Pairs: []ast.HashPair{
					{Key: &ast.LiteralExpr{Value: "one"}, Value: &ast.LiteralExpr{Value: 1}},
				},
			},
			Index: &ast.LiteralExpr{Value: "missing"},
		},
	}

	comp := compiler.New(nil)
	err := comp.Compile(input)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(comp.Bytecode())
	err = vm.Run()
	if err == nil {
		t.Fatalf("expected error for missing key, got none")
	}

	expected := "key 'missing' not found in hash map"
	if err.Error() != expected {
		t.Fatalf("wrong error. want=%q, got=%q", expected, err.Error())
	}
}

func TestSetIndexExpressions(t *testing.T) {
	tests := []vmTestCase{
		// var arr = [1, 2, 3]; arr[0] = 99; arr[0];
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.VarDeclStmt{
					Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "arr"},
					Initializer: &ast.ArrayLiteralExpr{
						Elements: []ast.Expr{
							&ast.LiteralExpr{Value: 1},
							&ast.LiteralExpr{Value: 2},
							&ast.LiteralExpr{Value: 3},
						},
					},
					IsConst: false,
				},
				&ast.ExprStmt{
					Expr: &ast.SetIndexExpr{
						Object: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "arr"},
						},
						Index: &ast.LiteralExpr{Value: 0},
						Value: &ast.LiteralExpr{Value: 99},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.IndexExpr{
						Object: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "arr"},
						},
						Index: &ast.LiteralExpr{Value: 0},
					},
				},
			},
		}, 99},
		// var hash = {"one": 1}; hash["two"] = 2; hash["two"];
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.VarDeclStmt{
					Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "hash"},
					Initializer: &ast.HashLiteralExpr{
						Pairs: []ast.HashPair{
							{Key: &ast.LiteralExpr{Value: "one"}, Value: &ast.LiteralExpr{Value: 1}},
						},
					},
					IsConst: false,
				},
				&ast.ExprStmt{
					Expr: &ast.SetIndexExpr{
						Object: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "hash"},
						},
						Index: &ast.LiteralExpr{Value: "two"},
						Value: &ast.LiteralExpr{Value: 2},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.IndexExpr{
						Object: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "hash"},
						},
						Index: &ast.LiteralExpr{Value: "two"},
					},
				},
			},
		}, 2},
	}

	runVmTests(t, tests)
}

func TestLogicalExpressions(t *testing.T) {
	tests := []vmTestCase{
		// true and true -> true (evaluates right)
		{&ast.ExprStmt{
			Expr: &ast.LogicalExpr{
				Left:     &ast.LiteralExpr{Value: true},
				Operator: &token.Token{Type: token.AND},
				Right:    &ast.LiteralExpr{Value: true},
			},
		}, true},
		// true and false -> false (evaluates right)
		{&ast.ExprStmt{
			Expr: &ast.LogicalExpr{
				Left:     &ast.LiteralExpr{Value: true},
				Operator: &token.Token{Type: token.AND},
				Right:    &ast.LiteralExpr{Value: false},
			},
		}, false},
		// false and true -> false (short-circuits, returns left)
		{&ast.ExprStmt{
			Expr: &ast.LogicalExpr{
				Left:     &ast.LiteralExpr{Value: false},
				Operator: &token.Token{Type: token.AND},
				Right:    &ast.LiteralExpr{Value: true},
			},
		}, false},
		// false and false -> false (short-circuits, returns left)
		{&ast.ExprStmt{
			Expr: &ast.LogicalExpr{
				Left:     &ast.LiteralExpr{Value: false},
				Operator: &token.Token{Type: token.AND},
				Right:    &ast.LiteralExpr{Value: false},
			},
		}, false},
		// true or true -> true (short-circuits, returns left)
		{&ast.ExprStmt{
			Expr: &ast.LogicalExpr{
				Left:     &ast.LiteralExpr{Value: true},
				Operator: &token.Token{Type: token.OR},
				Right:    &ast.LiteralExpr{Value: true},
			},
		}, true},
		// true or false -> true (short-circuits, returns left)
		{&ast.ExprStmt{
			Expr: &ast.LogicalExpr{
				Left:     &ast.LiteralExpr{Value: true},
				Operator: &token.Token{Type: token.OR},
				Right:    &ast.LiteralExpr{Value: false},
			},
		}, true},
		// false or true -> true (evaluates right)
		{&ast.ExprStmt{
			Expr: &ast.LogicalExpr{
				Left:     &ast.LiteralExpr{Value: false},
				Operator: &token.Token{Type: token.OR},
				Right:    &ast.LiteralExpr{Value: true},
			},
		}, true},
		// false or false -> false (evaluates right)
		{&ast.ExprStmt{
			Expr: &ast.LogicalExpr{
				Left:     &ast.LiteralExpr{Value: false},
				Operator: &token.Token{Type: token.OR},
				Right:    &ast.LiteralExpr{Value: false},
			},
		}, false},

		// 0 and 42 -> 0 (0 is falsy, returns left value)
		{&ast.ExprStmt{
			Expr: &ast.LogicalExpr{
				Left:     &ast.LiteralExpr{Value: 0},
				Operator: &token.Token{Type: token.AND},
				Right:    &ast.LiteralExpr{Value: 42},
			},
		}, 0},
		// 5 and 42 -> 42 (5 is truthy, returns right value)
		{&ast.ExprStmt{
			Expr: &ast.LogicalExpr{
				Left:     &ast.LiteralExpr{Value: 5},
				Operator: &token.Token{Type: token.AND},
				Right:    &ast.LiteralExpr{Value: 42},
			},
		}, 42},
		// 0 or 42 -> 42 (0 is falsy, returns right value)
		{&ast.ExprStmt{
			Expr: &ast.LogicalExpr{
				Left:     &ast.LiteralExpr{Value: 0},
				Operator: &token.Token{Type: token.OR},
				Right:    &ast.LiteralExpr{Value: 42},
			},
		}, 42},
		// 5 or 42 -> 5 (5 is truthy, returns left value)
		{&ast.ExprStmt{
			Expr: &ast.LogicalExpr{
				Left:     &ast.LiteralExpr{Value: 5},
				Operator: &token.Token{Type: token.OR},
				Right:    &ast.LiteralExpr{Value: 42},
			},
		}, 5},

		// "" and "hello" -> "" (empty string is falsy)
		{&ast.ExprStmt{
			Expr: &ast.LogicalExpr{
				Left:     &ast.LiteralExpr{Value: ""},
				Operator: &token.Token{Type: token.AND},
				Right:    &ast.LiteralExpr{Value: "hello"},
			},
		}, ""},
		// "yes" and "no" -> "no" (both truthy, returns right)
		{&ast.ExprStmt{
			Expr: &ast.LogicalExpr{
				Left:     &ast.LiteralExpr{Value: "yes"},
				Operator: &token.Token{Type: token.AND},
				Right:    &ast.LiteralExpr{Value: "no"},
			},
		}, "no"},
		// "" or "hello" -> "hello" (empty string is falsy)
		{&ast.ExprStmt{
			Expr: &ast.LogicalExpr{
				Left:     &ast.LiteralExpr{Value: ""},
				Operator: &token.Token{Type: token.OR},
				Right:    &ast.LiteralExpr{Value: "hello"},
			},
		}, "hello"},
		// "yes" or "no" -> "yes" (truthy, returns left)
		{&ast.ExprStmt{
			Expr: &ast.LogicalExpr{
				Left:     &ast.LiteralExpr{Value: "yes"},
				Operator: &token.Token{Type: token.OR},
				Right:    &ast.LiteralExpr{Value: "no"},
			},
		}, "yes"},

		// (1 < 2) and (3 < 4) -> true
		{&ast.ExprStmt{
			Expr: &ast.LogicalExpr{
				Left: &ast.BinaryExpr{
					Left:     &ast.LiteralExpr{Value: 1},
					Right:    &ast.LiteralExpr{Value: 2},
					Operator: &token.Token{Type: token.LESS},
				},
				Operator: &token.Token{Type: token.AND},
				Right: &ast.BinaryExpr{
					Left:     &ast.LiteralExpr{Value: 3},
					Right:    &ast.LiteralExpr{Value: 4},
					Operator: &token.Token{Type: token.LESS},
				},
			},
		}, true},
		// (1 > 2) or (3 < 4) -> true
		{&ast.ExprStmt{
			Expr: &ast.LogicalExpr{
				Left: &ast.BinaryExpr{
					Left:     &ast.LiteralExpr{Value: 1},
					Right:    &ast.LiteralExpr{Value: 2},
					Operator: &token.Token{Type: token.GREATER},
				},
				Operator: &token.Token{Type: token.OR},
				Right: &ast.BinaryExpr{
					Left:     &ast.LiteralExpr{Value: 3},
					Right:    &ast.LiteralExpr{Value: 4},
					Operator: &token.Token{Type: token.LESS},
				},
			},
		}, true},
	}

	runVmTests(t, tests)
}

func TestWhileStatements(t *testing.T) {
	tests := []vmTestCase{
		// var x = 0; while (x < 3) { x = x + 1; } x;
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.VarDeclStmt{
					Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "x"},
					Initializer: &ast.LiteralExpr{Value: 0},
					IsConst:     false,
				},
				&ast.WhileStmt{
					Condition: &ast.BinaryExpr{
						Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "x"}},
						Right:    &ast.LiteralExpr{Value: 3},
						Operator: &token.Token{Type: token.LESS},
					},
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.ExprStmt{
								Expr: &ast.AssignExpr{
									Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "x"},
									Value: &ast.BinaryExpr{
										Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "x"}},
										Right:    &ast.LiteralExpr{Value: 1},
										Operator: &token.Token{Type: token.PLUS},
									},
								},
							},
						},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "x"}},
				},
			},
		}, 3},
		// var sum = 0; var i = 1; while (i < 5) { sum = sum + i; i = i + 1; } sum;
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.VarDeclStmt{
					Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"},
					Initializer: &ast.LiteralExpr{Value: 0},
					IsConst:     false,
				},
				&ast.VarDeclStmt{
					Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "i"},
					Initializer: &ast.LiteralExpr{Value: 1},
					IsConst:     false,
				},
				&ast.WhileStmt{
					Condition: &ast.BinaryExpr{
						Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
						Right:    &ast.LiteralExpr{Value: 5},
						Operator: &token.Token{Type: token.LESS},
					},
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.ExprStmt{
								Expr: &ast.AssignExpr{
									Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"},
									Value: &ast.BinaryExpr{
										Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"}},
										Right:    &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
										Operator: &token.Token{Type: token.PLUS},
									},
								},
							},
							&ast.ExprStmt{
								Expr: &ast.AssignExpr{
									Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"},
									Value: &ast.BinaryExpr{
										Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
										Right:    &ast.LiteralExpr{Value: 1},
										Operator: &token.Token{Type: token.PLUS},
									},
								},
							},
						},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"}},
				},
			},
		}, 10}, // 1 + 2 + 3 + 4 = 10
	}

	runVmTests(t, tests)
}

func TestForStatements(t *testing.T) {
	tests := []vmTestCase{
		// Full for loop: for (var i = 0; i < 5; i = i + 1) { sum = sum + i; }
		// var sum = 0; for (var i = 0; i < 5; i = i + 1) { sum = sum + i; } sum;
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.VarDeclStmt{
					Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"},
					Initializer: &ast.LiteralExpr{Value: 0},
					IsConst:     false,
				},
				&ast.ForStmt{
					Initializer: &ast.VarDeclStmt{
						Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "i"},
						Initializer: &ast.LiteralExpr{Value: 0},
						IsConst:     false,
					},
					Condition: &ast.BinaryExpr{
						Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
						Right:    &ast.LiteralExpr{Value: 5},
						Operator: &token.Token{Type: token.LESS},
					},
					Increment: &ast.AssignExpr{
						Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"},
						Value: &ast.BinaryExpr{
							Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
							Right:    &ast.LiteralExpr{Value: 1},
							Operator: &token.Token{Type: token.PLUS},
						},
					},
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.ExprStmt{
								Expr: &ast.AssignExpr{
									Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"},
									Value: &ast.BinaryExpr{
										Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"}},
										Right:    &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
										Operator: &token.Token{Type: token.PLUS},
									},
								},
							},
						},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"}},
				},
			},
		}, 10}, // 0 + 1 + 2 + 3 + 4 = 10

		// No initializer: var i = 0; for (; i < 5; i = i + 1) { sum = sum + i; }
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.VarDeclStmt{
					Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"},
					Initializer: &ast.LiteralExpr{Value: 0},
					IsConst:     false,
				},
				&ast.VarDeclStmt{
					Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "i"},
					Initializer: &ast.LiteralExpr{Value: 0},
					IsConst:     false,
				},
				&ast.ForStmt{
					Initializer: nil, // No initializer
					Condition: &ast.BinaryExpr{
						Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
						Right:    &ast.LiteralExpr{Value: 5},
						Operator: &token.Token{Type: token.LESS},
					},
					Increment: &ast.AssignExpr{
						Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"},
						Value: &ast.BinaryExpr{
							Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
							Right:    &ast.LiteralExpr{Value: 1},
							Operator: &token.Token{Type: token.PLUS},
						},
					},
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.ExprStmt{
								Expr: &ast.AssignExpr{
									Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"},
									Value: &ast.BinaryExpr{
										Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"}},
										Right:    &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
										Operator: &token.Token{Type: token.PLUS},
									},
								},
							},
						},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"}},
				},
			},
		}, 10}, // 0 + 1 + 2 + 3 + 4 = 10

		// No increment: for (var i = 0; i < 5;) { sum = sum + i; i = i + 1; }
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.VarDeclStmt{
					Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"},
					Initializer: &ast.LiteralExpr{Value: 0},
					IsConst:     false,
				},
				&ast.ForStmt{
					Initializer: &ast.VarDeclStmt{
						Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "i"},
						Initializer: &ast.LiteralExpr{Value: 0},
						IsConst:     false,
					},
					Condition: &ast.BinaryExpr{
						Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
						Right:    &ast.LiteralExpr{Value: 5},
						Operator: &token.Token{Type: token.LESS},
					},
					Increment: nil, // No increment
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.ExprStmt{
								Expr: &ast.AssignExpr{
									Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"},
									Value: &ast.BinaryExpr{
										Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"}},
										Right:    &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
										Operator: &token.Token{Type: token.PLUS},
									},
								},
							},
							&ast.ExprStmt{
								Expr: &ast.AssignExpr{
									Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"},
									Value: &ast.BinaryExpr{
										Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
										Right:    &ast.LiteralExpr{Value: 1},
										Operator: &token.Token{Type: token.PLUS},
									},
								},
							},
						},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"}},
				},
			},
		}, 10}, // 0 + 1 + 2 + 3 + 4 = 10

		// No initializer and no increment: var i = 0; for (; i < 5;) { ... }
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.VarDeclStmt{
					Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"},
					Initializer: &ast.LiteralExpr{Value: 0},
					IsConst:     false,
				},
				&ast.VarDeclStmt{
					Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "i"},
					Initializer: &ast.LiteralExpr{Value: 0},
					IsConst:     false,
				},
				&ast.ForStmt{
					Initializer: nil, // No initializer
					Condition: &ast.BinaryExpr{
						Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
						Right:    &ast.LiteralExpr{Value: 5},
						Operator: &token.Token{Type: token.LESS},
					},
					Increment: nil, // No increment
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.ExprStmt{
								Expr: &ast.AssignExpr{
									Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"},
									Value: &ast.BinaryExpr{
										Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"}},
										Right:    &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
										Operator: &token.Token{Type: token.PLUS},
									},
								},
							},
							&ast.ExprStmt{
								Expr: &ast.AssignExpr{
									Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"},
									Value: &ast.BinaryExpr{
										Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
										Right:    &ast.LiteralExpr{Value: 1},
										Operator: &token.Token{Type: token.PLUS},
									},
								},
							},
						},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"}},
				},
			},
		}, 10}, // 0 + 1 + 2 + 3 + 4 = 10

		// No condition (infinite loop with break): for (var i = 0;; i = i + 1) { if (i == 5) break; sum = sum + i; }
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.VarDeclStmt{
					Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"},
					Initializer: &ast.LiteralExpr{Value: 0},
					IsConst:     false,
				},
				&ast.ForStmt{
					Initializer: &ast.VarDeclStmt{
						Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "i"},
						Initializer: &ast.LiteralExpr{Value: 0},
						IsConst:     false,
					},
					Condition: nil, // No condition - infinite loop
					Increment: &ast.AssignExpr{
						Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"},
						Value: &ast.BinaryExpr{
							Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
							Right:    &ast.LiteralExpr{Value: 1},
							Operator: &token.Token{Type: token.PLUS},
						},
					},
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.IfStmt{
								Condition: &ast.BinaryExpr{
									Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
									Right:    &ast.LiteralExpr{Value: 5},
									Operator: &token.Token{Type: token.EQUAL_EQUAL},
								},
								ThenBranch: &ast.BlockStmt{
									Statements: []ast.Stmt{
										&ast.BreakStmt{Keyword: &token.Token{Type: token.BREAK}},
									},
								},
							},
							&ast.ExprStmt{
								Expr: &ast.AssignExpr{
									Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"},
									Value: &ast.BinaryExpr{
										Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"}},
										Right:    &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
										Operator: &token.Token{Type: token.PLUS},
									},
								},
							},
						},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"}},
				},
			},
		}, 10}, // 0 + 1 + 2 + 3 + 4 = 10 (breaks when i == 5)

		// All parts missing (infinite loop): for (;;) { if (x == 5) break; x = x + 1; }
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.VarDeclStmt{
					Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "x"},
					Initializer: &ast.LiteralExpr{Value: 0},
					IsConst:     false,
				},
				&ast.ForStmt{
					Initializer: nil, // No initializer
					Condition:   nil, // No condition
					Increment:   nil, // No increment
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.IfStmt{
								Condition: &ast.BinaryExpr{
									Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "x"}},
									Right:    &ast.LiteralExpr{Value: 5},
									Operator: &token.Token{Type: token.EQUAL_EQUAL},
								},
								ThenBranch: &ast.BlockStmt{
									Statements: []ast.Stmt{
										&ast.BreakStmt{Keyword: &token.Token{Type: token.BREAK}},
									},
								},
							},
							&ast.ExprStmt{
								Expr: &ast.AssignExpr{
									Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "x"},
									Value: &ast.BinaryExpr{
										Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "x"}},
										Right:    &ast.LiteralExpr{Value: 1},
										Operator: &token.Token{Type: token.PLUS},
									},
								},
							},
						},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "x"}},
				},
			},
		}, 5}, // x increments until it equals 5, then breaks

		// Continue in for loop with increment - should still run increment
		// for (var i = 0; i < 10; i = i + 1) { if (i == 3) continue; if (i == 7) break; sum = sum + i; }
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.VarDeclStmt{
					Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"},
					Initializer: &ast.LiteralExpr{Value: 0},
					IsConst:     false,
				},
				&ast.ForStmt{
					Initializer: &ast.VarDeclStmt{
						Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "i"},
						Initializer: &ast.LiteralExpr{Value: 0},
						IsConst:     false,
					},
					Condition: &ast.BinaryExpr{
						Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
						Right:    &ast.LiteralExpr{Value: 10},
						Operator: &token.Token{Type: token.LESS},
					},
					Increment: &ast.AssignExpr{
						Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"},
						Value: &ast.BinaryExpr{
							Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
							Right:    &ast.LiteralExpr{Value: 1},
							Operator: &token.Token{Type: token.PLUS},
						},
					},
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							// if (i == 3) continue;
							&ast.IfStmt{
								Condition: &ast.BinaryExpr{
									Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
									Right:    &ast.LiteralExpr{Value: 3},
									Operator: &token.Token{Type: token.EQUAL_EQUAL},
								},
								ThenBranch: &ast.BlockStmt{
									Statements: []ast.Stmt{
										&ast.ContinueStmt{Keyword: &token.Token{Type: token.CONTINUE}},
									},
								},
							},
							// if (i == 7) break;
							&ast.IfStmt{
								Condition: &ast.BinaryExpr{
									Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
									Right:    &ast.LiteralExpr{Value: 7},
									Operator: &token.Token{Type: token.EQUAL_EQUAL},
								},
								ThenBranch: &ast.BlockStmt{
									Statements: []ast.Stmt{
										&ast.BreakStmt{Keyword: &token.Token{Type: token.BREAK}},
									},
								},
							},
							// sum = sum + i;
							&ast.ExprStmt{
								Expr: &ast.AssignExpr{
									Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"},
									Value: &ast.BinaryExpr{
										Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"}},
										Right:    &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
										Operator: &token.Token{Type: token.PLUS},
									},
								},
							},
						},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"}},
				},
			},
		}, 18}, // 0 + 1 + 2 + 4 + 5 + 6 = 18 (skips 3, breaks at 7)
	}

	runVmTests(t, tests)
}

func TestBreakStatement(t *testing.T) {
	tests := []vmTestCase{
		// var x = 0; while (true) { x = x + 1; if (x == 5) { break; } } x;
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.VarDeclStmt{
					Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "x"},
					Initializer: &ast.LiteralExpr{Value: 0},
					IsConst:     false,
				},
				&ast.WhileStmt{
					Condition: &ast.LiteralExpr{Value: true},
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.ExprStmt{
								Expr: &ast.AssignExpr{
									Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "x"},
									Value: &ast.BinaryExpr{
										Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "x"}},
										Right:    &ast.LiteralExpr{Value: 1},
										Operator: &token.Token{Type: token.PLUS},
									},
								},
							},
							&ast.IfStmt{
								Condition: &ast.BinaryExpr{
									Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "x"}},
									Right:    &ast.LiteralExpr{Value: 5},
									Operator: &token.Token{Type: token.EQUAL_EQUAL},
								},
								ThenBranch: &ast.BlockStmt{
									Statements: []ast.Stmt{
										&ast.BreakStmt{Keyword: &token.Token{Type: token.BREAK}},
									},
								},
							},
						},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "x"}},
				},
			},
		}, 5},
		// var sum = 0; for (var i = 0; i < 100; i = i + 1) { if (i == 5) { break; } sum = sum + i; } sum;
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.VarDeclStmt{
					Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"},
					Initializer: &ast.LiteralExpr{Value: 0},
					IsConst:     false,
				},
				&ast.ForStmt{
					Initializer: &ast.VarDeclStmt{
						Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "i"},
						Initializer: &ast.LiteralExpr{Value: 0},
						IsConst:     false,
					},
					Condition: &ast.BinaryExpr{
						Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
						Right:    &ast.LiteralExpr{Value: 100},
						Operator: &token.Token{Type: token.LESS},
					},
					Increment: &ast.AssignExpr{
						Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"},
						Value: &ast.BinaryExpr{
							Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
							Right:    &ast.LiteralExpr{Value: 1},
							Operator: &token.Token{Type: token.PLUS},
						},
					},
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.IfStmt{
								Condition: &ast.BinaryExpr{
									Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
									Right:    &ast.LiteralExpr{Value: 5},
									Operator: &token.Token{Type: token.EQUAL_EQUAL},
								},
								ThenBranch: &ast.BlockStmt{
									Statements: []ast.Stmt{
										&ast.BreakStmt{Keyword: &token.Token{Type: token.BREAK}},
									},
								},
							},
							&ast.ExprStmt{
								Expr: &ast.AssignExpr{
									Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"},
									Value: &ast.BinaryExpr{
										Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"}},
										Right:    &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
										Operator: &token.Token{Type: token.PLUS},
									},
								},
							},
						},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"}},
				},
			},
		}, 10}, // 0 + 1 + 2 + 3 + 4 = 10 (breaks when i == 5)
	}

	runVmTests(t, tests)
}

func TestContinueStatement(t *testing.T) {
	tests := []vmTestCase{
		// var sum = 0; for (var i = 0; i < 10; i = i + 1) { if (i == 5) { continue; } sum = sum + i; } sum;
		// Should skip adding 5, so sum = 0+1+2+3+4+6+7+8+9 = 40
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.VarDeclStmt{
					Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"},
					Initializer: &ast.LiteralExpr{Value: 0},
					IsConst:     false,
				},
				&ast.ForStmt{
					Initializer: &ast.VarDeclStmt{
						Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "i"},
						Initializer: &ast.LiteralExpr{Value: 0},
						IsConst:     false,
					},
					Condition: &ast.BinaryExpr{
						Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
						Right:    &ast.LiteralExpr{Value: 10},
						Operator: &token.Token{Type: token.LESS},
					},
					Increment: &ast.AssignExpr{
						Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"},
						Value: &ast.BinaryExpr{
							Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
							Right:    &ast.LiteralExpr{Value: 1},
							Operator: &token.Token{Type: token.PLUS},
						},
					},
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.IfStmt{
								Condition: &ast.BinaryExpr{
									Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
									Right:    &ast.LiteralExpr{Value: 5},
									Operator: &token.Token{Type: token.EQUAL_EQUAL},
								},
								ThenBranch: &ast.BlockStmt{
									Statements: []ast.Stmt{
										&ast.ContinueStmt{Keyword: &token.Token{Type: token.CONTINUE}},
									},
								},
							},
							&ast.ExprStmt{
								Expr: &ast.AssignExpr{
									Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"},
									Value: &ast.BinaryExpr{
										Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"}},
										Right:    &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
										Operator: &token.Token{Type: token.PLUS},
									},
								},
							},
						},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"}},
				},
			},
		}, 40}, // 0+1+2+3+4+6+7+8+9 = 40 (skips 5)
		// var x = 0; var i = 0; while (i < 10) { i = i + 1; if (i == 5) { continue; } x = x + i; } x;
		// Should skip adding 5, so x = 1+2+3+4+6+7+8+9+10 = 50
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.VarDeclStmt{
					Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "x"},
					Initializer: &ast.LiteralExpr{Value: 0},
					IsConst:     false,
				},
				&ast.VarDeclStmt{
					Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "i"},
					Initializer: &ast.LiteralExpr{Value: 0},
					IsConst:     false,
				},
				&ast.WhileStmt{
					Condition: &ast.BinaryExpr{
						Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
						Right:    &ast.LiteralExpr{Value: 10},
						Operator: &token.Token{Type: token.LESS},
					},
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.ExprStmt{
								Expr: &ast.AssignExpr{
									Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"},
									Value: &ast.BinaryExpr{
										Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
										Right:    &ast.LiteralExpr{Value: 1},
										Operator: &token.Token{Type: token.PLUS},
									},
								},
							},
							&ast.IfStmt{
								Condition: &ast.BinaryExpr{
									Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
									Right:    &ast.LiteralExpr{Value: 5},
									Operator: &token.Token{Type: token.EQUAL_EQUAL},
								},
								ThenBranch: &ast.BlockStmt{
									Statements: []ast.Stmt{
										&ast.ContinueStmt{Keyword: &token.Token{Type: token.CONTINUE}},
									},
								},
							},
							&ast.ExprStmt{
								Expr: &ast.AssignExpr{
									Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "x"},
									Value: &ast.BinaryExpr{
										Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "x"}},
										Right:    &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
										Operator: &token.Token{Type: token.PLUS},
									},
								},
							},
						},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "x"}},
				},
			},
		}, 50}, // 1+2+3+4+6+7+8+9+10 = 50 (skips 5)
	}

	runVmTests(t, tests)
}

func TestCallingFunctionsWithoutArguments(t *testing.T) {
	tests := []vmTestCase{
		// fun fivePlusTen() { return 5 + 10; } fivePlusTen();
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.FunctionStmt{
					Name:   &token.Token{Type: token.IDENTIFIER, Lexeme: "fivePlusTen"},
					Params: []*token.Token{},
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.ReturnStmt{
								Keyword: &token.Token{Type: token.RETURN},
								Value: &ast.BinaryExpr{
									Left:     &ast.LiteralExpr{Value: 5},
									Right:    &ast.LiteralExpr{Value: 10},
									Operator: &token.Token{Type: token.PLUS},
								},
							},
						},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.CallExpr{
						Callee: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "fivePlusTen"},
						},
						Arguments: []ast.Expr{},
					},
				},
			},
		}, 15},
		// fun one() { return 1; } fun two() { return 2; } one() + two();
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.FunctionStmt{
					Name:   &token.Token{Type: token.IDENTIFIER, Lexeme: "one"},
					Params: []*token.Token{},
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.ReturnStmt{
								Keyword: &token.Token{Type: token.RETURN},
								Value:   &ast.LiteralExpr{Value: 1},
							},
						},
					},
				},
				&ast.FunctionStmt{
					Name:   &token.Token{Type: token.IDENTIFIER, Lexeme: "two"},
					Params: []*token.Token{},
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.ReturnStmt{
								Keyword: &token.Token{Type: token.RETURN},
								Value:   &ast.LiteralExpr{Value: 2},
							},
						},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.BinaryExpr{
						Left: &ast.CallExpr{
							Callee: &ast.VariableExpr{
								Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "one"},
							},
							Arguments: []ast.Expr{},
						},
						Right: &ast.CallExpr{
							Callee: &ast.VariableExpr{
								Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "two"},
							},
							Arguments: []ast.Expr{},
						},
						Operator: &token.Token{Type: token.PLUS},
					},
				},
			},
		}, 3},
	}

	runVmTests(t, tests)
}

func TestFunctionsWithReturnStatement(t *testing.T) {
	tests := []vmTestCase{
		// fun earlyExit() { return 99; 100; } earlyExit();
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.FunctionStmt{
					Name:   &token.Token{Type: token.IDENTIFIER, Lexeme: "earlyExit"},
					Params: []*token.Token{},
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.ReturnStmt{
								Keyword: &token.Token{Type: token.RETURN},
								Value:   &ast.LiteralExpr{Value: 99},
							},
							&ast.ExprStmt{Expr: &ast.LiteralExpr{Value: 100}},
						},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.CallExpr{
						Callee: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "earlyExit"},
						},
						Arguments: []ast.Expr{},
					},
				},
			},
		}, 99},
	}

	runVmTests(t, tests)
}

func TestFunctionsWithoutReturnValue(t *testing.T) {
	tests := []vmTestCase{
		// fun noReturn() { } noReturn();
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.FunctionStmt{
					Name:   &token.Token{Type: token.IDENTIFIER, Lexeme: "noReturn"},
					Params: []*token.Token{},
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.CallExpr{
						Callee: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "noReturn"},
						},
						Arguments: []ast.Expr{},
					},
				},
			},
		}, nil},
		// fun noReturn() { return; } noReturn();
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.FunctionStmt{
					Name:   &token.Token{Type: token.IDENTIFIER, Lexeme: "noReturn"},
					Params: []*token.Token{},
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.ReturnStmt{
								Keyword: &token.Token{Type: token.RETURN},
								Value:   nil,
							},
						},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.CallExpr{
						Callee: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "noReturn"},
						},
						Arguments: []ast.Expr{},
					},
				},
			},
		}, nil},
	}

	runVmTests(t, tests)
}

func TestCallingFunctionsWithBindings(t *testing.T) {
	tests := []vmTestCase{
		// fun one() { var one = 1; return one; } one();
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.FunctionStmt{
					Name:   &token.Token{Type: token.IDENTIFIER, Lexeme: "one"},
					Params: []*token.Token{},
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.VarDeclStmt{
								Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "one"},
								Initializer: &ast.LiteralExpr{Value: 1},
								IsConst:     false,
							},
							&ast.ReturnStmt{
								Keyword: &token.Token{Type: token.RETURN},
								Value: &ast.VariableExpr{
									Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "one"},
								},
							},
						},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.CallExpr{
						Callee: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "one"},
						},
						Arguments: []ast.Expr{},
					},
				},
			},
		}, 1},
		// fun oneAndTwo() { var one = 1; var two = 2; return one + two; } oneAndTwo();
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.FunctionStmt{
					Name:   &token.Token{Type: token.IDENTIFIER, Lexeme: "oneAndTwo"},
					Params: []*token.Token{},
					Body: &ast.BlockStmt{
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
							&ast.ReturnStmt{
								Keyword: &token.Token{Type: token.RETURN},
								Value: &ast.BinaryExpr{
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
					},
				},
				&ast.ExprStmt{
					Expr: &ast.CallExpr{
						Callee: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "oneAndTwo"},
						},
						Arguments: []ast.Expr{},
					},
				},
			},
		}, 3},
	}

	runVmTests(t, tests)
}

func TestCallingFunctionsWithArguments(t *testing.T) {
	tests := []vmTestCase{
		// fun identity(a) { return a; } identity(4);
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.FunctionStmt{
					Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "identity"},
					Params: []*token.Token{
						{Type: token.IDENTIFIER, Lexeme: "a"},
					},
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.ReturnStmt{
								Keyword: &token.Token{Type: token.RETURN},
								Value: &ast.VariableExpr{
									Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "a"},
								},
							},
						},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.CallExpr{
						Callee: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "identity"},
						},
						Arguments: []ast.Expr{
							&ast.LiteralExpr{Value: 4},
						},
					},
				},
			},
		}, 4},
		// fun sum(a, b) { return a + b; } sum(1, 2);
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.FunctionStmt{
					Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"},
					Params: []*token.Token{
						{Type: token.IDENTIFIER, Lexeme: "a"},
						{Type: token.IDENTIFIER, Lexeme: "b"},
					},
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.ReturnStmt{
								Keyword: &token.Token{Type: token.RETURN},
								Value: &ast.BinaryExpr{
									Left: &ast.VariableExpr{
										Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "a"},
									},
									Right: &ast.VariableExpr{
										Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "b"},
									},
									Operator: &token.Token{Type: token.PLUS},
								},
							},
						},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.CallExpr{
						Callee: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"},
						},
						Arguments: []ast.Expr{
							&ast.LiteralExpr{Value: 1},
							&ast.LiteralExpr{Value: 2},
						},
					},
				},
			},
		}, 3},
		// fun sum(a, b) { var c = a + b; return c; } sum(1, 2);
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.FunctionStmt{
					Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"},
					Params: []*token.Token{
						{Type: token.IDENTIFIER, Lexeme: "a"},
						{Type: token.IDENTIFIER, Lexeme: "b"},
					},
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.VarDeclStmt{
								Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "c"},
								Initializer: &ast.BinaryExpr{
									Left: &ast.VariableExpr{
										Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "a"},
									},
									Right: &ast.VariableExpr{
										Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "b"},
									},
									Operator: &token.Token{Type: token.PLUS},
								},
								IsConst: false,
							},
							&ast.ReturnStmt{
								Keyword: &token.Token{Type: token.RETURN},
								Value: &ast.VariableExpr{
									Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "c"},
								},
							},
						},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.CallExpr{
						Callee: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sum"},
						},
						Arguments: []ast.Expr{
							&ast.LiteralExpr{Value: 1},
							&ast.LiteralExpr{Value: 2},
						},
					},
				},
			},
		}, 3},
	}

	runVmTests(t, tests)
}

func TestCallingFunctionsWithArgumentsAndBindings(t *testing.T) {
	tests := []vmTestCase{
		// fun sumPlusGlobal(a, b) { var c = a + b; return c + globalNum; } var globalNum = 10; sumPlusGlobal(1, 2);
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.VarDeclStmt{
					Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "globalNum"},
					Initializer: &ast.LiteralExpr{Value: 10},
					IsConst:     false,
				},
				&ast.FunctionStmt{
					Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sumPlusGlobal"},
					Params: []*token.Token{
						{Type: token.IDENTIFIER, Lexeme: "a"},
						{Type: token.IDENTIFIER, Lexeme: "b"},
					},
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.VarDeclStmt{
								Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "c"},
								Initializer: &ast.BinaryExpr{
									Left: &ast.VariableExpr{
										Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "a"},
									},
									Right: &ast.VariableExpr{
										Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "b"},
									},
									Operator: &token.Token{Type: token.PLUS},
								},
								IsConst: false,
							},
							&ast.ReturnStmt{
								Keyword: &token.Token{Type: token.RETURN},
								Value: &ast.BinaryExpr{
									Left: &ast.VariableExpr{
										Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "c"},
									},
									Right: &ast.VariableExpr{
										Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "globalNum"},
									},
									Operator: &token.Token{Type: token.PLUS},
								},
							},
						},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.CallExpr{
						Callee: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "sumPlusGlobal"},
						},
						Arguments: []ast.Expr{
							&ast.LiteralExpr{Value: 1},
							&ast.LiteralExpr{Value: 2},
						},
					},
				},
			},
		}, 13}, // 1 + 2 + 10 = 13
	}

	runVmTests(t, tests)
}

func TestFirstClassFunctions(t *testing.T) {
	tests := []vmTestCase{
		// var returnsOne = fun() { return 1; }; returnsOne();
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.VarDeclStmt{
					Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "returnsOne"},
					Initializer: &ast.FunctionExpr{
						Params: []*token.Token{},
						Body: &ast.BlockStmt{
							Statements: []ast.Stmt{
								&ast.ReturnStmt{
									Keyword: &token.Token{Type: token.RETURN},
									Value:   &ast.LiteralExpr{Value: 1},
								},
							},
						},
					},
					IsConst: false,
				},
				&ast.ExprStmt{
					Expr: &ast.CallExpr{
						Callee: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "returnsOne"},
						},
						Arguments: []ast.Expr{},
					},
				},
			},
		}, 1},
		// var returnsOneReturner = fun() { var returnsOne = fun() { return 1; }; return returnsOne; }; returnsOneReturner()();
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.VarDeclStmt{
					Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "returnsOneReturner"},
					Initializer: &ast.FunctionExpr{
						Params: []*token.Token{},
						Body: &ast.BlockStmt{
							Statements: []ast.Stmt{
								&ast.VarDeclStmt{
									Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "returnsOne"},
									Initializer: &ast.FunctionExpr{
										Params: []*token.Token{},
										Body: &ast.BlockStmt{
											Statements: []ast.Stmt{
												&ast.ReturnStmt{
													Keyword: &token.Token{Type: token.RETURN},
													Value:   &ast.LiteralExpr{Value: 1},
												},
											},
										},
									},
									IsConst: false,
								},
								&ast.ReturnStmt{
									Keyword: &token.Token{Type: token.RETURN},
									Value: &ast.VariableExpr{
										Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "returnsOne"},
									},
								},
							},
						},
					},
					IsConst: false,
				},
				&ast.ExprStmt{
					Expr: &ast.CallExpr{
						Callee: &ast.CallExpr{
							Callee: &ast.VariableExpr{
								Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "returnsOneReturner"},
							},
							Arguments: []ast.Expr{},
						},
						Arguments: []ast.Expr{},
					},
				},
			},
		}, 1},
	}

	runVmTests(t, tests)
}

func TestRecursiveFunctions(t *testing.T) {
	tests := []vmTestCase{
		// fun countDown(x) { if (x == 0) { return 0; } return countDown(x - 1); } countDown(1);
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.FunctionStmt{
					Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "countDown"},
					Params: []*token.Token{
						{Type: token.IDENTIFIER, Lexeme: "x"},
					},
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.IfStmt{
								Condition: &ast.BinaryExpr{
									Left: &ast.VariableExpr{
										Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "x"},
									},
									Right:    &ast.LiteralExpr{Value: 0},
									Operator: &token.Token{Type: token.EQUAL_EQUAL},
								},
								ThenBranch: &ast.BlockStmt{
									Statements: []ast.Stmt{
										&ast.ReturnStmt{
											Keyword: &token.Token{Type: token.RETURN},
											Value:   &ast.LiteralExpr{Value: 0},
										},
									},
								},
							},
							&ast.ReturnStmt{
								Keyword: &token.Token{Type: token.RETURN},
								Value: &ast.CallExpr{
									Callee: &ast.VariableExpr{
										Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "countDown"},
									},
									Arguments: []ast.Expr{
										&ast.BinaryExpr{
											Left: &ast.VariableExpr{
												Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "x"},
											},
											Right:    &ast.LiteralExpr{Value: 1},
											Operator: &token.Token{Type: token.MINUS},
										},
									},
								},
							},
						},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.CallExpr{
						Callee: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "countDown"},
						},
						Arguments: []ast.Expr{
							&ast.LiteralExpr{Value: 1},
						},
					},
				},
			},
		}, 0},
		// Fibonacci: fun fib(n) { if (n < 2) { return n; } return fib(n - 1) + fib(n - 2); } fib(15);
		{&ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.FunctionStmt{
					Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "fib"},
					Params: []*token.Token{
						{Type: token.IDENTIFIER, Lexeme: "n"},
					},
					Body: &ast.BlockStmt{
						Statements: []ast.Stmt{
							&ast.IfStmt{
								Condition: &ast.BinaryExpr{
									Left: &ast.VariableExpr{
										Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "n"},
									},
									Right:    &ast.LiteralExpr{Value: 2},
									Operator: &token.Token{Type: token.LESS},
								},
								ThenBranch: &ast.BlockStmt{
									Statements: []ast.Stmt{
										&ast.ReturnStmt{
											Keyword: &token.Token{Type: token.RETURN},
											Value: &ast.VariableExpr{
												Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "n"},
											},
										},
									},
								},
							},
							&ast.ReturnStmt{
								Keyword: &token.Token{Type: token.RETURN},
								Value: &ast.BinaryExpr{
									Left: &ast.CallExpr{
										Callee: &ast.VariableExpr{
											Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "fib"},
										},
										Arguments: []ast.Expr{
											&ast.BinaryExpr{
												Left: &ast.VariableExpr{
													Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "n"},
												},
												Right:    &ast.LiteralExpr{Value: 1},
												Operator: &token.Token{Type: token.MINUS},
											},
										},
									},
									Right: &ast.CallExpr{
										Callee: &ast.VariableExpr{
											Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "fib"},
										},
										Arguments: []ast.Expr{
											&ast.BinaryExpr{
												Left: &ast.VariableExpr{
													Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "n"},
												},
												Right:    &ast.LiteralExpr{Value: 2},
												Operator: &token.Token{Type: token.MINUS},
											},
										},
									},
									Operator: &token.Token{Type: token.PLUS},
								},
							},
						},
					},
				},
				&ast.ExprStmt{
					Expr: &ast.CallExpr{
						Callee: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "fib"},
						},
						Arguments: []ast.Expr{
							&ast.LiteralExpr{Value: 15},
						},
					},
				},
			},
		}, 610}, // fib(15) = 610
	}

	runVmTests(t, tests)
}

func TestNativeFunctions(t *testing.T) {
	tests := []vmTestCase{
		// len("hello") => 5
		{&ast.ExprStmt{
			Expr: &ast.CallExpr{
				Callee: &ast.VariableExpr{
					Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "len"},
				},
				Arguments: []ast.Expr{
					&ast.LiteralExpr{Value: "hello"},
				},
			},
		}, 5},
		// len([1, 2, 3]) => 3
		{&ast.ExprStmt{
			Expr: &ast.CallExpr{
				Callee: &ast.VariableExpr{
					Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "len"},
				},
				Arguments: []ast.Expr{
					&ast.ArrayLiteralExpr{
						Elements: []ast.Expr{
							&ast.LiteralExpr{Value: 1},
							&ast.LiteralExpr{Value: 2},
							&ast.LiteralExpr{Value: 3},
						},
					},
				},
			},
		}, 3},
		// len({}) => 0
		{&ast.ExprStmt{
			Expr: &ast.CallExpr{
				Callee: &ast.VariableExpr{
					Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "len"},
				},
				Arguments: []ast.Expr{
					&ast.HashLiteralExpr{Pairs: []ast.HashPair{}},
				},
			},
		}, 0},
		// len({"a": 1, "b": 2}) => 2
		{&ast.ExprStmt{
			Expr: &ast.CallExpr{
				Callee: &ast.VariableExpr{
					Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "len"},
				},
				Arguments: []ast.Expr{
					&ast.HashLiteralExpr{
						Pairs: []ast.HashPair{
							{Key: &ast.LiteralExpr{Value: "a"}, Value: &ast.LiteralExpr{Value: 1}},
							{Key: &ast.LiteralExpr{Value: "b"}, Value: &ast.LiteralExpr{Value: 2}},
						},
					},
				},
			},
		}, 2},
	}

	runVmTests(t, tests)
}

func TestClockNativeFunction(t *testing.T) {
	// clock() returns a number (unix timestamp)
	comp := compiler.New(nil)
	err := comp.Compile(&ast.ExprStmt{
		Expr: &ast.CallExpr{
			Callee: &ast.VariableExpr{
				Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "clock"},
			},
			Arguments: []ast.Expr{},
		},
	})
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := New(comp.Bytecode())
	err = vm.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}

	result := vm.LastPoppedStackElem()
	if result.Type() != objects.TypeNumber {
		t.Fatalf("expected number, got %s", result.Type())
	}

	num := result.(*objects.Number).Value
	if num <= 0 {
		t.Fatalf("clock() should return a positive number, got %f", num)
	}
}
