package vm

import (
	"fmt"

	"github.com/harshagw/viri/internal/code"
	"github.com/harshagw/viri/internal/compiler"
	"github.com/harshagw/viri/internal/objects"
)

const StackSize = 2048
const GlobalsSize = 65536
const MaxFrames = 1024

type VM struct {
	constants []objects.Object

	stack []objects.Object
	sp    int // Always points to the next value. Top of stack is stack[sp-1]

	globals []objects.Object

	frames      []*Frame
	framesIndex int // Always points to the next frame to be used. Top of frame is frames[framesIndex-1]

	onStep func()   // Debug callback, called before each opcode execution
	output []string // Capture print output
}

func New(bytecode *compiler.Bytecode) *VM {
	// Wrap the main program in a compiled function
	mainFn := &objects.CompiledFunction{Instructions: bytecode.Instructions}
	mainFrame := NewFrame(objects.NewClosure(mainFn, nil), 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants: bytecode.Constants,

		stack:   make([]objects.Object, StackSize),
		sp:      0,
		globals: make([]objects.Object, GlobalsSize),

		frames:      frames,
		framesIndex: 1,
	}
}

func (vm *VM) UpdateGlobals(globals []objects.Object) {
	vm.globals = globals
}

func (vm *VM) SetOnStep(fn func()) {
	vm.onStep = fn
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

func (vm *VM) StackTop() objects.Object {
	if vm.sp == 0 {
		return nil
	}
	return unwrapCell(vm.stack[vm.sp-1])
}

// readUint16 reads a 2-byte big-endian operand at ip+1
func readUint16(ins code.Instructions, ip int) int {
	return int(ins[ip+1])<<8 | int(ins[ip+2])
}

// readUint8 reads a 1-byte operand at ip+1
func readUint8(ins code.Instructions, ip int) int {
	return int(ins[ip+1])
}

func (vm *VM) Run() error {
	var ip int
	var ins code.Instructions
	var op code.Opcode
	var frame *Frame

	frame = vm.currentFrame()
	ins = frame.cl.Fn.Instructions

	for frame.ip < len(ins)-1 {
		frame.ip++
		ip = frame.ip
		op = code.Opcode(ins[ip])

		if vm.onStep != nil {
			vm.onStep()
		}

		switch op {
		case code.OpConstant:
			constIndex := readUint16(ins, ip)
			frame.ip += 2
			if err := vm.push(vm.constants[constIndex]); err != nil {
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

		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
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
			condition := vm.pop()
			if !objects.IsTruthy(condition) {
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
			vm.globals[globalIndex] = vm.pop()

		case code.OpGetGlobal:
			globalIndex := readUint16(ins, ip)
			frame.ip += 2
			if err := vm.push(vm.globals[globalIndex]); err != nil {
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
				// Normal mode - print to stdout
				fmt.Println(output)
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
			}

		case code.OpReturnValue:
			returnValue := vm.pop()

			poppedFrame := vm.popFrame()
			frame = vm.currentFrame()
			vm.sp = poppedFrame.basePointer - 1

			if err := vm.push(returnValue); err != nil {
				return err
			}
			ins = frame.cl.Fn.Instructions

		case code.OpReturn:
			poppedFrame := vm.popFrame()
			frame = vm.currentFrame()
			vm.sp = poppedFrame.basePointer - 1

			if err := vm.push(objects.NilValue); err != nil {
				return err
			}
			ins = frame.cl.Fn.Instructions

		case code.OpClosure:
			constIndex := readUint16(ins, ip)
			numFree := int(ins[ip+3])
			frame.ip += 3

			if err := vm.pushClosure(constIndex, numFree); err != nil {
				return err
			}

		case code.OpGetNative:
			nativeIndex := readUint8(ins, ip)
			frame.ip += 1

			nativeFn := objects.GetNativeFunctionByIndex(nativeIndex)
			if nativeFn == nil {
				return fmt.Errorf("native function index out of bounds: %d", nativeIndex)
			}

			if err := vm.push(nativeFn); err != nil {
				return err
			}

		case code.OpGetFree:
			freeIndex := readUint8(ins, ip)
			frame.ip += 1

			currentClosure := frame.cl
			cell := currentClosure.Free[freeIndex]
			if err := vm.push(cell); err != nil {
				return err
			}

		case code.OpSetFree:
			freeIndex := readUint8(ins, ip)
			frame.ip += 1

			currentClosure := frame.cl
			currentClosure.Free[freeIndex].Value = vm.pop()

		case code.OpCurrentClosure:
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
		}
	}

	return nil
}

func (vm *VM) pushClosure(constIndex int, numFree int) error {
	fnObj := vm.constants[constIndex]
	function, ok := fnObj.(*objects.CompiledFunction)
	if !ok {
		return fmt.Errorf("expected compiled function, got %s", fnObj.Type())
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
	return vm.push(cl)
}

func (vm *VM) executeCall(numArgs int) (*Frame, error) {
	callee := unwrapCell(vm.stack[vm.sp-1-numArgs])

	switch fn := callee.(type) {
	case *objects.Closure:
		return vm.callClosure(fn, numArgs)
	case *objects.NativeFunction:
		return nil, vm.callNativeFunction(fn, numArgs)
	default:
		return nil, fmt.Errorf("calling non-function: %s", callee.Type())
	}
}

func (vm *VM) callNativeFunction(fn *objects.NativeFunction, numArgs int) error {
	// Unwrap any Cell arguments
	args := make([]objects.Object, numArgs)
	for i := 0; i < numArgs; i++ {
		args[i] = unwrapCell(vm.stack[vm.sp-numArgs+i])
	}

	result, err := fn.Fn(args...)
	if err != nil {
		return err
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
		return nil, fmt.Errorf("wrong number of arguments: want=%d, got=%d",
			fn.NumParameters, numArgs)
	}

	frame := NewFrame(cl, vm.sp-numArgs)
	vm.pushFrame(frame)
	vm.sp = frame.basePointer + fn.NumLocals

	return frame, nil
}

func (vm *VM) push(o objects.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

// unwrapCell returns the value inside a Cell, or the object itself if not a Cell
func unwrapCell(o objects.Object) objects.Object {
	if cell, ok := o.(*objects.Cell); ok {
		return cell.Value
	}
	return o
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

	return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
}

func (vm *VM) executeBinaryStringOperation(op code.Opcode, left, right objects.Object) error {
	if op != code.OpAdd {
		return fmt.Errorf("unknown string operator: %d", op)
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
		result = leftVal / rightVal
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
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
		return fmt.Errorf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())
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
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()
	return vm.push(objects.NewBool(!objects.IsTruthy(operand)))
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	if operand.Type() != objects.TypeNumber {
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}

	value := operand.(*objects.Number).Value
	return vm.push(&objects.Number{Value: -value})
}

func (vm *VM) buildArray(startIndex, endIndex int) objects.Object {
	elements := make([]objects.Object, endIndex-startIndex)

	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]
	}

	return &objects.Array{Elements: elements}
}

func (vm *VM) buildHash(startIndex, endIndex int) (objects.Object, error) {
	hash := objects.NewHash()

	for i := startIndex; i < endIndex; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]

		keyStr, err := vm.hashKey(key)
		if err != nil {
			return nil, err
		}

		hash.Set(keyStr, value)
	}

	return hash, nil
}

func (vm *VM) hashKey(key objects.Object) (string, error) {
	switch k := key.(type) {
	case *objects.String:
		return k.Value, nil
	case *objects.Number:
		return k.Inspect(), nil
	case *objects.Bool:
		return k.Inspect(), nil
	default:
		return "", fmt.Errorf("unusable as hash key: %s", key.Type())
	}
}

func (vm *VM) executeIndexExpression(left, index objects.Object) error {
	switch {
	case left.Type() == objects.TypeArray && index.Type() == objects.TypeNumber:
		return vm.executeArrayIndex(left, index)
	case left.Type() == objects.TypeHash:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s[%s]", left.Type(), index.Type())
	}
}

func (vm *VM) executeArrayIndex(array, index objects.Object) error {
	arrayObj := array.(*objects.Array)
	idx := int(index.(*objects.Number).Value)

	if idx < 0 || idx >= len(arrayObj.Elements) {
		return fmt.Errorf("index out of bounds")
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
		return fmt.Errorf("key '%s' not found in hash map", key)
	}

	return vm.push(value)
}

func (vm *VM) executeSetIndexExpression(left, index, value objects.Object) error {
	switch {
	case left.Type() == objects.TypeArray && index.Type() == objects.TypeNumber:
		return vm.executeArraySetIndex(left, index, value)
	case left.Type() == objects.TypeHash:
		return vm.executeHashSetIndex(left, index, value)
	default:
		return fmt.Errorf("index assignment not supported: %s[%s]", left.Type(), index.Type())
	}
}

func (vm *VM) executeArraySetIndex(array, index, value objects.Object) error {
	arrayObj := array.(*objects.Array)
	idx := int(index.(*objects.Number).Value)

	if idx < 0 || idx >= len(arrayObj.Elements) {
		return fmt.Errorf("index out of bounds: %d", idx)
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
