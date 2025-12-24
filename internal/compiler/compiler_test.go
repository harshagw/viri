package compiler

import (
	"fmt"
	"testing"

	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/code"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/token"
)

type compilerTestCase struct {
	input                interface{} // Can be ast.Expr or ast.Stmt
	expectedConstants    []interface{}
	expectedInstructions []code.Instructions
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: &ast.BinaryExpr{
				Left:     &ast.LiteralExpr{Value: 1},
				Right:    &ast.LiteralExpr{Value: 2},
				Operator: &token.Token{Type: token.PLUS},
			},
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
			},
		},
		{
			input: &ast.BinaryExpr{
				Left:     &ast.LiteralExpr{Value: 1},
				Right:    &ast.LiteralExpr{Value: 2},
				Operator: &token.Token{Type: token.MINUS},
			},
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSub),
			},
		},
		{
			input: &ast.BinaryExpr{
				Left:     &ast.LiteralExpr{Value: 2},
				Right:    &ast.LiteralExpr{Value: 3},
				Operator: &token.Token{Type: token.STAR},
			},
			expectedConstants: []interface{}{2, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpMul),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             &ast.LiteralExpr{Value: true},
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
			},
		},
		{
			input:             &ast.LiteralExpr{Value: false},
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpFalse),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestUnaryExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: &ast.UnaryExpr{
				Operator: &token.Token{Type: token.MINUS},
				Expr:     &ast.LiteralExpr{Value: 5},
			},
			expectedConstants: []interface{}{5},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpMinus),
			},
		},
		{
			input: &ast.UnaryExpr{
				Operator: &token.Token{Type: token.BANG},
				Expr:     &ast.LiteralExpr{Value: true},
			},
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpBang),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestGroupingExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			// (5)
			input:             &ast.GroupingExpr{Expr: &ast.LiteralExpr{Value: 5}},
			expectedConstants: []interface{}{5},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
			},
		},
		{
			// (1 + 2)
			input: &ast.GroupingExpr{
				Expr: &ast.BinaryExpr{
					Left:     &ast.LiteralExpr{Value: 1},
					Right:    &ast.LiteralExpr{Value: 2},
					Operator: &token.Token{Type: token.PLUS},
				},
			},
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
			},
		},
		{
			// (1 + 2) * 3
			input: &ast.BinaryExpr{
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
			expectedConstants: []interface{}{1, 2, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpMul),
			},
		},
		{
			// Nested grouping: ((5))
			input: &ast.GroupingExpr{
				Expr: &ast.GroupingExpr{
					Expr: &ast.LiteralExpr{Value: 5},
				},
			},
			expectedConstants: []interface{}{5},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestExpressionStatements(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: &ast.ExprStmt{
				Expr: &ast.BinaryExpr{
					Left:     &ast.LiteralExpr{Value: 1},
					Right:    &ast.LiteralExpr{Value: 2},
					Operator: &token.Token{Type: token.PLUS},
				},
			},
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
		{
			input: &ast.ExprStmt{
				Expr: &ast.LiteralExpr{Value: 5},
			},
			expectedConstants: []interface{}{5},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []compilerTestCase{
		{
			// if (true) { 10; }
			input: &ast.IfStmt{
				Condition: &ast.LiteralExpr{Value: true},
				ThenBranch: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.ExprStmt{Expr: &ast.LiteralExpr{Value: 10}},
					},
				},
				ElseBranch: nil,
			},
			expectedConstants: []interface{}{10},
			expectedInstructions: []code.Instructions{
				// 0000
				code.Make(code.OpTrue),
				// 0001
				code.Make(code.OpJumpNotTruthy, 8),
				// 0004
				code.Make(code.OpConstant, 0),
				// 0007
				code.Make(code.OpPop),
				// 0008
			},
		},
		{
			// if (true) { 10; } else { 20; }
			input: &ast.IfStmt{
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
			},
			expectedConstants: []interface{}{10, 20},
			expectedInstructions: []code.Instructions{
				// 0000
				code.Make(code.OpTrue),
				// 0001
				code.Make(code.OpJumpNotTruthy, 11),
				// 0004
				code.Make(code.OpConstant, 0),
				// 0007
				code.Make(code.OpPop),
				// 0008
				code.Make(code.OpJump, 15),
				// 0011
				code.Make(code.OpConstant, 1),
				// 0014
				code.Make(code.OpPop),
				// 0015
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestBlockStatements(t *testing.T) {
	tests := []compilerTestCase{
		{
			// { 1; 2; }
			input: &ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.ExprStmt{Expr: &ast.LiteralExpr{Value: 1}},
					&ast.ExprStmt{Expr: &ast.LiteralExpr{Value: 2}},
				},
			},
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		{
			// { 1 + 2; }
			input: &ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.ExprStmt{
						Expr: &ast.BinaryExpr{
							Left:     &ast.LiteralExpr{Value: 1},
							Right:    &ast.LiteralExpr{Value: 2},
							Operator: &token.Token{Type: token.PLUS},
						},
					},
				},
			},
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func runCompilerTests(t *testing.T, tests []compilerTestCase) {
	t.Helper()

	for _, tt := range tests {
		compiler := New(nil)
		err := compiler.Compile(tt.input)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := compiler.Bytecode()

		err = testInstructions(tt.expectedInstructions, bytecode.Instructions)
		if err != nil {
			t.Fatalf("testInstructions failed: %s", err)
		}

		err = testConstants(t, tt.expectedConstants, bytecode.Constants)
		if err != nil {
			t.Fatalf("testConstants failed: %s", err)
		}
	}
}

func testInstructions(expected []code.Instructions, actual code.Instructions) error {
	concatted := concatInstructions(expected)

	if len(actual) != len(concatted) {
		return fmt.Errorf("wrong instructions length.\nwant=%q\ngot =%q",
			concatted, actual)
	}

	for i, ins := range concatted {
		if actual[i] != ins {
			return fmt.Errorf("wrong instruction at %d.\nwant=%q\ngot =%q",
				i, concatted, actual)
		}
	}

	return nil
}

func concatInstructions(s []code.Instructions) code.Instructions {
	out := code.Instructions{}
	for _, ins := range s {
		out = append(out, ins...)
	}
	return out
}

func testConstants(t *testing.T, expected []interface{}, actual []objects.Object) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("wrong number of constants. want=%d, got=%d",
			len(expected), len(actual))
	}

	for i, constant := range expected {
		switch constant := constant.(type) {
		case int:
			err := testIntegerObject(int64(constant), actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testIntegerObject failed: %s",
					i, err)
			}
		case string:
			err := testStringObject(constant, actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testStringObject failed: %s",
					i, err)
			}
		}
	}

	return nil
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

func TestGlobalVarStatements(t *testing.T) {
	tests := []compilerTestCase{
		{
			// var one = 1;
			input: &ast.VarDeclStmt{
				Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "one"},
				Initializer: &ast.LiteralExpr{Value: 1},
				IsConst:     false,
			},
			expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			// var one = 1; var two = 2;
			input: &ast.BlockStmt{
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
				},
			},
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSetGlobal, 1),
			},
		},
		{
			// var one = 1; one;
			input: &ast.BlockStmt{
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
			},
			expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
		{
			// var one = 1; var two = one; two;
			input: &ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.VarDeclStmt{
						Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "one"},
						Initializer: &ast.LiteralExpr{Value: 1},
						IsConst:     false,
					},
					&ast.VarDeclStmt{
						Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "two"},
						Initializer: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "one"},
						},
						IsConst: false,
					},
					&ast.ExprStmt{
						Expr: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "two"},
						},
					},
				},
			},
			expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestGlobalConstStatements(t *testing.T) {
	tests := []compilerTestCase{
		{
			// const PI = 3;
			input: &ast.VarDeclStmt{
				Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "PI"},
				Initializer: &ast.LiteralExpr{Value: 3},
				IsConst:     true,
			},
			expectedConstants: []interface{}{3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestConstAssignmentError(t *testing.T) {
	// const PI = 3; PI = 4; should fail
	input := &ast.BlockStmt{
		Statements: []ast.Stmt{
			&ast.VarDeclStmt{
				Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "PI"},
				Initializer: &ast.LiteralExpr{Value: 3},
				IsConst:     true,
			},
			&ast.ExprStmt{
				Expr: &ast.AssignExpr{
					Name:  &token.Token{Type: token.IDENTIFIER, Lexeme: "PI"},
					Value: &ast.LiteralExpr{Value: 4},
				},
			},
		},
	}

	compiler := New(nil)
	err := compiler.Compile(input)
	if err == nil {
		t.Fatalf("expected error for const assignment, got none")
	}

	expected := "cannot assign to constant PI"
	if err.Error() != expected {
		t.Fatalf("wrong error. want=%q, got=%q", expected, err.Error())
	}
}

func TestStringExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             &ast.LiteralExpr{Value: "hello"},
			expectedConstants: []interface{}{"hello"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []compilerTestCase{
		{
			// []
			input:             &ast.ArrayLiteralExpr{Elements: []ast.Expr{}},
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpArray, 0),
			},
		},
		{
			// [1, 2, 3]
			input: &ast.ArrayLiteralExpr{
				Elements: []ast.Expr{
					&ast.LiteralExpr{Value: 1},
					&ast.LiteralExpr{Value: 2},
					&ast.LiteralExpr{Value: 3},
				},
			},
			expectedConstants: []interface{}{1, 2, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
			},
		},
		{
			// [1 + 2, 3 - 4, 5 * 6]
			input: &ast.ArrayLiteralExpr{
				Elements: []ast.Expr{
					&ast.BinaryExpr{
						Left:     &ast.LiteralExpr{Value: 1},
						Right:    &ast.LiteralExpr{Value: 2},
						Operator: &token.Token{Type: token.PLUS},
					},
					&ast.BinaryExpr{
						Left:     &ast.LiteralExpr{Value: 3},
						Right:    &ast.LiteralExpr{Value: 4},
						Operator: &token.Token{Type: token.MINUS},
					},
					&ast.BinaryExpr{
						Left:     &ast.LiteralExpr{Value: 5},
						Right:    &ast.LiteralExpr{Value: 6},
						Operator: &token.Token{Type: token.STAR},
					},
				},
			},
			expectedConstants: []interface{}{1, 2, 3, 4, 5, 6},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpSub),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpMul),
				code.Make(code.OpArray, 3),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestHashLiterals(t *testing.T) {
	tests := []compilerTestCase{
		{
			// {}
			input:             &ast.HashLiteralExpr{Pairs: []ast.HashPair{}},
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpHash, 0),
			},
		},
		{
			// {"one": 1, "two": 2}
			input: &ast.HashLiteralExpr{
				Pairs: []ast.HashPair{
					{Key: &ast.LiteralExpr{Value: "one"}, Value: &ast.LiteralExpr{Value: 1}},
					{Key: &ast.LiteralExpr{Value: "two"}, Value: &ast.LiteralExpr{Value: 2}},
				},
			},
			expectedConstants: []interface{}{"one", 1, "two", 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpHash, 4),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestIndexExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			// [1, 2, 3][1]
			input: &ast.IndexExpr{
				Object: &ast.ArrayLiteralExpr{
					Elements: []ast.Expr{
						&ast.LiteralExpr{Value: 1},
						&ast.LiteralExpr{Value: 2},
						&ast.LiteralExpr{Value: 3},
					},
				},
				Index: &ast.LiteralExpr{Value: 1},
			},
			expectedConstants: []interface{}{1, 2, 3, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpIndex),
			},
		},
		{
			// {"one": 1}["one"]
			input: &ast.IndexExpr{
				Object: &ast.HashLiteralExpr{
					Pairs: []ast.HashPair{
						{Key: &ast.LiteralExpr{Value: "one"}, Value: &ast.LiteralExpr{Value: 1}},
					},
				},
				Index: &ast.LiteralExpr{Value: "one"},
			},
			expectedConstants: []interface{}{"one", 1, "one"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpHash, 2),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpIndex),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestSetIndexExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			// [1, 2, 3][0] = 99
			input: &ast.SetIndexExpr{
				Object: &ast.ArrayLiteralExpr{
					Elements: []ast.Expr{
						&ast.LiteralExpr{Value: 1},
						&ast.LiteralExpr{Value: 2},
						&ast.LiteralExpr{Value: 3},
					},
				},
				Index: &ast.LiteralExpr{Value: 0},
				Value: &ast.LiteralExpr{Value: 99},
			},
			expectedConstants: []interface{}{1, 2, 3, 0, 99},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpSetIndex),
			},
		},
		{
			// {"one": 1}["two"] = 2
			input: &ast.SetIndexExpr{
				Object: &ast.HashLiteralExpr{
					Pairs: []ast.HashPair{
						{Key: &ast.LiteralExpr{Value: "one"}, Value: &ast.LiteralExpr{Value: 1}},
					},
				},
				Index: &ast.LiteralExpr{Value: "two"},
				Value: &ast.LiteralExpr{Value: 2},
			},
			expectedConstants: []interface{}{"one", 1, "two", 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpHash, 2),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpSetIndex),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestPrintStatements(t *testing.T) {
	tests := []compilerTestCase{
		{
			// print 1;
			input: &ast.PrintStmt{
				Expr: &ast.LiteralExpr{Value: 1},
			},
			expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPrint),
			},
		},
		{
			// print "hello";
			input: &ast.PrintStmt{
				Expr: &ast.LiteralExpr{Value: "hello"},
			},
			expectedConstants: []interface{}{"hello"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPrint),
			},
		},
		{
			// print "hello" + "world";
			input: &ast.PrintStmt{
				Expr: &ast.BinaryExpr{
					Left:     &ast.LiteralExpr{Value: "hello"},
					Right:    &ast.LiteralExpr{Value: "world"},
					Operator: &token.Token{Type: token.PLUS},
				},
			},
			expectedConstants: []interface{}{"hello", "world"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPrint),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestGlobalAssignmentStatements(t *testing.T) {
	tests := []compilerTestCase{
		{
			// var x = 1; x = 2;
			input: &ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.VarDeclStmt{
						Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "x"},
						Initializer: &ast.LiteralExpr{Value: 1},
						IsConst:     false,
					},
					&ast.ExprStmt{
						Expr: &ast.AssignExpr{
							Name:  &token.Token{Type: token.IDENTIFIER, Lexeme: "x"},
							Value: &ast.LiteralExpr{Value: 2},
						},
					},
				},
			},
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}
