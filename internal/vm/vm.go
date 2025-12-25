package vm

import (
	"fmt"

	"github.com/harshagw/viri/internal/code"
	"github.com/harshagw/viri/internal/compiler"
	"github.com/harshagw/viri/internal/objects"
)

const StackSize = 2048
const GlobalsSize = 65536

type VM struct {
	constants    []objects.Object
	instructions code.Instructions

	stack []objects.Object
	sp    int // Always points to the next value. Top of stack is stack[sp-1]

	globals []objects.Object
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,

		stack:   make([]objects.Object, StackSize),
		sp:      0,
		globals: make([]objects.Object, GlobalsSize),
	}
}

func (vm *VM) StackTop() objects.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])

		switch op {
		case code.OpConstant:
			def, _ := code.Lookup(byte(op))
			operands, read := code.ReadOperands(def, vm.instructions[ip+1:])
			ip += read
			if err := vm.push(vm.constants[operands[0]]); err != nil {
				return err
			}

		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			if err := vm.executeBinaryOperation(op); err != nil {
				return err
			}

		case code.OpTrue:
			if err := vm.push(objects.NewBool(true)); err != nil {
				return err
			}

		case code.OpFalse:
			if err := vm.push(objects.NewBool(false)); err != nil {
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
			def, _ := code.Lookup(byte(op))
			pos, _ := code.ReadOperands(def, vm.instructions[ip+1:])
			ip = pos[0] - 1 // -1 because the loop will increment ip

		case code.OpJumpNotTruthy:
			def, _ := code.Lookup(byte(op))
			pos, read := code.ReadOperands(def, vm.instructions[ip+1:])
			ip += read

			condition := vm.pop()
			if !objects.IsTruthy(condition) {
				ip = pos[0] - 1 // -1 because the loop will increment ip
			}

		case code.OpPop:
			vm.pop()

		case code.OpDup:
			if err := vm.push(vm.stack[vm.sp-1]); err != nil {
				return err
			}

		case code.OpSetGlobal:
			def, _ := code.Lookup(byte(op))
			operands, read := code.ReadOperands(def, vm.instructions[ip+1:])
			ip += read

			globalIndex := operands[0]
			vm.globals[globalIndex] = vm.pop()

		case code.OpGetGlobal:
			def, _ := code.Lookup(byte(op))
			operands, read := code.ReadOperands(def, vm.instructions[ip+1:])
			ip += read

			if err := vm.push(vm.globals[operands[0]]); err != nil {
				return err
			}

		case code.OpArray:
			def, _ := code.Lookup(byte(op))
			operands, read := code.ReadOperands(def, vm.instructions[ip+1:])
			ip += read

			numElements := operands[0]
			array := vm.buildArray(vm.sp-numElements, vm.sp)
			vm.sp = vm.sp - numElements

			if err := vm.push(array); err != nil {
				return err
			}

		case code.OpHash:
			def, _ := code.Lookup(byte(op))
			operands, read := code.ReadOperands(def, vm.instructions[ip+1:])
			ip += read

			numElements := operands[0]
			hash, err := vm.buildHash(vm.sp-numElements, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - numElements

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
			fmt.Println(objects.Stringify(value))
		}
	}

	return nil
}

func (vm *VM) push(o objects.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

func (vm *VM) pop() objects.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

func (vm *VM) LastPoppedStackElem() objects.Object {
	return vm.stack[vm.sp]
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
