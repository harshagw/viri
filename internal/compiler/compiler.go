package compiler

import (
	"fmt"

	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/code"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/token"
)

type CompilationScope struct {
	instructions code.Instructions
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []objects.Object
}

// ClassCompiler tracks state while compiling a class.
// This enables compile-time validation of this/super usage
// and provides context for future static field support.
type ClassCompiler struct {
	name            string
	hasSuperClass   bool
	enclosing       *ClassCompiler // for nested classes
	isCompilingInit bool           // true when compiling the init method
}

type Compiler struct {
	constants   []objects.Object
	symbolTable *SymbolTable
	loopStack   *LoopStack

	scopes     []CompilationScope
	scopeIndex int

	classCompiler     *ClassCompiler // nil when not compiling a class
	diagnosticHandler objects.DiagnosticHandler
}

func New(diagnosticHandler objects.DiagnosticHandler) *Compiler {
	return NewWithState(diagnosticHandler, nil)
}

func NewWithState(diagnosticHandler objects.DiagnosticHandler, symbolTable *SymbolTable) *Compiler {
	mainScope := CompilationScope{
		instructions: code.Instructions{},
	}

	if symbolTable == nil {
		symbolTable = NewSymbolTable()
	}

	// Ensure native functions are always defined in the symbol table
	for i, nativeFn := range objects.NativeFunctions {
		if _, exists := symbolTable.store[nativeFn.Name]; !exists {
			symbolTable.DefineNative(i, nativeFn.Name)
		}
	}

	return &Compiler{
		constants:         []objects.Object{},
		diagnosticHandler: diagnosticHandler,
		symbolTable:       symbolTable,
		loopStack:         NewLoopStack(),
		scopes:            []CompilationScope{mainScope},
		scopeIndex:        0,
	}
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(),
		Constants:    c.constants,
	}
}

func (c *Compiler) currentInstructions() code.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) enterScope(functionName string) {
	scope := CompilationScope{
		instructions: code.Instructions{},
	}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex++
	c.symbolTable = NewFunctionScope(c.symbolTable, functionName)
}

func (c *Compiler) leaveScope() code.Instructions {
	instructions := c.currentInstructions()

	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--
	c.symbolTable = c.symbolTable.Outer

	return instructions
}

func (c *Compiler) Compile(node interface{}) error {
	switch node := node.(type) {
	case ast.Stmt:
		return c.compileStatement(node)
	case ast.Expr:
		return c.compileExpression(node)
	default:
		return fmt.Errorf("unknown node type: %T", node)
	}
}

