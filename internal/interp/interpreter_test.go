package interp

import (
	"testing"

	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/token"
)

func TestInterpreter_EvalLiteral(t *testing.T) {
	i := NewInterpreter(nil)
	
	tests := []struct {
		name     string
		expr     ast.Expr
		expected interface{}
	}{
		{
			name:     "number literal",
			expr:     &ast.LiteralExpr{Value: 42.0},
			expected: 42.0,
		},
		{
			name:     "string literal",
			expr:     &ast.LiteralExpr{Value: "hello"},
			expected: "hello",
		},
		{
			name:     "bool literal",
			expr:     &ast.LiteralExpr{Value: true},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := i.evalExpr(tt.expr)
			if err != nil {
				t.Fatalf("evalExpr() error = %v", err)
			}
			
			var actual interface{}
			switch v := result.(type) {
			case *objects.Number:
				actual = v.Value
			case *objects.String:
				actual = v.Value
			case *objects.Bool:
				actual = v.Value
			}

			if actual != tt.expected {
				t.Errorf("got %v, want %v", actual, tt.expected)
			}
		})
	}
}

func TestInterpreter_EvalBinary(t *testing.T) {
	i := NewInterpreter(nil)

	plusTok := token.New(token.PLUS, "+", nil, 1, nil)
	
	expr := &ast.BinaryExpr{
		Left:     &ast.LiteralExpr{Value: 10.0},
		Operator: &plusTok,
		Right:    &ast.LiteralExpr{Value: 20.0},
	}

	result, err := i.evalExpr(expr)
	if err != nil {
		t.Fatalf("evalExpr() error = %v", err)
	}

	num, ok := result.(*objects.Number)
	if !ok {
		t.Fatalf("expected Number, got %T", result)
	}

	if num.Value != 30.0 {
		t.Errorf("got %v, want 30.0", num.Value)
	}
}

func TestInterpreter_EvalVarDecl(t *testing.T) {
	globals := objects.NewEnvironment(nil)
	i := NewInterpreter(globals)

	nameTok := token.New(token.IDENTIFIER, "x", nil, 1, nil)
	
	stmt := &ast.VarDeclStmt{
		Name: &nameTok,
		Initializer: &ast.LiteralExpr{Value: 100.0},
	}

	_, err := i.evalStmt(stmt)
	if err != nil {
		t.Fatalf("evalStmt() error = %v", err)
	}

	val, err := globals.Get("x")
	if err != nil {
		t.Fatalf("globals.Get(x) error = %v", err)
	}

	num := val.(*objects.Number)
	if num.Value != 100.0 {
		t.Errorf("got %v, want 100.0", num.Value)
	}
}

func TestInterpreter_EvalIf(t *testing.T) {
	globals := objects.NewEnvironment(nil)
	i := NewInterpreter(globals)

	// x = 0; if (true) x = 1;
	xTok := token.New(token.IDENTIFIER, "x", nil, 1, nil)
	globals.Define("x", objects.NewNumber(0))

	stmt := &ast.IfStmt{
		Condition: &ast.LiteralExpr{Value: true},
		ThenBranch: &ast.ExprStmt{
			Expr: &ast.AssignExpr{
				Name:  &xTok,
				Value: &ast.LiteralExpr{Value: 1.0},
			},
		},
	}

	_, err := i.evalStmt(stmt)
	if err != nil {
		t.Fatalf("evalStmt() error = %v", err)
	}

	val, _ := globals.Get("x")
	if val.(*objects.Number).Value != 1.0 {
		t.Errorf("got %v, want 1.0", val.(*objects.Number).Value)
	}
}

