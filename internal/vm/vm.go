package vm

import (
	"fmt"
	"io"
	"os"

	"github.com/harshagw/viri/internal/code"
	"github.com/harshagw/viri/internal/objects"
)

const StackSize = 2048
const GlobalsSize = 65536
const MaxFrames = 1024

// ModuleInstance represents a module at runtime
type ModuleInstance struct {
	Globals      []objects.Object // module-local globals
	Exports      []int            // export index -> global slot mapping
	MainFn       *objects.Closure // pre-created main closure for this module
	DebugInfoIdx int              // index into DebugInfo for line table and file path
}

type VM struct {
	constants []objects.Object
	debugInfo *objects.DebugInfo // debug information (line tables, file paths)

	stack []objects.Object
	sp    int // Always points to the next value. Top of stack is stack[sp-1]

	modules       []ModuleInstance // module instances with per-module globals
	numModules    int              // cached length for bounds checking
	currentModule int              // index of currently executing module

	frames      []*Frame
	framesIndex int // Always points to the next frame to be used. Top of frame is frames[framesIndex-1]

	onStep func()   // Debug callback, called before each opcode execution
	output []string // Capture print output
	stdout io.Writer
}

func New(program *objects.CompiledProgram) *VM {
	numModules := len(program.Modules)

	modules := make([]ModuleInstance, numModules)
	for i, compiledMod := range program.Modules {
		mainFn := &objects.CompiledFunction{
			Instructions: compiledMod.Instructions,
			NumLocals:    compiledMod.NumLocals,
			DebugInfoIdx: compiledMod.DebugInfoIdx,
			ModuleIdx:    i,
		}
		mainClosure := objects.NewClosure(mainFn, nil)

		modules[i] = ModuleInstance{
			Globals:      make([]objects.Object, compiledMod.NumGlobals),
			Exports:      compiledMod.Exports,
			MainFn:       mainClosure,
			DebugInfoIdx: compiledMod.DebugInfoIdx,
		}
	}

	vm := &VM{
		constants:   program.Constants,
		debugInfo:   program.DebugInfo,
		stack:       make([]objects.Object, StackSize),
		sp:          0,
		modules:     modules,
		numModules:  numModules,
		frames:      make([]*Frame, MaxFrames),
		framesIndex: 0,
		stdout:      os.Stdout,
	}

	if numModules > 0 {
		vm.frames[0] = NewFrame(modules[0].MainFn, 0)
		vm.framesIndex = 1
	}

	return vm
}

func (vm *VM) SetOnStep(fn func()) {
	vm.onStep = fn
}

// SetStdout redirects print output (mirrors Interpreter.SetStdout), so the
// wasm playground and in-process tests can capture what a program prints.
func (vm *VM) SetStdout(w io.Writer) {
	vm.stdout = w
}

// GetModuleGlobals returns the globals array for a specific module
func (vm *VM) GetModuleGlobals(moduleIdx int) []objects.Object {
	if moduleIdx < 0 || moduleIdx >= len(vm.modules) {
		return nil
	}
	return vm.modules[moduleIdx].Globals
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) pushFrame(f *Frame) error {
	if vm.framesIndex >= MaxFrames {
		return vm.runtimeError("Stack overflow.")
	}
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
	return nil
}

func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