func (c *Compiler) compileStatement(stmt ast.Stmt) error {
	switch stmt := stmt.(type) {
	case *ast.ExprStmt:
		if err := c.compileExpression(stmt.Expr); err != nil {
			return err
		}
		c.emit(code.OpPop)
		return nil

	case *ast.BlockStmt:
		// Create a new block scope (same frame, new lexical scope)
		c.symbolTable = NewBlockScope(c.symbolTable)
		for _, s := range stmt.Statements {
			if err := c.compileStatement(s); err != nil {
				return err
			}
		}
		// Restore parent scope
		c.symbolTable = c.symbolTable.Outer
		return nil

	case *ast.IfStmt:
		if err := c.compileExpression(stmt.Condition); err != nil {
			return err
		}

		// Emit an OpJumpNotTruthy with a placeholder offset
		jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)

		if err := c.compileStatement(stmt.ThenBranch); err != nil {
			return err
		}

		if stmt.ElseBranch != nil {
			// Emit an OpJump to skip the else branch after executing if branch
			jumpPos := c.emit(code.OpJump, 9999)

			// Patch the OpJumpNotTruthy to jump to else branch
			c.changeOperand(jumpNotTruthyPos, len(c.currentInstructions()))

			if err := c.compileStatement(stmt.ElseBranch); err != nil {
				return err
			}

			// Patch the OpJump to jump past else branch
			c.changeOperand(jumpPos, len(c.currentInstructions()))
		} else {
			// No else branch - patch jump to skip if branch
			c.changeOperand(jumpNotTruthyPos, len(c.currentInstructions()))
		}

		return nil

	case *ast.VarDeclStmt:
		symbol := c.symbolTable.Define(stmt.Name.Lexeme, stmt.IsConst)

		if stmt.Initializer != nil {
			// Check if the initializer is a function expression for recursive support
			// We don't want to body of the function to create a block scope, so we compile the function directly
			if fnExpr, ok := stmt.Initializer.(*ast.FunctionExpr); ok {
				if err := c.compileFunction(fnExpr.Params, fnExpr.Body, stmt.Name.Lexeme); err != nil {
					return err
				}
			} else {
				if err := c.compileExpression(stmt.Initializer); err != nil {
					return err
				}
			}
		} else {
			c.emit(code.OpNil) // Default to nil if no initializer
		}

		c.emitSetSymbol(symbol)
		return nil

	case *ast.PrintStmt:
		if err := c.compileExpression(stmt.Expr); err != nil {
			return err
		}
		c.emit(code.OpPrint)
		return nil

	case *ast.WhileStmt:
		// Record the start of the loop (where continue jumps to)
		loopStart := len(c.currentInstructions())
		c.loopStack.Push(loopStart)

		// Compile condition
		if err := c.compileExpression(stmt.Condition); err != nil {
			return err
		}

		// Jump out of loop if condition is false
		exitJump := c.emit(code.OpJumpNotTruthy, 9999)

		// Compile body
		if err := c.compileStatement(stmt.Body); err != nil {
			return err
		}

		// Jump back to condition
		c.emit(code.OpJump, loopStart)

		// Patch jumps and exit loop context
		loopEnd := len(c.currentInstructions())
		c.changeOperand(exitJump, loopEnd)
		c.patchBreakJumps(loopEnd)
		c.loopStack.Pop()
		return nil

	case *ast.ForStmt:
		if stmt.Initializer != nil {
			if err := c.compileStatement(stmt.Initializer); err != nil {
				return err
			}
		}

		// Record the start of the condition check
		loopStart := len(c.currentInstructions())

		// For for-loops, continue target is not yet known (it's before increment)
		c.loopStack.Push(-1)

		var exitJump int
		if stmt.Condition != nil {
			if err := c.compileExpression(stmt.Condition); err != nil {
				return err
			}
			exitJump = c.emit(code.OpJumpNotTruthy, 9999)
		}

		// Compile body
		if err := c.compileStatement(stmt.Body); err != nil {
			return err
		}

		// Set and patch continue jumps to point here (before increment)
		c.patchContinueJumps(len(c.currentInstructions()))

		// Compile increment (if present)
		if stmt.Increment != nil {
			if err := c.compileExpression(stmt.Increment); err != nil {
				return err
			}
			c.emit(code.OpPop) // discard increment result
		}

		// Jump back to condition
		c.emit(code.OpJump, loopStart)

		// Patch exit and break jumps
		loopEnd := len(c.currentInstructions())
		if stmt.Condition != nil {
			c.changeOperand(exitJump, loopEnd)
		}
		c.patchBreakJumps(loopEnd)
		c.loopStack.Pop()
		return nil

	case *ast.BreakStmt:
		if !c.loopStack.IsInLoop() {
			return c.error(stmt.Keyword, "break statement outside of loop")
		}
		jumpPos := c.emit(code.OpJump, 9999)
		c.loopStack.AddBreakJump(jumpPos)
		return nil

	case *ast.ContinueStmt:
		if !c.loopStack.IsInLoop() {
			return c.error(stmt.Keyword, "continue statement outside of loop")
		}
		if c.loopStack.ContinuePos() == -1 {
			// For for-loops, continuePos isn't known yet, record jump for patching
			jumpPos := c.emit(code.OpJump, 9999)
			c.loopStack.AddContinueJump(jumpPos)
		} else {
			// For while-loops, continuePos is already known
			c.emit(code.OpJump, c.loopStack.ContinuePos())
		}
		return nil

	case *ast.FunctionStmt:
		// Define the function name in the current scope
		symbol := c.symbolTable.Define(stmt.Name.Lexeme, false)

		// Compile the function body
		if err := c.compileFunction(stmt.Params, stmt.Body, stmt.Name.Lexeme); err != nil {
			return err
		}

		c.emitSetSymbol(symbol)
		return nil

	case *ast.ReturnStmt:
		// In init methods, always return 'this' regardless of what the user wrote
		if c.classCompiler != nil && c.classCompiler.isCompilingInit {
			c.emit(code.OpGetLocal, 0) // 'this' is always local 0
			c.emit(code.OpReturnValue)
		} else if stmt.Value != nil {
			if err := c.compileExpression(stmt.Value); err != nil {
				return err
			}
			c.emit(code.OpReturnValue)
		} else {
			c.emit(code.OpReturn)
		}
		return nil

	case *ast.ClassStmt:
		return c.compileClassStmt(stmt)

	default:
		return fmt.Errorf("unsupported statement type: %T", stmt)
	}
}