func TestInterpreter_EvalFunction(t *testing.T) {
	globals := objects.NewEnvironment(nil)
	i := NewInterpreter(globals)

	// fun add(a, b) { return a + b; }
	// add(1, 2)
	addTok := token.New(token.IDENTIFIER, "add", nil, 1, nil)
	aTok := token.New(token.IDENTIFIER, "a", nil, 1, nil)
	bTok := token.New(token.IDENTIFIER, "b", nil, 1, nil)
	plusTok := token.New(token.PLUS, "+", nil, 1, nil)

	aVarExpr := &ast.VariableExpr{Name: &aTok}
	bVarExpr := &ast.VariableExpr{Name: &bTok}

	funStmt := &ast.FunctionStmt{
		Name: &addTok,
		Params: []*token.Token{&aTok, &bTok},
		Body: &ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.ReturnStmt{
					Value: &ast.BinaryExpr{
						Left:     aVarExpr,
						Operator: &plusTok,
						Right:    bVarExpr,
					},
				},
			},
		},
	}

	// Manually set locals to avoid dependency on resolver
	i.SetLocals(map[ast.Expr]int{
		aVarExpr: 0,
		bVarExpr: 0,
	})

	_, err := i.evalStmt(funStmt)
	if err != nil {
		t.Fatalf("evalStmt(funStmt) error = %v", err)
	}

	callExpr := &ast.CallExpr{
		Callee: &ast.VariableExpr{Name: &addTok},
		Arguments: []ast.Expr{
			&ast.LiteralExpr{Value: 1.0},
			&ast.LiteralExpr{Value: 2.0},
		},
	}
	result, err := i.evalExpr(callExpr)
	if err != nil {
		t.Fatalf("evalExpr(callExpr) error = %v", err)
	}

	num := result.(*objects.Number)
	if num.Value != 3.0 {
		t.Errorf("got %v, want 3.0", num.Value)
	}
}

func TestInterpreter_EvalClass(t *testing.T) {
	globals := objects.NewEnvironment(nil)
	i := NewInterpreter(globals)

	// class Point { init(x, y) { this.x = x; this.y = y; } }
	// var p = Point(1, 2); p.x
	pointTok := token.New(token.IDENTIFIER, "Point", nil, 1, nil)
	initTok := token.New(token.IDENTIFIER, "init", nil, 1, nil)
	xTok := token.New(token.IDENTIFIER, "x", nil, 1, nil)
	yTok := token.New(token.IDENTIFIER, "y", nil, 1, nil)
	thisTok := token.New(token.THIS, "this", nil, 1, nil)

	thisX := &ast.ThisExpr{Keyword: &thisTok}
	valX := &ast.VariableExpr{Name: &xTok}
	thisY := &ast.ThisExpr{Keyword: &thisTok}
	valY := &ast.VariableExpr{Name: &yTok}

	classStmt := &ast.ClassStmt{
		Name: &pointTok,
		Methods: []*ast.FunctionStmt{
			{
				Name: &initTok,
				Params: []*token.Token{&xTok, &yTok},
				Body: &ast.BlockStmt{
					Statements: []ast.Stmt{
						&ast.ExprStmt{
							Expr: &ast.SetExpr{
								Object: thisX,
								Name:   &xTok,
								Value:  valX,
							},
						},
						&ast.ExprStmt{
							Expr: &ast.SetExpr{
								Object: thisY,
								Name:   &yTok,
								Value:  valY,
							},
						},
					},
				},
			},
		},
	}

	// Manually set locals for 'this' (depth 1) and parameters (depth 0)
	i.SetLocals(map[ast.Expr]int{
		thisX: 1,
		valX:  0,
		thisY: 1,
		valY:  0,
	})

	_, err := i.evalStmt(classStmt)
	if err != nil {
		t.Fatalf("evalStmt(classStmt) error = %v", err)
	}

	pNameTok := token.New(token.IDENTIFIER, "p", nil, 0, nil)
	decl := &ast.VarDeclStmt{
		Name: &pNameTok,
		Initializer: &ast.CallExpr{
			Callee: &ast.VariableExpr{Name: &pointTok},
			Arguments: []ast.Expr{
				&ast.LiteralExpr{Value: 10.0},
				&ast.LiteralExpr{Value: 20.0},
			},
		},
	}
	_, err = i.evalStmt(decl)
	if err != nil {
		t.Fatalf("evalStmt(decl) error = %v", err)
	}

	getExpr := &ast.GetExpr{
		Object: &ast.VariableExpr{Name: &pNameTok},
		Name:   &xTok,
	}
	result, err := i.evalExpr(getExpr)
	if err != nil {
		t.Fatalf("evalExpr(getExpr) error = %v", err)
	}

	num := result.(*objects.Number)
	if num.Value != 10.0 {
		t.Errorf("got %v, want 10.0", num.Value)
	}
}

