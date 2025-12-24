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
		}
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