func (c *Compiler) compileExpression(node ast.Expr) error {
	switch node := node.(type) {
	case *ast.LiteralExpr:
		switch value := node.Value.(type) {
		case int:
			integer := &objects.Number{Value: float64(value)}
			constIdx := c.addConstant(integer)
			c.emit(code.OpGetConstant, constIdx)
		case float64:
			number := &objects.Number{Value: value}
			constIdx := c.addConstant(number)
			c.emit(code.OpGetConstant, constIdx)
		case string:
			str := &objects.String{Value: value}
			constIdx := c.addConstant(str)
			c.emit(code.OpGetConstant, constIdx)
		case bool:
			if value {
				c.emit(code.OpTrue)
			} else {
				c.emit(code.OpFalse)
			}
		case nil:
			c.emit(code.OpNil)
		}

	case *ast.UnaryExpr:
		if err := c.compileExpression(node.Expr); err != nil {
			return err
		}

		switch node.Operator.Type {
		case token.BANG:
			c.emit(code.OpBang)
		case token.MINUS:
			c.emit(code.OpMinus)
		default:
			return c.error(node.Operator, fmt.Sprintf("unknown unary operator %s", node.Operator.Lexeme))
		}

	case *ast.GroupingExpr:
		return c.compileExpression(node.Expr)

	case *ast.BinaryExpr:
		// a < b  => swap operands, then b > a
		if node.Operator.Type == token.LESS {
			if err := c.compileExpression(node.Right); err != nil {
				return err
			}
			if err := c.compileExpression(node.Left); err != nil {
				return err
			}
			c.emit(code.OpGreaterThan)
			return nil
		}

		// a <= b  => !(a > b)
		if node.Operator.Type == token.LESS_EQUAL {
			if err := c.compileExpression(node.Left); err != nil {
				return err
			}
			if err := c.compileExpression(node.Right); err != nil {
				return err
			}
			c.emit(code.OpGreaterThan)
			c.emit(code.OpBang)
			return nil
		}

		// a >= b  => !(a < b) => !(b > a)
		if node.Operator.Type == token.GREATER_EQUAL {
			if err := c.compileExpression(node.Right); err != nil {
				return err
			}
			if err := c.compileExpression(node.Left); err != nil {
				return err
			}
			c.emit(code.OpGreaterThan)
			c.emit(code.OpBang)
			return nil
		}

		if err := c.compileExpression(node.Left); err != nil {
			return err
		}
		if err := c.compileExpression(node.Right); err != nil {
			return err
		}

		switch node.Operator.Type {
		case token.PLUS:
			c.emit(code.OpAdd)
		case token.MINUS:
			c.emit(code.OpSub)
		case token.STAR:
			c.emit(code.OpMul)
		case token.SLASH:
			c.emit(code.OpDiv)
		case token.GREATER:
			c.emit(code.OpGreaterThan)
		case token.EQUAL_EQUAL:
			c.emit(code.OpEqual)
		case token.BANG_EQUAL:
			c.emit(code.OpNotEqual)
		default:
			return c.error(node.Operator, fmt.Sprintf("unknown operator %s", node.Operator.Lexeme))
		}

	case *ast.VariableExpr:
		symbol, ok := c.symbolTable.Resolve(node.Name.Lexeme)
		if !ok {
			return c.error(node.Name, fmt.Sprintf("undefined variable %s", node.Name.Lexeme))
		}
		c.emitGetSymbol(symbol)

	case *ast.AssignExpr:
		// Prevent assignment to 'this'
		if node.Name.Lexeme == "this" {
			return c.error(node.Name, "cannot assign to 'this'")
		}

		symbol, ok := c.symbolTable.Resolve(node.Name.Lexeme)
		if !ok {
			return c.error(node.Name, fmt.Sprintf("undefined variable %s", node.Name.Lexeme))
		}
		if symbol.IsConst {
			return c.error(node.Name, fmt.Sprintf("cannot assign to constant %s", node.Name.Lexeme))
		}

		if err := c.compileExpression(node.Value); err != nil {
			return err
		}

		c.emitSetSymbol(symbol)
		// Assignment is an expression, so we need to leave the value on the stack
		c.emitGetSymbol(symbol)

	case *ast.ArrayLiteralExpr:
		for _, elem := range node.Elements {
			if err := c.compileExpression(elem); err != nil {
				return err
			}
		}
		c.emit(code.OpArray, len(node.Elements))

	case *ast.HashLiteralExpr:
		for _, pair := range node.Pairs {
			if err := c.compileExpression(pair.Key); err != nil {
				return err
			}
			if err := c.compileExpression(pair.Value); err != nil {
				return err
			}
		}
		c.emit(code.OpHash, len(node.Pairs)*2)

	case *ast.IndexExpr:
		if err := c.compileExpression(node.Object); err != nil {
			return err
		}
		if err := c.compileExpression(node.Index); err != nil {
			return err
		}
		c.emit(code.OpIndex)

	case *ast.SetIndexExpr:
		if err := c.compileExpression(node.Object); err != nil {
			return err
		}
		if err := c.compileExpression(node.Index); err != nil {
			return err
		}
		if err := c.compileExpression(node.Value); err != nil {
			return err
		}
		c.emit(code.OpSetIndex)

	case *ast.LogicalExpr:
		if err := c.compileExpression(node.Left); err != nil {
			return err
		}

		switch node.Operator.Type {
		case token.AND:
			// Short-circuit AND: if left is falsy, return left; else return right
			c.emit(code.OpDup)
			jumpIfFalsy := c.emit(code.OpJumpNotTruthy, 9999)
			c.emit(code.OpPop)
			if err := c.compileExpression(node.Right); err != nil {
				return err
			}
			c.changeOperand(jumpIfFalsy, len(c.currentInstructions()))

		case token.OR:
			// Short-circuit OR: if left is truthy, return left; else return right
			c.emit(code.OpDup)
			jumpIfFalsy := c.emit(code.OpJumpNotTruthy, 9999)
			jumpToEnd := c.emit(code.OpJump, 9999)
			c.changeOperand(jumpIfFalsy, len(c.currentInstructions()))
			c.emit(code.OpPop)
			if err := c.compileExpression(node.Right); err != nil {
				return err
			}
			c.changeOperand(jumpToEnd, len(c.currentInstructions()))
		}

	case *ast.CallExpr:
		if err := c.compileExpression(node.Callee); err != nil {
			return err
		}

		for _, arg := range node.Arguments {
			if err := c.compileExpression(arg); err != nil {
				return err
			}
		}

		c.emit(code.OpCall, len(node.Arguments))

	case *ast.FunctionExpr:
		// Anonymous function - no name for recursion
		if err := c.compileFunction(node.Params, node.Body, ""); err != nil {
			return err
		}

	case *ast.GetExpr:
		if err := c.compileExpression(node.Object); err != nil {
			return err
		}
		nameIdx := c.addConstant(&objects.String{Value: node.Name.Lexeme})
		c.emit(code.OpGetProperty, nameIdx)

	case *ast.SetExpr:
		if err := c.compileExpression(node.Object); err != nil {
			return err
		}
		if err := c.compileExpression(node.Value); err != nil {
			return err
		}
		nameIdx := c.addConstant(&objects.String{Value: node.Name.Lexeme})
		c.emit(code.OpSetProperty, nameIdx)

	case *ast.ThisExpr:
		// Validate: must be inside a class
		if c.classCompiler == nil {
			return c.error(node.Keyword, "cannot use 'this' outside of a class")
		}

		// Validate: must be inside a method (this is defined)
		symbol, ok := c.symbolTable.Resolve("this")
		if !ok {
			return c.error(node.Keyword, "cannot use 'this' outside of a method")
		}

		c.emitGetSymbol(symbol)

	case *ast.SuperExpr:
		// Validate: must be inside a class
		if c.classCompiler == nil {
			return c.error(node.Keyword, "cannot use 'super' outside of a class")
		}

		// Validate: class must have a superclass
		if !c.classCompiler.hasSuperClass {
			return c.error(node.Keyword, "cannot use 'super' in a class with no superclass")
		}

		// Validate: must be inside a method
		symbol, ok := c.symbolTable.Resolve("this")
		if !ok {
			return c.error(node.Keyword, "cannot use 'super' outside of a method")
		}

		// Push 'this' instance (super method will be bound to it)
		c.emitGetSymbol(symbol)

		// Emit OpGetSuper with method name
		nameIdx := c.addConstant(&objects.String{Value: node.Method.Lexeme})
		c.emit(code.OpGetSuper, nameIdx)
	}
	return nil
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := len(c.currentInstructions())
	c.scopes[c.scopeIndex].instructions = append(c.scopes[c.scopeIndex].instructions, ins...)
	return pos
}