func TestInterpreter_EvalCollections(t *testing.T) {
	i := NewInterpreter(nil)

	// [1, 2, 3]
	arrExpr := &ast.ArrayLiteralExpr{
		Elements: []ast.Expr{
			&ast.LiteralExpr{Value: 1.0},
			&ast.LiteralExpr{Value: 2.0},
			&ast.LiteralExpr{Value: 3.0},
		},
	}
	result, _ := i.evalExpr(arrExpr)
	arr := result.(*objects.Array)
	if len(arr.Elements) != 3 {
		t.Errorf("got length %d, want 3", len(arr.Elements))
	}

	// {"a": 1}
	hashExpr := &ast.HashLiteralExpr{
		Pairs: []ast.HashPair{
			{
				Key:   &ast.LiteralExpr{Value: "a"},
				Value: &ast.LiteralExpr{Value: 1.0},
			},
		},
	}
	result, _ = i.evalExpr(hashExpr)
	hash := result.(*objects.Hash)
	if len(hash.Pairs) != 1 {
		t.Errorf("got %d pairs, want 1", len(hash.Pairs))
	}
}

func TestInterpreter_EvalIndexing(t *testing.T) {
	i := NewInterpreter(nil)

	// var a = [10, 20]; a[0] = 30; a[0]
	aTok := token.New(token.IDENTIFIER, "a", nil, 1, nil)
	bracketTok := token.New(token.RIGHT_BRACKET, "]", nil, 1, nil)

	i.globals.Define("a", objects.NewArray([]objects.Object{objects.NewNumber(10), objects.NewNumber(20)}))

	// a[0] = 30
	setIndex := &ast.SetIndexExpr{
		Object:  &ast.VariableExpr{Name: &aTok},
		Index:   &ast.LiteralExpr{Value: 0.0},
		Value:   &ast.LiteralExpr{Value: 30.0},
		Bracket: &bracketTok,
	}
	i.evalExpr(setIndex)

	// a[0]
	getIndex := &ast.IndexExpr{
		Object:  &ast.VariableExpr{Name: &aTok},
		Index:   &ast.LiteralExpr{Value: 0.0},
		Bracket: &bracketTok,
	}
	result, _ := i.evalExpr(getIndex)
	num := result.(*objects.Number)
	if num.Value != 30.0 {
		t.Errorf("got %v, want 30.0", num.Value)
	}
}

func TestInterpreter_EvalLoopControl(t *testing.T) {
	globals := objects.NewEnvironment(nil)
	i := NewInterpreter(globals)

	// var x = 0; while (true) { x = x + 1; break; }
	xTok := token.New(token.IDENTIFIER, "x", nil, 1, nil)
	globals.Define("x", objects.NewNumber(0))
	plusTok := token.New(token.PLUS, "+", nil, 1, nil)

	stmt := &ast.WhileStmt{
		Condition: &ast.LiteralExpr{Value: true},
		Body: &ast.BlockStmt{
			Statements: []ast.Stmt{
				&ast.ExprStmt{
					Expr: &ast.AssignExpr{
						Name: &xTok,
						Value: &ast.BinaryExpr{
							Left:     &ast.VariableExpr{Name: &xTok},
							Operator: &plusTok,
							Right:    &ast.LiteralExpr{Value: 1.0},
						},
					},
				},
				&ast.BreakStmt{},
			},
		},
	}

	_, err := i.evalStmt(stmt)
	if err != nil && err.Error() != "break" { // In reality loops catch these specifically
		t.Fatalf("evalStmt() unexpected error = %v", err)
	}

	val, _ := globals.Get("x")
	if val.(*objects.Number).Value != 1.0 {
		t.Errorf("got %v, want 1.0", val.(*objects.Number).Value)
	}
}

