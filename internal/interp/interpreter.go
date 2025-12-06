package interp

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/token"
)

type Interpreter struct {
	environment *objects.Environment
	globals     *objects.Environment
	locals      map[ast.Expr]int
}

func NewInterpreter(globals *objects.Environment) *Interpreter {
	if globals == nil {
		globals = objects.NewEnvironment(nil)
	}
	globals.Define("clock", objects.NewClock())
	globals.Define("len", objects.NewLen())
	return &Interpreter{
		environment: globals,
		globals:     globals,
		locals:      make(map[ast.Expr]int),
	}
}

func (i *Interpreter) SetLocals(locals map[ast.Expr]int) {
	i.locals = locals
}

func (i *Interpreter) Interpret(stmts []ast.Stmt) error {
	for _, stmt := range stmts {
		if _, err := i.evalStmt(stmt); err != nil {
			return err
		}
	}
	return nil
}

// ExecuteBlock executes a block in a provided environment.
func (i *Interpreter) ExecuteBlock(block *ast.BlockStmt, env *objects.Environment) (objects.Object, error) {
	previous := i.environment
	i.environment = env
	defer func() { i.environment = previous }()

	for _, stmt := range block.Statements {
		if _, err := i.evalStmt(stmt); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

// Statements

func (i *Interpreter) evalStmt(stmt ast.Stmt) (objects.Object, error) {
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		newEnv := objects.NewEnvironment(i.environment)
		return i.ExecuteBlock(s, newEnv)
	case *ast.FunctionStmt:
		return i.visitFunction(s)
	case *ast.ClassStmt:
		return i.visitClass(s)
	case *ast.ReturnStmt:
		return i.visitReturnStmt(s)
	case *ast.VarDeclStmt:
		return i.visitVarDeclStmt(s)
	case *ast.PrintStmt:
		return i.visitPrintStmt(s)
	case *ast.ExprStmt:
		return i.visitExprStmt(s)
	case *ast.IfStmt:
		return i.visitIfStmt(s)
	case *ast.WhileStmt:
		return i.visitWhileStmt(s)
	case *ast.BreakStmt:
		return nil, &objects.BreakError{}
	default:
		return nil, errors.New("invalid statement")
	}
}

func (i *Interpreter) visitFunction(function *ast.FunctionStmt) (objects.Object, error) {
	fn := objects.NewFunction(function, i.environment, false)
	i.environment.Define(function.Name.Lexeme, fn)
	return nil, nil
}

func (i *Interpreter) visitClass(class *ast.ClassStmt) (objects.Object, error) {
	i.environment.Define(class.Name.Lexeme, objects.NewNil())

	var superclass objects.Object
	var err error
	methodEnvironment := i.environment

	if class.SuperClass != nil {
		superclass, err = i.evalExpr(class.SuperClass)
		if err != nil {
			return nil, err
		}
		if superclass.Type() != objects.TypeClass {
			return nil, i.runtimeError(class.SuperClass.Name, "Superclass must be a class.")
		}
	}

	var superClassObj *objects.Class
	if superclass != nil {
		superClassObj = superclass.(*objects.Class)
		methodEnvironment = objects.NewEnvironment(i.environment)
		methodEnvironment.Define("super", superClassObj)
	}

	methods := make(map[string]*objects.Function, len(class.Methods))
	for _, method := range class.Methods {
		function := objects.NewFunction(method, methodEnvironment, method.Name.Lexeme == "init")
		methods[method.Name.Lexeme] = function
	}

	classObj := objects.NewClass(class.Name.Lexeme, superClassObj, methods)

	if err := methodEnvironment.Assign(class.Name.Lexeme, classObj); err != nil {
		return nil, i.runtimeError(class.Name, err.Error())
	}
	return nil, nil
}

func (i *Interpreter) visitReturnStmt(ret *ast.ReturnStmt) (objects.Object, error) {
	var value objects.Object = objects.NewNil()
	if ret.Value != nil {
		v, err := i.evalExpr(ret.Value)
		if err != nil {
			return nil, err
		}
		value = v
	}
	return value, &objects.ReturnError{Value: value}
}

func (i *Interpreter) visitVarDeclStmt(decl *ast.VarDeclStmt) (objects.Object, error) {
	if decl.Initializer != nil {
		val, err := i.evalExpr(decl.Initializer)
		if err != nil {
			return nil, err
		}
		i.environment.Define(decl.Name.Lexeme, val)
	} else {
		i.environment.Define(decl.Name.Lexeme, objects.NewNil())
	}
	return nil, nil
}

func (i *Interpreter) visitPrintStmt(printStmt *ast.PrintStmt) (objects.Object, error) {
	value, err := i.evalExpr(printStmt.Expr)
	if err != nil {
		return nil, err
	}
	fmt.Println(objects.Stringify(value))
	return value, nil
}

func (i *Interpreter) visitExprStmt(exprStmt *ast.ExprStmt) (objects.Object, error) {
	value, err := i.evalExpr(exprStmt.Expr)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (i *Interpreter) visitIfStmt(ifStmt *ast.IfStmt) (objects.Object, error) {
	cond, err := i.evalExpr(ifStmt.Condition)
	if err != nil {
		return nil, err
	}
	if objects.IsTruthy(cond) {
		return i.evalStmt(ifStmt.ThenBranch)
	} else if ifStmt.ElseBranch != nil {
		return i.evalStmt(ifStmt.ElseBranch)
	}
	return nil, nil
}

func (i *Interpreter) visitWhileStmt(whileStmt *ast.WhileStmt) (objects.Object, error) {
	for {
		cond, err := i.evalExpr(whileStmt.Condition)
		if err != nil {
			return nil, err
		}
		if !objects.IsTruthy(cond) {
			break
		}
		if _, err = i.evalStmt(whileStmt.Body); err != nil {
			if _, ok := err.(*objects.BreakError); ok {
				break
			}
			return nil, err
		}
	}
	return nil, nil
}

// Expressions

func (i *Interpreter) evalExpr(expr ast.Expr) (objects.Object, error) {
	switch e := expr.(type) {
	case *ast.BinaryExpr:
		return i.visitBinaryExpr(e)
	case *ast.GroupingExpr:
		return i.evalExpr(e.Expr)
	case *ast.LiteralExpr:
		return literalToObject(e.Value), nil
	case *ast.ArrayLiteralExpr:
		return i.visitArrayLiteralExpr(e)
	case *ast.HashLiteralExpr:
		return i.visitHashLiteralExpr(e)
	case *ast.UnaryExpr:
		return i.visitUnaryExpr(e)
	case *ast.VariableExpr:
		return i.visitVariableExpr(e)
	case *ast.AssignExpr:
		return i.visitAssignExpr(e)
	case *ast.SetIndexExpr:
		return i.visitSetIndexExpr(e)
	case *ast.LogicalExpr:
		return i.visitLogicalExpr(e)
	case *ast.CallExpr:
		return i.visitCallExpr(e)
	case *ast.GetExpr:
		return i.visitGetExpr(e)
	case *ast.SetExpr:
		return i.visitSetExpr(e)
	case *ast.IndexExpr:
		return i.visitIndexExpr(e)
	case *ast.ThisExpr:
		return i.visitThisExpr(e)
	case *ast.SuperExpr:
		return i.visitSuperExpr(e)
	default:
		return nil, errors.New("invalid expression")
	}
}

func (i *Interpreter) visitBinaryExpr(exp *ast.BinaryExpr) (objects.Object, error) {
	right, err := i.evalExpr(exp.Right)
	if err != nil {
		return nil, err
	}
	left, err := i.evalExpr(exp.Left)
	if err != nil {
		return nil, err
	}
	switch exp.Operator.Type {
	case token.PLUS:
		switch l := left.(type) {
		case *objects.Number:
			if r, ok := right.(*objects.Number); ok {
				return objects.NewNumber(l.Value + r.Value), nil
			}
			if r, ok := right.(*objects.String); ok {
				return objects.NewString(fmt.Sprintf("%g%s", l.Value, r.Value)), nil
			}
		case *objects.String:
			if r, ok := right.(*objects.String); ok {
				return objects.NewString(l.Value + r.Value), nil
			}
			if r, ok := right.(*objects.Number); ok {
				return objects.NewString(l.Value + fmt.Sprintf("%g", r.Value)), nil
			}
		}
		return nil, i.runtimeError(exp.Operator, fmt.Sprintf("Operands to '+' must both be numbers or both be strings or one string and other number (left: %T, right: %T).", left, right))
	case token.MINUS, token.STAR, token.SLASH, token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL:
		lf, lok := left.(*objects.Number)
		rf, rok := right.(*objects.Number)
		if !lok || !rok {
			return nil, i.runtimeError(exp.Operator, "Operands must be numbers.")
		}
		switch exp.Operator.Type {
		case token.MINUS:
			return objects.NewNumber(lf.Value - rf.Value), nil
		case token.STAR:
			return objects.NewNumber(lf.Value * rf.Value), nil
		case token.SLASH:
			if rf.Value == 0 {
				return nil, i.runtimeError(exp.Operator, "Division by zero.")
			}
			return objects.NewNumber(lf.Value / rf.Value), nil
		case token.GREATER:
			return objects.NewBool(lf.Value > rf.Value), nil
		case token.GREATER_EQUAL:
			return objects.NewBool(lf.Value >= rf.Value), nil
		case token.LESS:
			return objects.NewBool(lf.Value < rf.Value), nil
		case token.LESS_EQUAL:
			return objects.NewBool(lf.Value <= rf.Value), nil
		}
	case token.EQUAL_EQUAL:
		return objects.NewBool(objects.IsEqual(left, right)), nil
	case token.BANG_EQUAL:
		return objects.NewBool(!objects.IsEqual(left, right)), nil
	}
	return nil, i.runtimeError(exp.Operator, "Invalid operator.")
}

func (i *Interpreter) visitUnaryExpr(unary *ast.UnaryExpr) (objects.Object, error) {
	right, err := i.evalExpr(unary.Expr)
	if err != nil {
		return nil, err
	}
	switch unary.Operator.Type {
	case token.MINUS:
		num, ok := right.(*objects.Number)
		if !ok {
			return nil, i.runtimeError(unary.Operator, "Operand must be a number.")
		}
		return objects.NewNumber(-num.Value), nil
	case token.BANG:
		return objects.NewBool(!objects.IsTruthy(right)), nil
	}
	return nil, i.runtimeError(unary.Operator, "Invalid operator.")
}

func (i *Interpreter) visitVariableExpr(variable *ast.VariableExpr) (objects.Object, error) {
	return i.findVariable(variable, variable.Name)
}

func (i *Interpreter) visitThisExpr(t *ast.ThisExpr) (objects.Object, error) {
	return i.findVariable(t, t.Keyword)
}

func (i *Interpreter) findVariable(expr ast.Expr, name *token.Token) (objects.Object, error) {
	if dist, ok := i.locals[expr]; ok {
		val, err := i.environment.GetAt(dist, name.Lexeme)
		if err != nil {
			return nil, i.runtimeError(name, err.Error())
		}
		return val, nil
	}
	val, err := i.globals.Get(name.Lexeme)
	if err != nil {
		return nil, i.runtimeError(name, err.Error())
	}
	return val, nil
}

func (i *Interpreter) visitAssignExpr(assign *ast.AssignExpr) (objects.Object, error) {
	value, err := i.evalExpr(assign.Value)
	if err != nil {
		return nil, err
	}
	if dist, ok := i.locals[assign]; ok {
		if err := i.environment.AssignAt(dist, assign.Name.Lexeme, value); err != nil {
			return nil, i.runtimeError(assign.Name, err.Error())
		}
	} else {
		if err := i.globals.Assign(assign.Name.Lexeme, value); err != nil {
			return nil, i.runtimeError(assign.Name, err.Error())
		}
	}
	return value, nil
}

func (i *Interpreter) visitArrayLiteralExpr(array *ast.ArrayLiteralExpr) (objects.Object, error) {
	items := make([]objects.Object, 0, len(array.Elements))
	for _, el := range array.Elements {
		val, err := i.evalExpr(el)
		if err != nil {
			return nil, err
		}
		items = append(items, val)
	}
	return &objects.Array{Elements: items}, nil
}

func (i *Interpreter) visitHashLiteralExpr(hash *ast.HashLiteralExpr) (objects.Object, error) {
	table := objects.NewHash()
	for _, pair := range hash.Pairs {
		keyVal, err := i.evalExpr(pair.Key)
		if err != nil {
			return nil, err
		}
		keyStr, ok := keyVal.(*objects.String)
		if !ok {
			return nil, i.runtimeError(hash.Brace, "Hash map keys must be strings.")
		}
		valueVal, err := i.evalExpr(pair.Value)
		if err != nil {
			return nil, err
		}
		table.Set(keyStr.Value, valueVal)
	}
	return table, nil
}

func (i *Interpreter) visitIndexExpr(idx *ast.IndexExpr) (objects.Object, error) {
	obj, err := i.evalExpr(idx.Object)
	if err != nil {
		return nil, err
	}
	switch target := obj.(type) {
	case *objects.Array:
		indexVal, err := i.evalExpr(idx.Index)
		if err != nil {
			return nil, err
		}
		indexNum, ok := indexVal.(*objects.Number)
		if !ok {
			return nil, i.runtimeError(idx.Bracket, "Index must be a number.")
		}
		intIndex := int(indexNum.Value)
		if float64(intIndex) != indexNum.Value {
			return nil, i.runtimeError(idx.Bracket, "Index must be an integer.")
		}
		val, err := target.Get(intIndex)
		if err != nil {
			return nil, i.runtimeError(idx.Bracket, err.Error())
		}
		return val, nil
	case *objects.Hash:
		keyVal, err := i.evalExpr(idx.Index)
		if err != nil {
			return nil, err
		}
		key, ok := keyVal.(*objects.String)
		if !ok {
			return nil, i.runtimeError(idx.Bracket, "Hash map keys must be strings.")
		}
		val, ok := target.Get(key.Value)
		if !ok {
			return nil, i.runtimeError(idx.Bracket, "Key '"+key.Value+"' not found in hash map.")
		}
		return val, nil
	default:
		return nil, i.runtimeError(idx.Bracket, "Indexing target must be an array or hash map.")
	}
}

func (i *Interpreter) visitSetIndexExpr(setIdx *ast.SetIndexExpr) (objects.Object, error) {
	obj, err := i.evalExpr(setIdx.Object)
	if err != nil {
		return nil, err
	}
	switch target := obj.(type) {
	case *objects.Array:
		indexVal, err := i.evalExpr(setIdx.Index)
		if err != nil {
			return nil, err
		}
		indexNum, ok := indexVal.(*objects.Number)
		if !ok {
			return nil, i.runtimeError(setIdx.Bracket, "Index must be a number.")
		}
		intIndex := int(indexNum.Value)
		if float64(intIndex) != indexNum.Value {
			return nil, i.runtimeError(setIdx.Bracket, "Index must be an integer.")
		}
		val, err := i.evalExpr(setIdx.Value)
		if err != nil {
			return nil, err
		}
		if err := target.Set(intIndex, val); err != nil {
			return nil, i.runtimeError(setIdx.Bracket, err.Error())
		}
		return val, nil
	case *objects.Hash:
		keyVal, err := i.evalExpr(setIdx.Index)
		if err != nil {
			return nil, err
		}
		key, ok := keyVal.(*objects.String)
		if !ok {
			return nil, i.runtimeError(setIdx.Bracket, "Hash map keys must be strings.")
		}
		val, err := i.evalExpr(setIdx.Value)
		if err != nil {
			return nil, err
		}
		target.Set(key.Value, val)
		return val, nil
	default:
		return nil, i.runtimeError(setIdx.Bracket, "Index assignment target must be an array or hash map.")
	}
}

func (i *Interpreter) visitLogicalExpr(logical *ast.LogicalExpr) (objects.Object, error) {
	left, err := i.evalExpr(logical.Left)
	if err != nil {
		return nil, err
	}
	if (logical.Operator.Type == token.OR && objects.IsTruthy(left)) ||
		(logical.Operator.Type == token.AND && !objects.IsTruthy(left)) {
		return left, nil
	}
	return i.evalExpr(logical.Right)
}

func (i *Interpreter) visitCallExpr(call *ast.CallExpr) (objects.Object, error) {
	callee, err := i.evalExpr(call.Callee)
	if err != nil {
		return nil, err
	}
	args := make([]objects.Object, 0, len(call.Arguments))
	for _, arg := range call.Arguments {
		val, err := i.evalExpr(arg)
		if err != nil {
			return nil, err
		}
		args = append(args, val)
	}
	callable, ok := callee.(objects.Callable)
	if !ok {
		return nil, i.runtimeError(call.ClosingParen, "Can only call functions or classes.")
	}
	if callable.Arity() != len(args) {
		return nil, i.runtimeError(call.ClosingParen, "Expected "+strconv.Itoa(callable.Arity())+" arguments but got "+strconv.Itoa(len(args))+".")
	}
	result, err := callable.Call(i, args)
	if err != nil {
		return nil, i.runtimeError(call.ClosingParen, err.Error())
	}
	return result, nil
}

func (i *Interpreter) visitGetExpr(get *ast.GetExpr) (objects.Object, error) {
	object, err := i.evalExpr(get.Object)
	if err != nil {
		return nil, err
	}
	instance, ok := object.(*objects.ClassInstance)
	if !ok {
		return nil, i.runtimeError(get.Name, "Only instances have properties.")
	}
	value, err := instance.Get(get.Name)
	if err != nil {
		return nil, i.runtimeError(get.Name, err.Error())
	}
	return value, nil
}

func (i *Interpreter) visitSetExpr(set *ast.SetExpr) (objects.Object, error) {
	object, err := i.evalExpr(set.Object)
	if err != nil {
		return nil, err
	}
	instance, ok := object.(*objects.ClassInstance)
	if !ok {
		return nil, i.runtimeError(set.Name, "Only instances have fields.")
	}
	value, err := i.evalExpr(set.Value)
	if err != nil {
		return nil, err
	}
	if err := instance.Set(set.Name, value); err != nil {
		return nil, i.runtimeError(set.Name, err.Error())
	}
	return value, nil
}

func (i *Interpreter) visitSuperExpr(super *ast.SuperExpr) (objects.Object, error) {
	dist, ok := i.locals[super]
	if !ok {
		return nil, i.runtimeError(super.Keyword, "Superclass not found.")
	}
	superClassObject, err := i.environment.GetAt(dist, "super")
	if err != nil {
		return nil, i.runtimeError(super.Keyword, err.Error())
	}
	superclass, ok := superClassObject.(*objects.Class)
	if !ok {
		return nil, i.runtimeError(super.Keyword, "Superclass must be a class.")
	}
	method, ok := superclass.LookupMethod(super.Method.Lexeme)
	if !ok {
		return nil, i.runtimeError(super.Method, "Undefined property '"+super.Method.Lexeme+"'.")
	}
	if method == nil {
		return nil, i.runtimeError(super.Method, "Undefined property '"+super.Method.Lexeme+"'.")
	}

	thisInstance, err := i.environment.GetAt(dist-1, "this")
	if err != nil {
		return nil, i.runtimeError(super.Keyword, err.Error())
	}
	subClassInstance, ok := thisInstance.(*objects.ClassInstance)
	if !ok {
		return nil, i.runtimeError(super.Keyword, "'this' must be a class instance.")
	}
	return method.Bind(subClassInstance), nil
}

// Utilities

func literalToObject(v interface{}) objects.Object {
	// Accept already-converted values to avoid double-wrapping.
	if obj, ok := v.(objects.Object); ok {
		return obj
	}
	switch val := v.(type) {
	case nil:
		return objects.NewNil()
	case bool:
		return objects.NewBool(val)
	case int:
		return objects.NewNumber(float64(val))
	case int64:
		return objects.NewNumber(float64(val))
	case int32:
		return objects.NewNumber(float64(val))
	case uint:
		return objects.NewNumber(float64(val))
	case uint64:
		return objects.NewNumber(float64(val))
	case uint32:
		return objects.NewNumber(float64(val))
	case float64:
		return objects.NewNumber(val)
	case string:
		return objects.NewString(val)
	default:
		panic(fmt.Sprintf("unsupported literal conversion for type %T", v))
	}
}

func (i *Interpreter) runtimeError(tok *token.Token, message string) error {
	return &objects.RuntimeError{
		Token:   tok,
		Message: message,
	}
}