func (c *Compiler) addConstant(obj objects.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) changeOperand(opPos int, operands ...int) {
	op := code.Opcode(c.currentInstructions()[opPos])
	newInstruction := code.Make(op, operands...)
	for i := 0; i < len(newInstruction); i++ {
		c.scopes[c.scopeIndex].instructions[opPos+i] = newInstruction[i]
	}
}

func (c *Compiler) emitGetSymbol(s Symbol) {
	switch s.Scope {
	case GlobalScope:
		c.emit(code.OpGetGlobal, s.Index)
	case LocalScope:
		c.emit(code.OpGetLocal, s.Index)
	case NativeScope:
		c.emit(code.OpGetNative, s.Index)
	case FreeScope:
		c.emit(code.OpGetFree, s.Index)
	case FunctionScope:
		c.emit(code.OpGetCurrentClosure)
	}
}

func (c *Compiler) emitSetSymbol(s Symbol) {
	switch s.Scope {
	case GlobalScope:
		c.emit(code.OpSetGlobal, s.Index)
	case LocalScope:
		c.emit(code.OpSetLocal, s.Index)
	case FreeScope:
		c.emit(code.OpSetFree, s.Index)
	}
}

func (c *Compiler) compileFunction(params []*token.Token, body *ast.BlockStmt, functionName string) error {
	c.enterScope(functionName)

	// Define parameters as local variables
	for _, param := range params {
		c.symbolTable.Define(param.Lexeme, false)
	}

	// Compile the function body statements directly (don't create an extra block scope)
	for _, stmt := range body.Statements {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}

	// If the function doesn't have an explicit return, emit OpReturn
	c.emit(code.OpReturn)

	fn := &objects.CompiledFunction{
		Instructions:  c.currentInstructions(),
		NumLocals:     c.symbolTable.NumDefinitions(),
		NumParameters: len(params),
		Name:          functionName,
	}
	freeSymbols := c.symbolTable.FreeSymbols

	c.leaveScope()
	c.emitClosure(fn, freeSymbols)
	return nil
}

