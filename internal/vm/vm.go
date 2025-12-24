package vm

import (
	"fmt"

	"github.com/harshagw/viri/internal/code"
	"github.com/harshagw/viri/internal/compiler"
	"github.com/harshagw/viri/internal/objects"
)

const StackSize = 2048

type VM struct {
	constants    []objects.Object
	instructions code.Instructions

	stack []objects.Object
	sp    int // Always points to the next value. Top of stack is stack[sp-1]
}

func New(bytecode *compiler.Bytecode) *VM {
	return &VM{
		instructions: bytecode.Instructions,
		constants:    bytecode.Constants,

		stack: make([]objects.Object, StackSize),
		sp:    0,
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
			constIndex, _ := code.ReadOperands(&code.Definition{OperandWidths: []int{2}}, vm.instructions[ip+1:])
			ip += 2
			if err := vm.push(vm.constants[constIndex[0]]); err != nil {
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

		case code.OpPop:
			vm.pop()
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

	return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
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