func TestInterpreter_EvalUnary(t *testing.T) {
	i := NewInterpreter(nil)

	minusTok := token.New(token.MINUS, "-", nil, 1, nil)
	bangTok := token.New(token.BANG, "!", nil, 1, nil)

	tests := []struct {
		name     string
		expr     ast.Expr
		expected interface{}
	}{
		{
			name: "-42",
			expr: &ast.UnaryExpr{
				Operator: &minusTok,
				Expr:     &ast.LiteralExpr{Value: 42.0},
			},
			expected: -42.0,
		},
		{
			name: "--42",
			expr: &ast.UnaryExpr{
				Operator: &minusTok,
				Expr:     &ast.UnaryExpr{
					Operator: &minusTok,
					Expr:     &ast.LiteralExpr{Value: 42.0},
				},
			},
			expected: 42.0,
		},
		{
			name: "!true",
			expr: &ast.UnaryExpr{
				Operator: &bangTok,
				Expr:     &ast.LiteralExpr{Value: true},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := i.evalExpr(tt.expr)
			if err != nil {
				t.Fatalf("evalExpr() error = %v", err)
			}
			var actual interface{}
			switch v := result.(type) {
			case *objects.Number:
				actual = v.Value
			case *objects.Bool:
				actual = v.Value
			}
			if actual != tt.expected {
				t.Errorf("got %v, want %v", actual, tt.expected)
			}
		})
	}
}

func TestInterpreter_EvalLogical(t *testing.T) {
	i := NewInterpreter(nil)

	andTok := token.New(token.AND, "and", nil, 1, nil)
	orTok := token.New(token.OR, "or", nil, 1, nil)

	tests := []struct {
		name     string
		expr     ast.Expr
		expected bool
	}{
		{
			name: "true and false",
			expr: &ast.LogicalExpr{
				Left:     &ast.LiteralExpr{Value: true},
				Operator: &andTok,
				Right:    &ast.LiteralExpr{Value: false},
			},
			expected: false,
		},
		{
			name: "true and true",
			expr: &ast.LogicalExpr{
				Left:     &ast.LiteralExpr{Value: true},
				Operator: &andTok,
				Right:    &ast.LiteralExpr{Value: true},
			},
			expected: true,
		},
		{
			name: "false and false",
			expr: &ast.LogicalExpr{
				Left:     &ast.LiteralExpr{Value: false},
				Operator: &andTok,
				Right:    &ast.LiteralExpr{Value: false},
			},
			expected: false,
		},
		{
			name: "true or false",
			expr: &ast.LogicalExpr{
				Left:     &ast.LiteralExpr{Value: true},
				Operator: &orTok,
				Right:    &ast.LiteralExpr{Value: false},
			},
			expected: true,
		},
		{
			name: "true or true",
			expr: &ast.LogicalExpr{
				Left:     &ast.LiteralExpr{Value: true},
				Operator: &orTok,
				Right:    &ast.LiteralExpr{Value: true},
			},
			expected: true,
		},
		{
			name: "false or false",
			expr: &ast.LogicalExpr{
				Left:     &ast.LiteralExpr{Value: false},
				Operator: &orTok,
				Right:    &ast.LiteralExpr{Value: false},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := i.evalExpr(tt.expr)
			if err != nil {
				t.Fatalf("evalExpr() error = %v", err)
			}
			if result.(*objects.Bool).Value != tt.expected {
				t.Errorf("got %v, want %v", result.(*objects.Bool).Value, tt.expected)
			}
		})
	}
}

func TestInterpreter_EvalWhile(t *testing.T) {
	globals := objects.NewEnvironment(nil)
	i := NewInterpreter(globals)

	// var x = 0; while (x < 3) { x = x + 1; }
	xTok := token.New(token.IDENTIFIER, "x", nil, 1, nil)
	globals.Define("x", objects.NewNumber(0))
	
	lessTok := token.New(token.LESS, "<", nil, 1, nil)
	plusTok := token.New(token.PLUS, "+", nil, 1, nil)

	stmt := &ast.WhileStmt{
		Condition: &ast.BinaryExpr{
			Left:     &ast.VariableExpr{Name: &xTok},
			Operator: &lessTok,
			Right:    &ast.LiteralExpr{Value: 3.0},
		},
		Body: &ast.ExprStmt{
			Expr: &ast.AssignExpr{
				Name: &xTok,
				Value: &ast.BinaryExpr{
					Left:     &ast.VariableExpr{Name: &xTok},
					Operator: &plusTok,
					Right:    &ast.LiteralExpr{Value: 1.0},
				},
			},
		},
	}

	_, err := i.evalStmt(stmt)
	if err != nil {
		t.Fatalf("evalStmt() error = %v", err)
	}

	val, _ := globals.Get("x")
	if val.(*objects.Number).Value != 3.0 {
		t.Errorf("got %v, want 3.0", val.(*objects.Number).Value)
	}
}