// emitClosure emits free variable loading instructions and OpGetClosure.
func (c *Compiler) emitClosure(fn *objects.CompiledFunction, freeSymbols []Symbol) {
	for _, s := range freeSymbols {
		switch s.Scope {
		case LocalScope:
			c.emit(code.OpMakeCell, s.Index)
		case FreeScope:
			c.emit(code.OpGetFree, s.Index)
		}
	}

	constIdx := c.addConstant(fn)
	c.emit(code.OpGetClosure, constIdx, len(freeSymbols))
}

func (c *Compiler) patchBreakJumps(loopEnd int) {
	for _, jumpPos := range c.loopStack.BreakJumps() {
		c.changeOperand(jumpPos, loopEnd)
	}
}

func (c *Compiler) patchContinueJumps(continueTarget int) {
	c.loopStack.SetContinuePos(continueTarget)
	for _, jumpPos := range c.loopStack.ContinueJumps() {
		c.changeOperand(jumpPos, continueTarget)
	}
}

func (c *Compiler) error(tok *token.Token, message string) error {
	if c.diagnosticHandler != nil && tok != nil {
		c.diagnosticHandler.Error(*tok, message)
	}
	return fmt.Errorf("%s", message)
}

func (c *Compiler) compileClassStmt(stmt *ast.ClassStmt) error {
	className := stmt.Name.Lexeme

	// Define class name in symbol table (allows recursive references)
	symbol := c.symbolTable.Define(className, false)

	// Enter class compilation context
	enclosingClass := c.classCompiler
	c.classCompiler = &ClassCompiler{
		name:          className,
		hasSuperClass: stmt.SuperClass != nil,
		enclosing:     enclosingClass,
	}

	// Push superclass or nil to stack
	if stmt.SuperClass != nil {
		if err := c.compileExpression(stmt.SuperClass); err != nil {
			return err
		}
		// Validate: can't inherit from self
		if stmt.SuperClass.Name.Lexeme == className {
			return c.error(stmt.SuperClass.Name, "a class cannot inherit from itself")
		}
	} else {
		c.emit(code.OpNil)
	}

	// Compile each method (pushes closures to stack)
	for _, method := range stmt.Methods {
		if err := c.compileMethod(method); err != nil {
			return err
		}
	}

	nameIdx := c.addConstant(&objects.String{Value: className})
	c.emit(code.OpClass, nameIdx, len(stmt.Methods))

	c.emitSetSymbol(symbol)

	// Exit class compilation context
	c.classCompiler = enclosingClass

	return nil
}