func (vm *VM) push(o objects.Object) error {
	if vm.sp >= StackSize {
		return vm.runtimeError("Stack overflow.")
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

func (vm *VM) pop() objects.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	// Unwrap Cell if needed (for free variables)
	if cell, ok := o.(*objects.Cell); ok {
		return cell.Value
	}
	return o
}

func (vm *VM) LastPoppedStackElem() objects.Object {
	return unwrapCell(vm.stack[vm.sp])
}

func (vm *VM) runtimeError(message string) error {
	frame := vm.currentFrame()
	ip := frame.ip
	debugIdx := frame.cl.Fn.DebugInfoIdx

	line := vm.debugInfo.GetLine(debugIdx, ip)
	filePath := vm.debugInfo.GetFilePath(debugIdx)

	return &objects.VMRuntimeError{
		Message:  message,
		Line:     line,
		FilePath: filePath,
	}
}

func (vm *VM) RunProgram() error {
	// Execute each module in topological order
	for moduleIdx := 0; moduleIdx < vm.numModules; moduleIdx++ {
		vm.currentModule = moduleIdx

		vm.frames[0] = NewFrame(vm.modules[moduleIdx].MainFn, 0)
		vm.framesIndex = 1
		// Reserve main-frame slots for module-level block-scoped variables
		vm.sp = vm.modules[moduleIdx].MainFn.Fn.NumLocals

		if err := vm.runModule(moduleIdx); err != nil {
			return err
		}
	}
	return nil
}

func (vm *VM) runModule(moduleIdx int) error {
	var ip int
	var ins code.Instructions
	var op code.Opcode
	var frame *Frame

	frame = vm.currentFrame()
	ins = frame.cl.Fn.Instructions
	// Globals are per-module: a function compiled in module M must read and
	// write M's globals even when called from another module, so this is
	// rebound on every call and return.
	moduleGlobals := vm.modules[frame.cl.Fn.ModuleIdx].Globals

	for frame.ip < len(ins)-1 {
		frame.ip++
		ip = frame.ip
		op = code.Opcode(ins[ip])

		if vm.onStep != nil {
			vm.onStep()
		}

		switch op {
		case code.OpGetConstant:
			constIndex := readUint16(ins, ip)
			frame.ip += 2
			if err := vm.push(vm.constants[constIndex]); err != nil {
				return err
			}

		case code.OpAddNumber:
			// The checker proved both operands are numbers, so there is
			// nothing to dispatch on and no failure case.
			right := vm.pop().(*objects.Number)
			left := vm.pop().(*objects.Number)
			if err := vm.push(objects.NewNumber(left.Value + right.Value)); err != nil {
				return err
			}

		case code.OpAddString:
			// Likewise for the string meaning of '+'.
			right := vm.pop().(*objects.String)
			left := vm.pop().(*objects.String)
			if err := vm.push(objects.NewString(left.Value + right.Value)); err != nil {
				return err
			}

		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			if err := vm.executeBinaryOperation(op); err != nil {
				return err
			}

		case code.OpTrue:
			if err := vm.push(objects.TrueValue); err != nil {
				return err
			}

		case code.OpFalse:
			if err := vm.push(objects.FalseValue); err != nil {
				return err
			}

		case code.OpNil:
			if err := vm.push(objects.NilValue); err != nil {
				return err
			}

		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan, code.OpLess:
			if err := vm.executeComparison(op); err != nil {
				return err
			}

		case code.OpBang:
			if err := vm.executeBangOperator(); err != nil {
				return err
			}

		case code.OpMinus:
			if err := vm.executeMinusOperator(); err != nil {
				return err
			}

		case code.OpJump:
			pos := readUint16(ins, ip)
			frame.ip = pos - 1

		case code.OpJumpNotTruthy:
			pos := readUint16(ins, ip)
			frame.ip += 2
			// Conditions are bool: the checker rejects anything else, so this
			// is an unwrap rather than a truthiness coercion.
			condition := unwrapCell(vm.pop())
			if b, ok := condition.(*objects.Bool); !ok || !b.Value {
				frame.ip = pos - 1
			}

		case code.OpPop:
			vm.pop()

		case code.OpDup:
			// Don't unwrap - we want to duplicate the exact value (including Cells)
			if err := vm.push(vm.stack[vm.sp-1]); err != nil {
				return err
			}

		case code.OpSetGlobal:
			globalIndex := readUint16(ins, ip)
			frame.ip += 2
			moduleGlobals[globalIndex] = vm.pop()

		case code.OpGetGlobal:
			globalIndex := readUint16(ins, ip)
			frame.ip += 2
			if err := vm.push(moduleGlobals[globalIndex]); err != nil {
				return err
			}

		case code.OpGetModuleExport:
			targetModuleIdx := readUint16(ins, ip)
			exportIdx := readUint16(ins, ip+2)
			frame.ip += 4

			// Look up the global slot for this export
			slot := vm.modules[targetModuleIdx].Exports[exportIdx]
			value := vm.modules[targetModuleIdx].Globals[slot]
			if err := vm.push(value); err != nil {
				return err
			}

		case code.OpGetStdlibExport:
			constIdx := readUint16(ins, ip)
			frame.ip += 2

			// The stdlib export object is stored directly in constants
			value := vm.constants[constIdx]
			if err := vm.push(value); err != nil {
				return err
			}

		case code.OpSetLocal:
			localIndex := readUint8(ins, ip)
			frame.ip += 1
			vm.stack[frame.basePointer+localIndex] = vm.pop()

		case code.OpGetLocal:
			localIndex := readUint8(ins, ip)
			frame.ip += 1
			if err := vm.push(vm.stack[frame.basePointer+localIndex]); err != nil {
				return err
			}

		case code.OpArray:
			numElements := readUint16(ins, ip)
			frame.ip += 2
			array := vm.buildArray(vm.sp-numElements, vm.sp)
			vm.sp -= numElements
			if err := vm.push(array); err != nil {
				return err
			}

		case code.OpHash:
			numElements := readUint16(ins, ip)
			frame.ip += 2
			hash, err := vm.buildHash(vm.sp-numElements, vm.sp)
			if err != nil {
				return err
			}
			vm.sp -= numElements
			if err := vm.push(hash); err != nil {
				return err
			}

		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()

			if err := vm.executeIndexExpression(left, index); err != nil {
				return err
			}

		case code.OpSetIndex:
			value := vm.pop()
			index := vm.pop()
			left := vm.pop()

			if err := vm.executeSetIndexExpression(left, index, value); err != nil {
				return err
			}

		case code.OpPrint:
			value := vm.pop()
			output := objects.Stringify(value)
			if vm.onStep != nil {
				// Debug mode
				vm.output = append(vm.output, output)
			} else {
				// Normal mode - print to the configured writer
				fmt.Fprintln(vm.stdout, output)
			}

		case code.OpReturnValue:
			returnValue := vm.pop()

			poppedFrame := vm.popFrame()
			// Check if we're returning to the module's main frame
			if vm.framesIndex == 0 {
				// Module execution complete
				return nil
			}
			frame = vm.currentFrame()
			vm.sp = poppedFrame.basePointer - 1

			if err := vm.push(returnValue); err != nil {
				return err
			}
			ins = frame.cl.Fn.Instructions
			moduleGlobals = vm.modules[frame.cl.Fn.ModuleIdx].Globals

		case code.OpReturn:
			poppedFrame := vm.popFrame()
			// Check if we're returning to the module's main frame
			if vm.framesIndex == 0 {
				// Module execution complete
				return nil
			}
			frame = vm.currentFrame()
			vm.sp = poppedFrame.basePointer - 1

			if err := vm.push(objects.NilValue); err != nil {
				return err
			}
			ins = frame.cl.Fn.Instructions
			moduleGlobals = vm.modules[frame.cl.Fn.ModuleIdx].Globals

		case code.OpGetNative:
			nativeIndex := readUint8(ins, ip)
			frame.ip += 1

			nativeFn := objects.GetNativeFunctionByIndex(nativeIndex)
			if nativeFn == nil {
				return vm.runtimeError(fmt.Sprintf("native function index out of bounds: %d", nativeIndex))
			}

			if err := vm.push(nativeFn); err != nil {
				return err
			}

		case code.OpGetFree:
			freeIndex := readUint8(ins, ip)
			frame.ip += 1

			cell := frame.cl.Free[freeIndex]
			if err := vm.push(cell); err != nil {
				return err
			}

		case code.OpSetFree:
			freeIndex := readUint8(ins, ip)
			frame.ip += 1

			frame.cl.Free[freeIndex].Value = vm.pop()

		case code.OpGetCurrentClosure:
			if err := vm.push(frame.cl); err != nil {
				return err
			}

		case code.OpMakeCell:
			localIndex := readUint8(ins, ip)
			frame.ip += 1

			slotIdx := frame.basePointer + localIndex
			obj := vm.stack[slotIdx]

			// If already a Cell, just push it
			if cell, ok := obj.(*objects.Cell); ok {
				if err := vm.push(cell); err != nil {
					return err
				}
			} else {
				// Wrap in a new Cell, store it back, and push it
				cell := objects.NewCell(obj)
				vm.stack[slotIdx] = cell
				if err := vm.push(cell); err != nil {
					return err
				}
			}

		case code.OpGetClosure:
			constIndex := readUint16(ins, ip)
			numFree := int(ins[ip+3])
			frame.ip += 3

			if err := vm.executeClosure(constIndex, numFree); err != nil {
				return err
			}

		case code.OpCall:
			numArgs := readUint8(ins, ip)
			frame.ip += 1

			newFrame, err := vm.executeCall(numArgs)
			if err != nil {
				return err
			}
			if newFrame != nil {
				frame = newFrame
				ins = frame.cl.Fn.Instructions
				moduleGlobals = vm.modules[frame.cl.Fn.ModuleIdx].Globals
			}

		case code.OpClass:
			nameIdx := readUint16(ins, ip)
			numMethods := readUint8(ins, ip+2)
			numFields := readUint8(ins, ip+3)
			frame.ip += 4

			// Get class name from constants
			className := vm.constants[nameIdx].(*objects.String).Value

			// Pop methods from stack (they have names in CompiledFunction)
			methods := make(map[string]*objects.Closure)
			for i := 0; i < numMethods; i++ {
				closure := vm.stack[vm.sp-numMethods+i].(*objects.Closure)
				methods[closure.Fn.Name] = closure
			}
			vm.sp -= numMethods

			// Pop superclass (nil or CompiledClass)
			var superClass *objects.CompiledClass
			superObj := vm.pop()
			if _, ok := superObj.(*objects.Nil); !ok {
				var ok bool
				superClass, ok = superObj.(*objects.CompiledClass)
				if !ok {
					return vm.runtimeError(fmt.Sprintf("superclass must be a class, got %s", superObj.Type()))
				}
			}

			class := &objects.CompiledClass{
				Name:       className,
				Methods:    methods,
				SuperClass: superClass,
				NumFields:  numFields,
			}
			// Stamp the lexical class on each method so 'super' resolves
			// against the class the method was declared in.
			for _, method := range methods {
				method.DefiningClass = class
			}
			if err := vm.push(class); err != nil {
				return err
			}

		case code.OpGetField:
			slot := readUint8(ins, ip)
			frame.ip++

			// The checker resolved this to a declared field on a known class,
			// so there is nothing to look up and nothing that can be missing.
			instance := vm.pop().(*objects.CompiledInstance)
			if err := vm.push(instance.Fields[slot]); err != nil {
				return err
			}

		case code.OpSetField:
			slot := readUint8(ins, ip)
			frame.ip++

			value := vm.pop()
			instance := vm.pop().(*objects.CompiledInstance)
			instance.Fields[slot] = value
			if err := vm.push(value); err != nil {
				return err
			}

		case code.OpGetProperty:
			// Only method access reaches here; field access compiles to
			// OpGetField. Methods still resolve by name because they live on
			// the class rather than the instance.
			nameIdx := readUint16(ins, ip)
			frame.ip += 2

			name := vm.constants[nameIdx].(*objects.String).Value
			instance := vm.pop().(*objects.CompiledInstance)

			method, ok := instance.Class.LookupMethod(name)
			if !ok {
				return vm.runtimeError("instance does not have method " + name)
			}
			if err := vm.push(&objects.BoundMethod{Receiver: instance, Method: method}); err != nil {
				return err
			}

		case code.OpGetSuper:
			nameIdx := readUint16(ins, ip)
			frame.ip += 2

			name := vm.constants[nameIdx].(*objects.String).Value
			instance := vm.pop().(*objects.CompiledInstance)

			// 'super' is lexical: it refers to the superclass of the class the
			// currently executing method was declared in, not the receiver's
			// runtime class (which would make B.hello's super resolve to B
			// again when running on a C < B < A instance).
			definingClass := frame.cl.DefiningClass
			if definingClass == nil || definingClass.SuperClass == nil {
				return vm.runtimeError("cannot use 'super' here")
			}
			superClass := definingClass.SuperClass

			method, ok := superClass.LookupMethod(name)
			if !ok {
				return vm.runtimeError(fmt.Sprintf("undefined method '%s' in superclass %s",
					name, superClass.Name))
			}

			bound := &objects.BoundMethod{Receiver: instance, Method: method}
			if err := vm.push(bound); err != nil {
				return err
			}

		}
	}

	return nil
}

func (vm *VM) executeClosure(constIndex int, numFree int) error {
	fnObj := vm.constants[constIndex]
	function, ok := fnObj.(*objects.CompiledFunction)
	if !ok {
		return vm.runtimeError(fmt.Sprintf("expected compiled function, got %s", fnObj.Type()))
	}

	// Collect free variables from the stack
	free := make([]*objects.Cell, numFree)
	for i := 0; i < numFree; i++ {
		obj := vm.stack[vm.sp-numFree+i]
		if cell, ok := obj.(*objects.Cell); ok {
			// Reuse the existing cell
			free[i] = cell
		} else {
			// Wrap in a new cell (for local variables)
			free[i] = objects.NewCell(obj)
		}
	}
	vm.sp = vm.sp - numFree

	cl := objects.NewClosure(function, free)
	// A closure created inside a method keeps that method's lexical class,
	// so 'super' keeps working in nested functions.
	cl.DefiningClass = vm.currentFrame().cl.DefiningClass
	return vm.push(cl)
}

func (vm *VM) executeCall(numArgs int) (*Frame, error) {
	callee := unwrapCell(vm.stack[vm.sp-1-numArgs])

	switch fn := callee.(type) {
	case *objects.Closure:
		return vm.callClosure(fn, numArgs)
	case *objects.NativeFunction:
		return nil, vm.callNativeFunction(fn, numArgs)
	case *objects.CompiledClass:
		return vm.callClass(fn, numArgs)
	case *objects.BoundMethod:
		return vm.callBoundMethod(fn, numArgs)
	default:
		return nil, vm.runtimeError("Can only call functions or classes.")
	}
}

func (vm *VM) callNativeFunction(fn *objects.NativeFunction, numArgs int) error {
	// Arity is checked at compile time now. This guard stays as a backstop for
	// the same reason as the one in callValue: one integer comparison per
	// call, against a whole class of silent misbehaviour if it were wrong.
	if fn.NumArgs >= 0 && numArgs != fn.NumArgs {
		return vm.runtimeError(fmt.Sprintf("Expected %d arguments but got %d.", fn.NumArgs, numArgs))
	}

	// Unwrap any Cell arguments
	args := make([]objects.Object, numArgs)
	for i := 0; i < numArgs; i++ {
		args[i] = unwrapCell(vm.stack[vm.sp-numArgs+i])
	}

	result, err := fn.Fn(args...)
	if err != nil {
		// Native errors carry no location; anchor them to the call site
		return vm.runtimeError(err.Error())
	}

	vm.sp = vm.sp - numArgs - 1 // pop arguments and the function itself

	if result != nil {
		return vm.push(result)
	}
	return vm.push(objects.NilValue)
}

func (vm *VM) callClosure(cl *objects.Closure, numArgs int) (*Frame, error) {
	fn := cl.Fn
	if numArgs != fn.NumParameters {
		return nil, vm.runtimeError(fmt.Sprintf("Expected %d arguments but got %d.",
			fn.NumParameters, numArgs))
	}

	// Arguments are passed by value: unwrap any cell so that a captured
	// variable passed as an argument does not alias the caller's box (a
	// callee that captures its parameter would otherwise mutate the
	// caller's variable).
	base := vm.sp - numArgs
	for i := 0; i < numArgs; i++ {
		vm.stack[base+i] = unwrapCell(vm.stack[base+i])
	}

	frame := NewFrame(cl, base)
	if err := vm.pushFrame(frame); err != nil {
		return nil, err
	}
	vm.sp = frame.basePointer + fn.NumLocals

	return frame, nil
}

func (vm *VM) callClass(class *objects.CompiledClass, numArgs int) (*Frame, error) {
	instance := objects.NewCompiledInstance(class)

	// Check for init method (including inherited)
	if init, ok := class.LookupMethod("init"); ok {
		expectedArgs := init.Fn.NumParameters - 1 // -1 for 'this'
		if expectedArgs != numArgs {
			return nil, vm.runtimeError(fmt.Sprintf("Expected %d arguments but got %d.",
				expectedArgs, numArgs))
		}

		bound := objects.NewBoundMethod(instance, init)

		// Replace class on stack with bound method
		vm.stack[vm.sp-1-numArgs] = bound

		// Call init - it will return nil but we'll fix that below
		frame, err := vm.callBoundMethod(bound, numArgs)
		if err != nil {
			return nil, err
		}

		return frame, nil
	}

	// No init method - must have zero arguments
	if numArgs != 0 {
		return nil, vm.runtimeError(fmt.Sprintf("Expected 0 arguments but got %d.", numArgs))
	}

	// Replace class with instance on stack
	vm.sp = vm.sp - 1 // pop class
	if err := vm.push(instance); err != nil {
		return nil, err
	}
	return nil, nil
}

// callBoundMethod calls a method with its bound receiver.
func (vm *VM) callBoundMethod(bm *objects.BoundMethod, numArgs int) (*Frame, error) {
	expectedArgs := bm.Method.Fn.NumParameters - 1
	if expectedArgs != numArgs {
		return nil, vm.runtimeError(fmt.Sprintf("Expected %d arguments but got %d.",
			expectedArgs, numArgs))
	}

	// Shift everything (including bound_method slot) up by 1
	for i := numArgs; i >= 0; i-- {
		vm.stack[vm.sp-numArgs+i] = vm.stack[vm.sp-numArgs-1+i]
	}
	vm.sp++

	// Replace bound_method with this
	thisSlot := vm.sp - numArgs - 1
	vm.stack[thisSlot] = bm.Receiver

	// Arguments are passed by value: unwrap any caller cells (see callClosure)
	for i := 1; i <= numArgs; i++ {
		vm.stack[thisSlot+i] = unwrapCell(vm.stack[thisSlot+i])
	}

	// basePointer = thisSlot, so local 0 = this
	frame := NewFrame(bm.Method, thisSlot)
	if err := vm.pushFrame(frame); err != nil {
		return nil, err
	}
	vm.sp = frame.basePointer + bm.Method.Fn.NumLocals

	return frame, nil
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	if leftType == objects.TypeNumber && rightType == objects.TypeNumber {
		return vm.executeBinaryIntegerOperation(op, left, right)
	}

	if leftType == objects.TypeString && rightType == objects.TypeString {
		return vm.executeBinaryStringOperation(op, left, right)
	}

	// Mixed type addition: number + string or string + number
	if op == code.OpAdd {
		if leftType == objects.TypeNumber && rightType == objects.TypeString {
			return vm.executeBinaryNumberStringOperation(left, right)
		}
		if leftType == objects.TypeString && rightType == objects.TypeNumber {
			return vm.executeBinaryStringNumberOperation(left, right)
		}
	}

	return vm.runtimeError("Operands must be numbers.")
}

func (vm *VM) executeBinaryStringOperation(op code.Opcode, left, right objects.Object) error {
	if op != code.OpAdd {
		return vm.runtimeError(fmt.Sprintf("unknown string operator: %d", op))
	}

	leftVal := left.(*objects.String).Value
	rightVal := right.(*objects.String).Value

	return vm.push(&objects.String{Value: leftVal + rightVal})
}

func (vm *VM) executeBinaryNumberStringOperation(left, right objects.Object) error {
	leftVal := left.(*objects.Number)
	rightVal := right.(*objects.String)
	return vm.push(&objects.String{Value: leftVal.Inspect() + rightVal.Value})
}

func (vm *VM) executeBinaryStringNumberOperation(left, right objects.Object) error {
	leftVal := left.(*objects.String)
	rightVal := right.(*objects.Number)
	return vm.push(&objects.String{Value: leftVal.Value + rightVal.Inspect()})
}

func (vm *VM) executeBinaryIntegerOperation(op code.Opcode, left, right objects.Object) error {
	leftVal := left.(*objects.Number).Value
	rightVal := right.(*objects.Number).Value

	var result float64

	switch op {
	case code.OpAdd:
		result = leftVal + rightVal
	case code.OpSub:
		result = leftVal - rightVal
	case code.OpMul:
		result = leftVal * rightVal
	case code.OpDiv:
		if rightVal == 0 {
			return vm.runtimeError("Division by zero.")
		}
		result = leftVal / rightVal
	default:
		return vm.runtimeError(fmt.Sprintf("unknown integer operator: %d", op))
	}

	return vm.push(&objects.Number{Value: result})
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == objects.TypeNumber && right.Type() == objects.TypeNumber {
		return vm.executeIntegerComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(objects.NewBool(objects.IsEqual(left, right)))
	case code.OpNotEqual:
		return vm.push(objects.NewBool(!objects.IsEqual(left, right)))
	default:
		return vm.runtimeError("Operands must be numbers.")
	}
}

func (vm *VM) executeIntegerComparison(op code.Opcode, left, right objects.Object) error {
	leftVal := left.(*objects.Number).Value
	rightVal := right.(*objects.Number).Value

	switch op {
	case code.OpEqual:
		return vm.push(objects.NewBool(leftVal == rightVal))
	case code.OpNotEqual:
		return vm.push(objects.NewBool(leftVal != rightVal))
	case code.OpGreaterThan:
		return vm.push(objects.NewBool(leftVal > rightVal))
	case code.OpLess:
		return vm.push(objects.NewBool(leftVal < rightVal))
	default:
		return vm.runtimeError(fmt.Sprintf("unknown operator: %d", op))
	}
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()
	return vm.push(objects.NewBool(!objects.IsTruthy(operand)))
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	if operand.Type() != objects.TypeNumber {
		return vm.runtimeError(fmt.Sprintf("unsupported type for negation: %s", operand.Type()))
	}

	value := operand.(*objects.Number).Value
	return vm.push(&objects.Number{Value: -value})
}

func (vm *VM) buildArray(startIndex, endIndex int) objects.Object {
	elements := make([]objects.Object, endIndex-startIndex)

	for i := startIndex; i < endIndex; i++ {
		// Unwrap cells: a captured variable's value goes in the array, not its box
		elements[i-startIndex] = unwrapCell(vm.stack[i])
	}

	return &objects.Array{Elements: elements}
}

func (vm *VM) buildHash(startIndex, endIndex int) (objects.Object, error) {
	hash := objects.NewHash()

	for i := startIndex; i < endIndex; i += 2 {
		// Unwrap cells: a captured variable's value goes in the hash, not its box
		key := unwrapCell(vm.stack[i])
		value := unwrapCell(vm.stack[i+1])

		keyStr, err := vm.hashKey(key)
		if err != nil {
			return nil, err
		}

		hash.Set(keyStr, value)
	}

	return hash, nil
}

func (vm *VM) hashKey(key objects.Object) (string, error) {
	if k, ok := key.(*objects.String); ok {
		return k.Value, nil
	}
	return "", vm.runtimeError("Hash map keys must be strings.")
}

func (vm *VM) executeIndexExpression(left, index objects.Object) error {
	switch {
	case left.Type() == objects.TypeArray:
		return vm.executeArrayIndex(left, index)
	case left.Type() == objects.TypeHash:
		return vm.executeHashIndex(left, index)
	default:
		return vm.runtimeError("Indexing target must be an array or hash map.")
	}
}

// arrayIndex validates that index is an integral number within bounds.
func (vm *VM) arrayIndex(arrayObj *objects.Array, index objects.Object) (int, error) {
	num, ok := index.(*objects.Number)
	if !ok {
		return 0, vm.runtimeError("Index must be a number.")
	}
	idx := int(num.Value)
	if float64(idx) != num.Value {
		return 0, vm.runtimeError("Index must be an integer.")
	}
	if idx < 0 || idx >= len(arrayObj.Elements) {
		return 0, vm.runtimeError("index out of bounds")
	}
	return idx, nil
}

func (vm *VM) executeArrayIndex(array, index objects.Object) error {
	arrayObj := array.(*objects.Array)
	idx, err := vm.arrayIndex(arrayObj, index)
	if err != nil {
		return err
	}

	return vm.push(arrayObj.Elements[idx])
}

func (vm *VM) executeHashIndex(hash, index objects.Object) error {
	hashObj := hash.(*objects.Hash)

	key, err := vm.hashKey(index)
	if err != nil {
		return err
	}

	value, ok := hashObj.Get(key)
	if !ok {
		return vm.runtimeError(fmt.Sprintf("Key '%s' not found in hash map.", key))
	}

	return vm.push(value)
}

func (vm *VM) executeSetIndexExpression(left, index, value objects.Object) error {
	switch {
	case left.Type() == objects.TypeArray:
		return vm.executeArraySetIndex(left, index, value)
	case left.Type() == objects.TypeHash:
		return vm.executeHashSetIndex(left, index, value)
	default:
		return vm.runtimeError("Index assignment target must be an array or hash map.")
	}
}

func (vm *VM) executeArraySetIndex(array, index, value objects.Object) error {
	arrayObj := array.(*objects.Array)
	idx, err := vm.arrayIndex(arrayObj, index)
	if err != nil {
		return err
	}

	arrayObj.Elements[idx] = value
	return vm.push(value)
}

func (vm *VM) executeHashSetIndex(hash, index, value objects.Object) error {
	hashObj := hash.(*objects.Hash)

	key, err := vm.hashKey(index)
	if err != nil {
		return err
	}

	hashObj.Set(key, value)
	return vm.push(value)
}
