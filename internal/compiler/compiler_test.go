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
				code.Make(code.OpGetConstant, 0),
				code.Make(code.OpGetConstant, 1),
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
				code.Make(code.OpGetConstant, 0),
				code.Make(code.OpGetConstant, 1),
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
				code.Make(code.OpGetConstant, 0),
				code.Make(code.OpGetConstant, 1),
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
				code.Make(code.OpGetConstant, 0),
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
				code.Make(code.OpGetConstant, 0),
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
				code.Make(code.OpGetConstant, 0),
				code.Make(code.OpGetConstant, 1),
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
				code.Make(code.OpGetConstant, 0),
				code.Make(code.OpGetConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpGetConstant, 2),
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
				code.Make(code.OpGetConstant, 0),
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
				code.Make(code.OpGetConstant, 0),
				code.Make(code.OpGetConstant, 1),
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
				code.Make(code.OpGetConstant, 0),
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
				code.Make(code.OpGetConstant, 0),
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
				code.Make(code.OpGetConstant, 0),
				// 0007
				code.Make(code.OpPop),
				// 0008
				code.Make(code.OpJump, 15),
				// 0011
				code.Make(code.OpGetConstant, 1),
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
				code.Make(code.OpGetConstant, 0),
				code.Make(code.OpPop),
				code.Make(code.OpGetConstant, 1),
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
				code.Make(code.OpGetConstant, 0),
				code.Make(code.OpGetConstant, 1),
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
		case []code.Instructions:
			fn, ok := actual[i].(*objects.CompiledFunction)
			if !ok {
				return fmt.Errorf("constant %d - not a function: %T",
					i, actual[i])
			}
			err := testInstructions(constant, fn.Instructions)
			if err != nil {
				return fmt.Errorf("constant %d - testInstructions failed: %s",
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

func TestVarStatements(t *testing.T) {
	tests := []compilerTestCase{
		// Global scope tests
		{
			// var one = 1;
			input: &ast.VarDeclStmt{
				Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "one"},
				Initializer: &ast.LiteralExpr{Value: 1},
				IsConst:     false,
			},
			expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpGetConstant, 0),
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
				code.Make(code.OpGetConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetConstant, 1),
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
				code.Make(code.OpGetConstant, 0),
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
				code.Make(code.OpGetConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpPop),
			},
		},
		// Local scope tests (inside functions)
		{
			// var num = 55; fun() { return num; }  (accessing global from function)
			input: &ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.VarDeclStmt{
						Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "num"},
						Initializer: &ast.LiteralExpr{Value: 55},
						IsConst:     false,
					},
					&ast.ExprStmt{
						Expr: &ast.FunctionExpr{
							Params: []*token.Token{},
							Body: &ast.BlockStmt{
								Statements: []ast.Stmt{
									&ast.ReturnStmt{
										Keyword: &token.Token{Type: token.RETURN},
										Value: &ast.VariableExpr{
											Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "num"},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedConstants: []interface{}{
				55,
				[]code.Instructions{
					code.Make(code.OpGetGlobal, 0),
					code.Make(code.OpReturnValue),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpGetConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetClosure, 1, 0),
				code.Make(code.OpPop),
			},
		},
		{
			// fun() { var num = 55; return num; }  (local variable)
			input: &ast.FunctionExpr{
				Params: []*token.Token{},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.VarDeclStmt{
							Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "num"},
							Initializer: &ast.LiteralExpr{Value: 55},
							IsConst:     false,
						},
						&ast.ReturnStmt{
							Keyword: &token.Token{Type: token.RETURN},
							Value: &ast.VariableExpr{
								Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "num"},
							},
						},
					},
				},
			},
			expectedConstants: []interface{}{
				55,
				[]code.Instructions{
					code.Make(code.OpGetConstant, 0),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpReturnValue),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpGetClosure, 1, 0),
			},
		},
		{
			// fun() { var a = 55; var b = 77; return a + b; }  (multiple locals)
			input: &ast.FunctionExpr{
				Params: []*token.Token{},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.VarDeclStmt{
							Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "a"},
							Initializer: &ast.LiteralExpr{Value: 55},
							IsConst:     false,
						},
						&ast.VarDeclStmt{
							Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "b"},
							Initializer: &ast.LiteralExpr{Value: 77},
							IsConst:     false,
						},
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
			expectedConstants: []interface{}{
				55,
				77,
				[]code.Instructions{
					code.Make(code.OpGetConstant, 0),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpGetConstant, 1),
					code.Make(code.OpSetLocal, 1),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpGetClosure, 2),
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
				code.Make(code.OpGetConstant, 0),
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
				code.Make(code.OpGetConstant, 0),
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
				code.Make(code.OpGetConstant, 0),
				code.Make(code.OpGetConstant, 1),
				code.Make(code.OpGetConstant, 2),
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
				code.Make(code.OpGetConstant, 0),
				code.Make(code.OpGetConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpGetConstant, 2),
				code.Make(code.OpGetConstant, 3),
				code.Make(code.OpSub),
				code.Make(code.OpGetConstant, 4),
				code.Make(code.OpGetConstant, 5),
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
				code.Make(code.OpGetConstant, 0),
				code.Make(code.OpGetConstant, 1),
				code.Make(code.OpGetConstant, 2),
				code.Make(code.OpGetConstant, 3),
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
				code.Make(code.OpGetConstant, 0),
				code.Make(code.OpGetConstant, 1),
				code.Make(code.OpGetConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpGetConstant, 3),
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
				code.Make(code.OpGetConstant, 0),
				code.Make(code.OpGetConstant, 1),
				code.Make(code.OpHash, 2),
				code.Make(code.OpGetConstant, 2),
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
				code.Make(code.OpGetConstant, 0),
				code.Make(code.OpGetConstant, 1),
				code.Make(code.OpGetConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpGetConstant, 3),
				code.Make(code.OpGetConstant, 4),
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
				code.Make(code.OpGetConstant, 0),
				code.Make(code.OpGetConstant, 1),
				code.Make(code.OpHash, 2),
				code.Make(code.OpGetConstant, 2),
				code.Make(code.OpGetConstant, 3),
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
				code.Make(code.OpGetConstant, 0),
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
				code.Make(code.OpGetConstant, 0),
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
				code.Make(code.OpGetConstant, 0),
				code.Make(code.OpGetConstant, 1),
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
				code.Make(code.OpGetConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetConstant, 1),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestLogicalExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			// true and false
			// With OpDup: duplicate left, check truthiness, if truthy pop and eval right
			input: &ast.LogicalExpr{
				Left:     &ast.LiteralExpr{Value: true},
				Operator: &token.Token{Type: token.AND},
				Right:    &ast.LiteralExpr{Value: false},
			},
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				// 0000: OpTrue (left)
				code.Make(code.OpTrue),
				// 0001: OpDup
				code.Make(code.OpDup),
				// 0002: OpJumpNotTruthy -> 7 (end, keep left)
				code.Make(code.OpJumpNotTruthy, 7),
				// 0005: OpPop (discard truthy left)
				code.Make(code.OpPop),
				// 0006: OpFalse (right side)
				code.Make(code.OpFalse),
				// 0007: end
			},
		},
		{
			// false or true
			// With OpDup: duplicate left, check truthiness, if falsy pop and eval right
			input: &ast.LogicalExpr{
				Left:     &ast.LiteralExpr{Value: false},
				Operator: &token.Token{Type: token.OR},
				Right:    &ast.LiteralExpr{Value: true},
			},
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				// 0000: OpFalse (left)
				code.Make(code.OpFalse),
				// 0001: OpDup
				code.Make(code.OpDup),
				// 0002: OpJumpNotTruthy -> 8 (falsy, eval right)
				code.Make(code.OpJumpNotTruthy, 8),
				// 0005: OpJump -> 10 (truthy, skip to end with left)
				code.Make(code.OpJump, 10),
				// 0008: OpPop (discard falsy left)
				code.Make(code.OpPop),
				// 0009: OpTrue (right side)
				code.Make(code.OpTrue),
				// 0010: end
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestWhileStatements(t *testing.T) {
	tests := []compilerTestCase{
		{
			// while (true) { 10; }
			input: &ast.WhileStmt{
				Condition: &ast.LiteralExpr{Value: true},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.ExprStmt{Expr: &ast.LiteralExpr{Value: 10}},
					},
				},
			},
			expectedConstants: []interface{}{10},
			expectedInstructions: []code.Instructions{
				// 0000: OpTrue (condition)
				code.Make(code.OpTrue),
				// 0001: OpJumpNotTruthy -> 11 (exit loop)
				code.Make(code.OpJumpNotTruthy, 11),
				// 0004: OpGetConstant 0 (10)
				code.Make(code.OpGetConstant, 0),
				// 0007: OpPop
				code.Make(code.OpPop),
				// 0008: OpJump -> 0 (back to condition)
				code.Make(code.OpJump, 0),
				// 0011: end
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestForStatements(t *testing.T) {
	tests := []compilerTestCase{
		{
			// Full for loop: for (var i = 0; i < 10; i = i + 1) { 5; }
			input: &ast.ForStmt{
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
						&ast.ExprStmt{Expr: &ast.LiteralExpr{Value: 5}},
					},
				},
			},
			expectedConstants: []interface{}{0, 10, 5, 1},
			expectedInstructions: []code.Instructions{
				// 0000: OpGetConstant 0 (0) - initializer
				code.Make(code.OpGetConstant, 0),
				// 0003: OpSetGlobal 0 (i)
				code.Make(code.OpSetGlobal, 0),
				// 0006: OpGetConstant 1 (10) - condition: i < 10 (compiled as 10 > i)
				code.Make(code.OpGetConstant, 1),
				// 0009: OpGetGlobal 0 (i)
				code.Make(code.OpGetGlobal, 0),
				// 0012: OpGreaterThan
				code.Make(code.OpGreaterThan),
				// 0013: OpJumpNotTruthy -> 37 (exit loop)
				code.Make(code.OpJumpNotTruthy, 37),
				// 0016: OpGetConstant 2 (5) - body
				code.Make(code.OpGetConstant, 2),
				// 0019: OpPop
				code.Make(code.OpPop),
				// 0020: OpGetGlobal 0 (i) - increment: i = i + 1
				code.Make(code.OpGetGlobal, 0),
				// 0023: OpGetConstant 3 (1)
				code.Make(code.OpGetConstant, 3),
				// 0026: OpAdd
				code.Make(code.OpAdd),
				// 0027: OpSetGlobal 0
				code.Make(code.OpSetGlobal, 0),
				// 0030: OpGetGlobal 0 (assignment returns value)
				code.Make(code.OpGetGlobal, 0),
				// 0033: OpPop (discard increment result)
				code.Make(code.OpPop),
				// 0034: OpJump -> 6 (back to condition)
				code.Make(code.OpJump, 6),
				// 0037: end
			},
		},
		{
			// No initializer: for (; i < 10; i = i + 1) { 5; }
			// Assumes i is already defined (index 0)
			input: &ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.VarDeclStmt{
						Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "i"},
						Initializer: &ast.LiteralExpr{Value: 0},
						IsConst:     false,
					},
					&ast.ForStmt{
						Initializer: nil, // No initializer
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
								&ast.ExprStmt{Expr: &ast.LiteralExpr{Value: 5}},
							},
						},
					},
				},
			},
			expectedConstants: []interface{}{0, 10, 5, 1},
			expectedInstructions: []code.Instructions{
				// var i = 0
				code.Make(code.OpGetConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				// for loop starts here (no initializer)
				// 0006: condition
				code.Make(code.OpGetConstant, 1),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpGreaterThan),
				code.Make(code.OpJumpNotTruthy, 37),
				// body
				code.Make(code.OpGetConstant, 2),
				code.Make(code.OpPop),
				// increment
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpGetConstant, 3),
				code.Make(code.OpAdd),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpPop),
				// jump back
				code.Make(code.OpJump, 6),
			},
		},
		{
			// No increment: for (var i = 0; i < 3;) { 5; }
			input: &ast.ForStmt{
				Initializer: &ast.VarDeclStmt{
					Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "i"},
					Initializer: &ast.LiteralExpr{Value: 0},
					IsConst:     false,
				},
				Condition: &ast.BinaryExpr{
					Left:     &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "i"}},
					Right:    &ast.LiteralExpr{Value: 3},
					Operator: &token.Token{Type: token.LESS},
				},
				Increment: nil, // No increment
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.ExprStmt{Expr: &ast.LiteralExpr{Value: 5}},
					},
				},
			},
			expectedConstants: []interface{}{0, 3, 5},
			expectedInstructions: []code.Instructions{
				// 0000: initializer: var i = 0
				code.Make(code.OpGetConstant, 0),
				// 0003:
				code.Make(code.OpSetGlobal, 0),
				// 0006: condition: i < 3 (compiled as 3 > i)
				code.Make(code.OpGetConstant, 1),
				// 0009:
				code.Make(code.OpGetGlobal, 0),
				// 0012:
				code.Make(code.OpGreaterThan),
				// 0013: jump to 23 if false
				code.Make(code.OpJumpNotTruthy, 23),
				// 0016: body: 5;
				code.Make(code.OpGetConstant, 2),
				// 0019:
				code.Make(code.OpPop),
				// 0020: no increment, just jump back to condition
				code.Make(code.OpJump, 6),
				// 0023: end
			},
		},
		{
			// No condition (infinite loop): for (var i = 0;; i = i + 1) { 5; }
			input: &ast.ForStmt{
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
						&ast.ExprStmt{Expr: &ast.LiteralExpr{Value: 5}},
					},
				},
			},
			expectedConstants: []interface{}{0, 5, 1},
			expectedInstructions: []code.Instructions{
				// initializer: var i = 0
				code.Make(code.OpGetConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				// no condition - body starts immediately at position 6
				// body: 5;
				code.Make(code.OpGetConstant, 1),
				code.Make(code.OpPop),
				// increment: i = i + 1
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpGetConstant, 2),
				code.Make(code.OpAdd),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpPop),
				// jump back to body (position 6)
				code.Make(code.OpJump, 6),
			},
		},
		{
			// All parts empty: for (;;) { 5; }
			input: &ast.ForStmt{
				Initializer: nil,
				Condition:   nil,
				Increment:   nil,
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.ExprStmt{Expr: &ast.LiteralExpr{Value: 5}},
					},
				},
			},
			expectedConstants: []interface{}{5},
			expectedInstructions: []code.Instructions{
				// no initializer, no condition - body starts at 0
				// body: 5;
				code.Make(code.OpGetConstant, 0),
				code.Make(code.OpPop),
				// no increment, jump back to start (position 0)
				code.Make(code.OpJump, 0),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestBreakStatement(t *testing.T) {
	tests := []compilerTestCase{
		{
			// while (true) { break; }
			input: &ast.WhileStmt{
				Condition: &ast.LiteralExpr{Value: true},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.BreakStmt{Keyword: &token.Token{Type: token.BREAK}},
					},
				},
			},
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				// 0000: OpTrue (condition)
				code.Make(code.OpTrue),
				// 0001: OpJumpNotTruthy -> 10 (exit loop)
				code.Make(code.OpJumpNotTruthy, 10),
				// 0004: OpJump -> 10 (break)
				code.Make(code.OpJump, 10),
				// 0007: OpJump -> 0 (back to condition)
				code.Make(code.OpJump, 0),
				// 0010: end
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestContinueStatement(t *testing.T) {
	tests := []compilerTestCase{
		{
			// while (true) { continue; 10; }
			input: &ast.WhileStmt{
				Condition: &ast.LiteralExpr{Value: true},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.ContinueStmt{Keyword: &token.Token{Type: token.CONTINUE}},
						&ast.ExprStmt{Expr: &ast.LiteralExpr{Value: 10}},
					},
				},
			},
			expectedConstants: []interface{}{10},
			expectedInstructions: []code.Instructions{
				// 0000: OpTrue (condition)
				code.Make(code.OpTrue),
				// 0001: OpJumpNotTruthy -> 14 (exit loop)
				code.Make(code.OpJumpNotTruthy, 14),
				// 0004: OpJump -> 0 (continue - back to condition)
				code.Make(code.OpJump, 0),
				// 0007: OpGetConstant 0 (10) - unreachable but compiled
				code.Make(code.OpGetConstant, 0),
				// 0010: OpPop
				code.Make(code.OpPop),
				// 0011: OpJump -> 0 (back to condition)
				code.Make(code.OpJump, 0),
				// 0014: end
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestBreakOutsideLoop(t *testing.T) {
	input := &ast.BreakStmt{Keyword: &token.Token{Type: token.BREAK, Lexeme: "break"}}

	compiler := New(nil)
	err := compiler.Compile(input)
	if err == nil {
		t.Fatalf("expected error for break outside loop, got none")
	}

	expected := "break statement outside of loop"
	if err.Error() != expected {
		t.Fatalf("wrong error. want=%q, got=%q", expected, err.Error())
	}
}

func TestContinueOutsideLoop(t *testing.T) {
	input := &ast.ContinueStmt{Keyword: &token.Token{Type: token.CONTINUE, Lexeme: "continue"}}

	compiler := New(nil)
	err := compiler.Compile(input)
	if err == nil {
		t.Fatalf("expected error for continue outside loop, got none")
	}

	expected := "continue statement outside of loop"
	if err.Error() != expected {
		t.Fatalf("wrong error. want=%q, got=%q", expected, err.Error())
	}
}

func TestFunctionExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			// fun() { return 5 + 10; }
			input: &ast.FunctionExpr{
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
			expectedConstants: []interface{}{
				5,
				10,
				[]code.Instructions{
					code.Make(code.OpGetConstant, 0),
					code.Make(code.OpGetConstant, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpGetClosure, 2, 0),
			},
		},
		{
			// fun() { 5 + 10; }  (no return, should return nil)
			input: &ast.FunctionExpr{
				Params: []*token.Token{},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.ExprStmt{
							Expr: &ast.BinaryExpr{
								Left:     &ast.LiteralExpr{Value: 5},
								Right:    &ast.LiteralExpr{Value: 10},
								Operator: &token.Token{Type: token.PLUS},
							},
						},
					},
				},
			},
			expectedConstants: []interface{}{
				5,
				10,
				[]code.Instructions{
					code.Make(code.OpGetConstant, 0),
					code.Make(code.OpGetConstant, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpPop),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpGetClosure, 2, 0),
			},
		},
		{
			// fun() { return; }
			input: &ast.FunctionExpr{
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
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpReturn),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpGetClosure, 0, 0),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestFunctionStatements(t *testing.T) {
	tests := []compilerTestCase{
		{
			// fun add() { return 5 + 10; }
			input: &ast.FunctionStmt{
				Name:   &token.Token{Type: token.IDENTIFIER, Lexeme: "add"},
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
			expectedConstants: []interface{}{
				5,
				10,
				[]code.Instructions{
					code.Make(code.OpGetConstant, 0),
					code.Make(code.OpGetConstant, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpGetClosure, 2, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			// fun identity(a) { return a; }
			input: &ast.FunctionStmt{
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
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpReturnValue),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpGetClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			// fun add(a, b) { return a + b; }
			input: &ast.FunctionStmt{
				Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "add"},
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
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpGetClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			// fun greet() { }  (empty body, returns nil)
			input: &ast.FunctionStmt{
				Name:   &token.Token{Type: token.IDENTIFIER, Lexeme: "greet"},
				Params: []*token.Token{},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{},
				},
			},
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpGetClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestClosures(t *testing.T) {
	tests := []compilerTestCase{
		{
			// fun(a) { return fun(b) { return a + b; }; }
			// Outer function takes 'a', inner function captures 'a' as free variable
			input: &ast.FunctionExpr{
				Params: []*token.Token{
					{Type: token.IDENTIFIER, Lexeme: "a"},
				},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.ReturnStmt{
							Keyword: &token.Token{Type: token.RETURN},
							Value: &ast.FunctionExpr{
								Params: []*token.Token{
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
						},
					},
				},
			},
			expectedConstants: []interface{}{
				// Inner function: uses free variable 'a' and local 'b'
				[]code.Instructions{
					code.Make(code.OpGetFree, 0),  // get 'a' from free variables
					code.Make(code.OpGetLocal, 0), // get 'b' from locals
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
					code.Make(code.OpReturn),
				},
				// Outer function
				[]code.Instructions{
					code.Make(code.OpMakeCell, 0),      // wrap 'a' in Cell for closure capture
					code.Make(code.OpGetClosure, 0, 1), // create closure with 1 free variable
					code.Make(code.OpReturnValue),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpGetClosure, 1, 0), // outer function has no free variables
			},
		},
		{
			// fun(a) { return fun(b) { return fun(c) { return a + b + c; }; }; }
			// Triple nested closure - innermost captures from two levels up
			input: &ast.FunctionExpr{
				Params: []*token.Token{
					{Type: token.IDENTIFIER, Lexeme: "a"},
				},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.ReturnStmt{
							Keyword: &token.Token{Type: token.RETURN},
							Value: &ast.FunctionExpr{
								Params: []*token.Token{
									{Type: token.IDENTIFIER, Lexeme: "b"},
								},
								Body: &ast.BlockStmt{
									Statements: []ast.Stmt{
										&ast.ReturnStmt{
											Keyword: &token.Token{Type: token.RETURN},
											Value: &ast.FunctionExpr{
												Params: []*token.Token{
													{Type: token.IDENTIFIER, Lexeme: "c"},
												},
												Body: &ast.BlockStmt{
													Statements: []ast.Stmt{
														&ast.ReturnStmt{
															Keyword: &token.Token{Type: token.RETURN},
															Value: &ast.BinaryExpr{
																Left: &ast.BinaryExpr{
																	Left: &ast.VariableExpr{
																		Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "a"},
																	},
																	Right: &ast.VariableExpr{
																		Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "b"},
																	},
																	Operator: &token.Token{Type: token.PLUS},
																},
																Right: &ast.VariableExpr{
																	Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "c"},
																},
																Operator: &token.Token{Type: token.PLUS},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedConstants: []interface{}{
				// Innermost function: uses free 'a', free 'b', and local 'c'
				[]code.Instructions{
					code.Make(code.OpGetFree, 0), // get 'a' from free variables
					code.Make(code.OpGetFree, 1), // get 'b' from free variables
					code.Make(code.OpAdd),
					code.Make(code.OpGetLocal, 0), // get 'c' from locals
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
					code.Make(code.OpReturn),
				},
				// Middle function: captures 'a' as free, has local 'b'
				[]code.Instructions{
					code.Make(code.OpGetFree, 0),       // push 'a' (already a Cell) for inner closure
					code.Make(code.OpMakeCell, 0),      // wrap 'b' in Cell for inner closure
					code.Make(code.OpGetClosure, 0, 2), // create inner closure with 2 free vars
					code.Make(code.OpReturnValue),
					code.Make(code.OpReturn),
				},
				// Outer function: has local 'a'
				[]code.Instructions{
					code.Make(code.OpMakeCell, 0),      // wrap 'a' in Cell for middle closure
					code.Make(code.OpGetClosure, 1, 1), // create middle closure with 1 free var
					code.Make(code.OpReturnValue),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpGetClosure, 2, 0), // outer function has no free variables
			},
		},
		{
			// Closure capturing a local variable defined in outer function
			// fun() { var x = 10; return fun() { return x; }; }
			input: &ast.FunctionExpr{
				Params: []*token.Token{},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.VarDeclStmt{
							Name:        &token.Token{Type: token.IDENTIFIER, Lexeme: "x"},
							Initializer: &ast.LiteralExpr{Value: 10},
							IsConst:     false,
						},
						&ast.ReturnStmt{
							Keyword: &token.Token{Type: token.RETURN},
							Value: &ast.FunctionExpr{
								Params: []*token.Token{},
								Body: &ast.BlockStmt{
									Statements: []ast.Stmt{
										&ast.ReturnStmt{
											Keyword: &token.Token{Type: token.RETURN},
											Value: &ast.VariableExpr{
												Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "x"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedConstants: []interface{}{
				10,
				// Inner function: captures 'x' as free variable
				[]code.Instructions{
					code.Make(code.OpGetFree, 0), // get 'x' from free variables
					code.Make(code.OpReturnValue),
					code.Make(code.OpReturn),
				},
				// Outer function
				[]code.Instructions{
					code.Make(code.OpGetConstant, 0),   // push 10
					code.Make(code.OpSetLocal, 0),      // set local 'x'
					code.Make(code.OpMakeCell, 0),      // wrap 'x' in Cell for closure
					code.Make(code.OpGetClosure, 1, 1), // create closure with 1 free variable
					code.Make(code.OpReturnValue),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpGetClosure, 2, 0),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestFunctionCalls(t *testing.T) {
	tests := []compilerTestCase{
		{
			// fun() { return 24; }();
			input: &ast.ExprStmt{
				Expr: &ast.CallExpr{
					Callee: &ast.FunctionExpr{
						Params: []*token.Token{},
						Body: &ast.BlockStmt{
							Statements: []ast.Stmt{
								&ast.ReturnStmt{
									Keyword: &token.Token{Type: token.RETURN},
									Value:   &ast.LiteralExpr{Value: 24},
								},
							},
						},
					},
					Arguments: []ast.Expr{},
				},
			},
			expectedConstants: []interface{}{
				24,
				[]code.Instructions{
					code.Make(code.OpGetConstant, 0),
					code.Make(code.OpReturnValue),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpGetClosure, 1, 0),
				code.Make(code.OpCall, 0),
				code.Make(code.OpPop),
			},
		},
		{
			// var noArg = fun() { return 24; }; noArg();
			input: &ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.VarDeclStmt{
						Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "noArg"},
						Initializer: &ast.FunctionExpr{
							Params: []*token.Token{},
							Body: &ast.BlockStmt{
								Statements: []ast.Stmt{
									&ast.ReturnStmt{
										Keyword: &token.Token{Type: token.RETURN},
										Value:   &ast.LiteralExpr{Value: 24},
									},
								},
							},
						},
						IsConst: false,
					},
					&ast.ExprStmt{
						Expr: &ast.CallExpr{
							Callee: &ast.VariableExpr{
								Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "noArg"},
							},
							Arguments: []ast.Expr{},
						},
					},
				},
			},
			expectedConstants: []interface{}{
				24,
				[]code.Instructions{
					code.Make(code.OpGetConstant, 0),
					code.Make(code.OpReturnValue),
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpGetClosure, 1, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpCall, 0),
				code.Make(code.OpPop),
			},
		},
		{
			// var oneArg = fun(a) { return a; }; oneArg(24);
			input: &ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.VarDeclStmt{
						Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "oneArg"},
						Initializer: &ast.FunctionExpr{
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
						IsConst: false,
					},
					&ast.ExprStmt{
						Expr: &ast.CallExpr{
							Callee: &ast.VariableExpr{
								Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "oneArg"},
							},
							Arguments: []ast.Expr{
								&ast.LiteralExpr{Value: 24},
							},
						},
					},
				},
			},
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpReturnValue),
					code.Make(code.OpReturn),
				},
				24,
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpGetClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpGetConstant, 1),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
			},
		},
		{
			// var manyArg = fun(a, b, c) { return a + b + c; }; manyArg(24, 25, 26);
			input: &ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.VarDeclStmt{
						Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "manyArg"},
						Initializer: &ast.FunctionExpr{
							Params: []*token.Token{
								{Type: token.IDENTIFIER, Lexeme: "a"},
								{Type: token.IDENTIFIER, Lexeme: "b"},
								{Type: token.IDENTIFIER, Lexeme: "c"},
							},
							Body: &ast.BlockStmt{
								Statements: []ast.Stmt{
									&ast.ReturnStmt{
										Keyword: &token.Token{Type: token.RETURN},
										Value: &ast.BinaryExpr{
											Left: &ast.BinaryExpr{
												Left: &ast.VariableExpr{
													Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "a"},
												},
												Right: &ast.VariableExpr{
													Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "b"},
												},
												Operator: &token.Token{Type: token.PLUS},
											},
											Right: &ast.VariableExpr{
												Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "c"},
											},
											Operator: &token.Token{Type: token.PLUS},
										},
									},
								},
							},
						},
						IsConst: false,
					},
					&ast.ExprStmt{
						Expr: &ast.CallExpr{
							Callee: &ast.VariableExpr{
								Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "manyArg"},
							},
							Arguments: []ast.Expr{
								&ast.LiteralExpr{Value: 24},
								&ast.LiteralExpr{Value: 25},
								&ast.LiteralExpr{Value: 26},
							},
						},
					},
				},
			},
			expectedConstants: []interface{}{
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpGetLocal, 2),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
					code.Make(code.OpReturn),
				},
				24,
				25,
				26,
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpGetClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpGetConstant, 1),
				code.Make(code.OpGetConstant, 2),
				code.Make(code.OpGetConstant, 3),
				code.Make(code.OpCall, 3),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestClassDeclaration(t *testing.T) {
	tests := []compilerTestCase{
		{
			// class Animal {}
			input: &ast.ClassStmt{
				Name:    &token.Token{Type: token.IDENTIFIER, Lexeme: "Animal"},
				Methods: []*ast.FunctionStmt{},
			},
			expectedConstants: []interface{}{
				"Animal",
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpNil),          // no superclass
				code.Make(code.OpClass, 0, 0),  // class name at constant 0, 0 methods
				code.Make(code.OpSetGlobal, 0), // store class in global
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestClassWithMethods(t *testing.T) {
	tests := []compilerTestCase{
		{
			// class Animal {
			//   fn speak() { return "sound"; }
			// }
			input: &ast.ClassStmt{
				Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "Animal"},
				Methods: []*ast.FunctionStmt{
					{
						Name:   &token.Token{Type: token.IDENTIFIER, Lexeme: "speak"},
						Params: []*token.Token{},
						Body: &ast.BlockStmt{
							Statements: []ast.Stmt{
								&ast.ReturnStmt{
									Keyword: &token.Token{Type: token.RETURN},
									Value:   &ast.LiteralExpr{Value: "sound"},
								},
							},
						},
					},
				},
			},
			expectedConstants: []interface{}{
				"sound",
				// Method: speak (this is local 0, no explicit params)
				[]code.Instructions{
					code.Make(code.OpGetConstant, 0), // "sound"
					code.Make(code.OpReturnValue),
					code.Make(code.OpReturn),
				},
				"Animal",
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpNil),              // no superclass
				code.Make(code.OpGetClosure, 1, 0), // speak method closure
				code.Make(code.OpClass, 2, 1),      // class name at constant 2, 1 method
				code.Make(code.OpSetGlobal, 0),     // store class in global
			},
		},
		{
			// class Animal {
			//   fn speak() { return "sound"; }
			//   fn eat() { return "eating"; }
			// }
			input: &ast.ClassStmt{
				Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "Animal"},
				Methods: []*ast.FunctionStmt{
					{
						Name:   &token.Token{Type: token.IDENTIFIER, Lexeme: "speak"},
						Params: []*token.Token{},
						Body: &ast.BlockStmt{
							Statements: []ast.Stmt{
								&ast.ReturnStmt{
									Keyword: &token.Token{Type: token.RETURN},
									Value:   &ast.LiteralExpr{Value: "sound"},
								},
							},
						},
					},
					{
						Name:   &token.Token{Type: token.IDENTIFIER, Lexeme: "eat"},
						Params: []*token.Token{},
						Body: &ast.BlockStmt{
							Statements: []ast.Stmt{
								&ast.ReturnStmt{
									Keyword: &token.Token{Type: token.RETURN},
									Value:   &ast.LiteralExpr{Value: "eating"},
								},
							},
						},
					},
				},
			},
			expectedConstants: []interface{}{
				"sound",
				[]code.Instructions{
					code.Make(code.OpGetConstant, 0),
					code.Make(code.OpReturnValue),
					code.Make(code.OpReturn),
				},
				"eating",
				[]code.Instructions{
					code.Make(code.OpGetConstant, 2),
					code.Make(code.OpReturnValue),
					code.Make(code.OpReturn),
				},
				"Animal",
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpNil),              // no superclass
				code.Make(code.OpGetClosure, 1, 0), // speak method closure
				code.Make(code.OpGetClosure, 3, 0), // eat method closure
				code.Make(code.OpClass, 4, 2),      // class name at constant 4, 2 methods
				code.Make(code.OpSetGlobal, 0),     // store class in global
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestClassWithInit(t *testing.T) {
	tests := []compilerTestCase{
		{
			// class Animal {
			//   fn init(name) { this.name = name; }
			// }
			input: &ast.ClassStmt{
				Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "Animal"},
				Methods: []*ast.FunctionStmt{
					{
						Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "init"},
						Params: []*token.Token{
							{Type: token.IDENTIFIER, Lexeme: "name"},
						},
						Body: &ast.BlockStmt{
							Statements: []ast.Stmt{
								&ast.ExprStmt{
									Expr: &ast.SetExpr{
										Object: &ast.ThisExpr{
											Keyword: &token.Token{Type: token.THIS, Lexeme: "this"},
										},
										Name:  &token.Token{Type: token.IDENTIFIER, Lexeme: "name"},
										Value: &ast.VariableExpr{Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "name"}},
									},
								},
							},
						},
					},
				},
			},
			expectedConstants: []interface{}{
				"name", // property name for SetProperty
				// init method: this=local0, name=local1
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),    // this
					code.Make(code.OpGetLocal, 1),    // name parameter
					code.Make(code.OpSetProperty, 0), // set this.name
					code.Make(code.OpPop),            // discard SetExpr result
					code.Make(code.OpGetLocal, 0),    // init returns this
					code.Make(code.OpReturnValue),
				},
				"Animal",
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpNil),              // no superclass
				code.Make(code.OpGetClosure, 1, 0), // init method closure
				code.Make(code.OpClass, 2, 1),      // class name at constant 2, 1 method
				code.Make(code.OpSetGlobal, 0),     // store class in global
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestClassThisCompiler(t *testing.T) {
	tests := []compilerTestCase{
		{
			// class Animal {
			//   fn getName() { return this.name; }
			// }
			input: &ast.ClassStmt{
				Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "Animal"},
				Methods: []*ast.FunctionStmt{
					{
						Name:   &token.Token{Type: token.IDENTIFIER, Lexeme: "getName"},
						Params: []*token.Token{},
						Body: &ast.BlockStmt{
							Statements: []ast.Stmt{
								&ast.ReturnStmt{
									Keyword: &token.Token{Type: token.RETURN},
									Value: &ast.GetExpr{
										Object: &ast.ThisExpr{
											Keyword: &token.Token{Type: token.THIS, Lexeme: "this"},
										},
										Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "name"},
									},
								},
							},
						},
					},
				},
			},
			expectedConstants: []interface{}{
				"name", // property name for GetProperty
				// getName method: this=local0
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),    // this
					code.Make(code.OpGetProperty, 0), // get this.name
					code.Make(code.OpReturnValue),
					code.Make(code.OpReturn),
				},
				"Animal",
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpNil),              // no superclass
				code.Make(code.OpGetClosure, 1, 0), // getName method closure
				code.Make(code.OpClass, 2, 1),      // class name at constant 2, 1 method
				code.Make(code.OpSetGlobal, 0),     // store class in global
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestClassInheritanceCompiler(t *testing.T) {
	tests := []compilerTestCase{
		{
			// class Animal {}
			// class Dog < Animal {}
			input: &ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.ClassStmt{
						Name:    &token.Token{Type: token.IDENTIFIER, Lexeme: "Animal"},
						Methods: []*ast.FunctionStmt{},
					},
					&ast.ClassStmt{
						Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "Dog"},
						SuperClass: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "Animal"},
						},
						Methods: []*ast.FunctionStmt{},
					},
				},
			},
			expectedConstants: []interface{}{
				"Animal",
				"Dog",
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpNil),          // Animal: no superclass
				code.Make(code.OpClass, 0, 0),  // Animal class
				code.Make(code.OpSetGlobal, 0), // store Animal
				code.Make(code.OpGetGlobal, 0), // Dog: get Animal as superclass
				code.Make(code.OpClass, 1, 0),  // Dog class
				code.Make(code.OpSetGlobal, 1), // store Dog
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestClassSuperCompiler(t *testing.T) {
	tests := []compilerTestCase{
		{
			// class Animal {
			//   fn speak() { return "generic"; }
			// }
			// class Dog < Animal {
			//   fn speak() { return super.speak(); }
			// }
			input: &ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.ClassStmt{
						Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "Animal"},
						Methods: []*ast.FunctionStmt{
							{
								Name:   &token.Token{Type: token.IDENTIFIER, Lexeme: "speak"},
								Params: []*token.Token{},
								Body: &ast.BlockStmt{
									Statements: []ast.Stmt{
										&ast.ReturnStmt{
											Keyword: &token.Token{Type: token.RETURN},
											Value:   &ast.LiteralExpr{Value: "generic"},
										},
									},
								},
							},
						},
					},
					&ast.ClassStmt{
						Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "Dog"},
						SuperClass: &ast.VariableExpr{
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "Animal"},
						},
						Methods: []*ast.FunctionStmt{
							{
								Name:   &token.Token{Type: token.IDENTIFIER, Lexeme: "speak"},
								Params: []*token.Token{},
								Body: &ast.BlockStmt{
									Statements: []ast.Stmt{
										&ast.ReturnStmt{
											Keyword: &token.Token{Type: token.RETURN},
											Value: &ast.CallExpr{
												Callee: &ast.SuperExpr{
													Keyword: &token.Token{Type: token.SUPER, Lexeme: "super"},
													Method:  &token.Token{Type: token.IDENTIFIER, Lexeme: "speak"},
												},
												Arguments: []ast.Expr{},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedConstants: []interface{}{
				"generic",
				// Animal.speak method
				[]code.Instructions{
					code.Make(code.OpGetConstant, 0),
					code.Make(code.OpReturnValue),
					code.Make(code.OpReturn),
				},
				"Animal",
				"speak", // method name for OpGetSuper
				// Dog.speak method
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0), // this
					code.Make(code.OpGetSuper, 3), // super.speak (method name at constant 3)
					code.Make(code.OpCall, 0),     // call super.speak()
					code.Make(code.OpReturnValue),
					code.Make(code.OpReturn),
				},
				"Dog",
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpNil),              // Animal: no superclass
				code.Make(code.OpGetClosure, 1, 0), // Animal.speak method
				code.Make(code.OpClass, 2, 1),      // Animal class
				code.Make(code.OpSetGlobal, 0),     // store Animal
				code.Make(code.OpGetGlobal, 0),     // Dog: get Animal as superclass
				code.Make(code.OpGetClosure, 4, 0), // Dog.speak method
				code.Make(code.OpClass, 5, 1),      // Dog class
				code.Make(code.OpSetGlobal, 1),     // store Dog
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestPropertyAccessCompiler(t *testing.T) {
	tests := []compilerTestCase{
		{
			// class Animal {}
			// var a = Animal();
			// a.name;
			input: &ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.ClassStmt{
						Name:    &token.Token{Type: token.IDENTIFIER, Lexeme: "Animal"},
						Methods: []*ast.FunctionStmt{},
					},
					&ast.VarDeclStmt{
						Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "a"},
						Initializer: &ast.CallExpr{
							Callee: &ast.VariableExpr{
								Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "Animal"},
							},
							Arguments: []ast.Expr{},
						},
					},
					&ast.ExprStmt{
						Expr: &ast.GetExpr{
							Object: &ast.VariableExpr{
								Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "a"},
							},
							Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "name"},
						},
					},
				},
			},
			expectedConstants: []interface{}{
				"Animal",
				"name",
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpNil),            // no superclass
				code.Make(code.OpClass, 0, 0),    // Animal class
				code.Make(code.OpSetGlobal, 0),   // store Animal
				code.Make(code.OpGetGlobal, 0),   // get Animal
				code.Make(code.OpCall, 0),        // Animal()
				code.Make(code.OpSetGlobal, 1),   // store a
				code.Make(code.OpGetGlobal, 1),   // get a
				code.Make(code.OpGetProperty, 1), // a.name
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestPropertySetCompiler(t *testing.T) {
	tests := []compilerTestCase{
		{
			// class Animal {}
			// var a = Animal();
			// a.name = "Dog";
			input: &ast.BlockStmt{
				Statements: []ast.Stmt{
					&ast.ClassStmt{
						Name:    &token.Token{Type: token.IDENTIFIER, Lexeme: "Animal"},
						Methods: []*ast.FunctionStmt{},
					},
					&ast.VarDeclStmt{
						Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "a"},
						Initializer: &ast.CallExpr{
							Callee: &ast.VariableExpr{
								Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "Animal"},
							},
							Arguments: []ast.Expr{},
						},
					},
					&ast.ExprStmt{
						Expr: &ast.SetExpr{
							Object: &ast.VariableExpr{
								Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "a"},
							},
							Name:  &token.Token{Type: token.IDENTIFIER, Lexeme: "name"},
							Value: &ast.LiteralExpr{Value: "Dog"},
						},
					},
				},
			},
			expectedConstants: []interface{}{
				"Animal",
				"Dog",
				"name",
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpNil),            // no superclass
				code.Make(code.OpClass, 0, 0),    // Animal class
				code.Make(code.OpSetGlobal, 0),   // store Animal
				code.Make(code.OpGetGlobal, 0),   // get Animal
				code.Make(code.OpCall, 0),        // Animal()
				code.Make(code.OpSetGlobal, 1),   // store a
				code.Make(code.OpGetGlobal, 1),   // get a
				code.Make(code.OpGetConstant, 1), // "Dog"
				code.Make(code.OpSetProperty, 2), // a.name = "Dog"
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestThisOutsideClass(t *testing.T) {
	// 'this' outside of a class should error
	input := &ast.ExprStmt{
		Expr: &ast.ThisExpr{
			Keyword: &token.Token{Type: token.THIS, Lexeme: "this"},
		},
	}

	comp := New(nil)
	err := comp.Compile(input)
	if err == nil {
		t.Fatalf("expected error for 'this' outside class, got none")
	}

	expected := "cannot use 'this' outside of a class"
	if err.Error() != expected {
		t.Fatalf("wrong error. want=%q, got=%q", expected, err.Error())
	}
}

func TestSuperOutsideClass(t *testing.T) {
	// 'super' outside of a class should error
	input := &ast.ExprStmt{
		Expr: &ast.SuperExpr{
			Keyword: &token.Token{Type: token.SUPER, Lexeme: "super"},
			Method:  &token.Token{Type: token.IDENTIFIER, Lexeme: "speak"},
		},
	}

	comp := New(nil)
	err := comp.Compile(input)
	if err == nil {
		t.Fatalf("expected error for 'super' outside class, got none")
	}

	expected := "cannot use 'super' outside of a class"
	if err.Error() != expected {
		t.Fatalf("wrong error. want=%q, got=%q", expected, err.Error())
	}
}

func TestSuperWithoutSuperclass(t *testing.T) {
	// 'super' in a class with no superclass should error
	input := &ast.ClassStmt{
		Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "Animal"},
		Methods: []*ast.FunctionStmt{
			{
				Name:   &token.Token{Type: token.IDENTIFIER, Lexeme: "speak"},
				Params: []*token.Token{},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.ReturnStmt{
							Keyword: &token.Token{Type: token.RETURN},
							Value: &ast.CallExpr{
								Callee: &ast.SuperExpr{
									Keyword: &token.Token{Type: token.SUPER, Lexeme: "super"},
									Method:  &token.Token{Type: token.IDENTIFIER, Lexeme: "speak"},
								},
								Arguments: []ast.Expr{},
							},
						},
					},
				},
			},
		},
	}

	comp := New(nil)
	err := comp.Compile(input)
	if err == nil {
		t.Fatalf("expected error for 'super' without superclass, got none")
	}

	expected := "cannot use 'super' in a class with no superclass"
	if err.Error() != expected {
		t.Fatalf("wrong error. want=%q, got=%q", expected, err.Error())
	}
}

func TestClassSelfInheritance(t *testing.T) {
	// class Animal < Animal {} should error
	input := &ast.ClassStmt{
		Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "Animal"},
		SuperClass: &ast.VariableExpr{
			Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "Animal"},
		},
		Methods: []*ast.FunctionStmt{},
	}

	comp := New(nil)
	err := comp.Compile(input)
	if err == nil {
		t.Fatalf("expected error for self-inheritance, got none")
	}

	expected := "a class cannot inherit from itself"
	if err.Error() != expected {
		t.Fatalf("wrong error. want=%q, got=%q", expected, err.Error())
	}
}

func TestClassMethodWithParameters(t *testing.T) {
	tests := []compilerTestCase{
		{
			// class Calculator {
			//   fn add(a, b) { return a + b; }
			// }
			input: &ast.ClassStmt{
				Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "Calculator"},
				Methods: []*ast.FunctionStmt{
					{
						Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "add"},
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
				},
			},
			expectedConstants: []interface{}{
				// add method: this=local0, a=local1, b=local2
				[]code.Instructions{
					code.Make(code.OpGetLocal, 1), // a
					code.Make(code.OpGetLocal, 2), // b
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
					code.Make(code.OpReturn),
				},
				"Calculator",
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpNil),              // no superclass
				code.Make(code.OpGetClosure, 0, 0), // add method closure
				code.Make(code.OpClass, 1, 1),      // class name at constant 1, 1 method
				code.Make(code.OpSetGlobal, 0),     // store class in global
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestClassInitReturnsThis(t *testing.T) {
	tests := []compilerTestCase{
		{
			// class Animal {
			//   fn init() { return; }  // explicit return in init should still return this
			// }
			input: &ast.ClassStmt{
				Name: &token.Token{Type: token.IDENTIFIER, Lexeme: "Animal"},
				Methods: []*ast.FunctionStmt{
					{
						Name:   &token.Token{Type: token.IDENTIFIER, Lexeme: "init"},
						Params: []*token.Token{},
						Body: &ast.BlockStmt{
							Statements: []ast.Stmt{
								&ast.ReturnStmt{
									Keyword: &token.Token{Type: token.RETURN},
									Value:   nil, // explicit return with no value
								},
							},
						},
					},
				},
			},
			expectedConstants: []interface{}{
				// init method: this=local0
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0), // return this (even with explicit return;)
					code.Make(code.OpReturnValue),
					code.Make(code.OpGetLocal, 0), // implicit this return at end
					code.Make(code.OpReturnValue),
				},
				"Animal",
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpNil),              // no superclass
				code.Make(code.OpGetClosure, 0, 0), // init method closure
				code.Make(code.OpClass, 1, 1),      // class name at constant 1, 1 method
				code.Make(code.OpSetGlobal, 0),     // store class in global
			},
		},
	}

	runCompilerTests(t, tests)
}