// compileMethod compiles a method with 'this' as implicit first parameter.
func (c *Compiler) compileMethod(method *ast.FunctionStmt) error {
	c.enterScope(method.Name.Lexeme)

	isInit := method.Name.Lexeme == "init"
	if isInit {
		c.classCompiler.isCompilingInit = true
	}

	// Define 'this' as first local (index 0)
	c.symbolTable.Define("this", true)

	for _, param := range method.Params {
		c.symbolTable.Define(param.Lexeme, false)
	}

	for _, stmt := range method.Body.Statements {
		if err := c.compileStatement(stmt); err != nil {
			return err
		}
	}

	// For init methods, implicitly return 'this' instead of nil
	if isInit {
		c.emit(code.OpGetLocal, 0) // 'this' is always local 0
		c.emit(code.OpReturnValue)
		c.classCompiler.isCompilingInit = false
	} else {
		c.emit(code.OpReturn)
	}

	fn := &objects.CompiledFunction{
		Instructions:  c.currentInstructions(),
		NumLocals:     c.symbolTable.NumDefinitions(),
		NumParameters: len(method.Params) + 1, // +1 for 'this'
		Name:          method.Name.Lexeme,
	}
	freeSymbols := c.symbolTable.FreeSymbols

	c.leaveScope()
	c.emitClosure(fn, freeSymbols)
	return nil
}
