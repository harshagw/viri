package vm

import (
	"github.com/harshagw/viri/internal/code"
	"github.com/harshagw/viri/internal/objects"
)

// VMState is a snapshot of VM state for debugging
type VMState struct {
	// Instruction info
	IP           int
	OpCode       code.Opcode
	OpName       string
	Operands     []int
	Instructions code.Instructions

	// Stack info
	SP    int
	Stack []objects.Object

	// Frame info
	FrameIndex int
	Frames     []FrameInfo

	// Module info
	CurrentModule int
	NumModules    int
	ModuleGlobals [][]objects.Object // globals for each module

	// Other
	Constants []objects.Object
	Globals   []objects.Object // current module's globals (for convenience)
	Output    []string
}

// FrameInfo is a snapshot of a call frame
type FrameInfo struct {
	IP              int
	BasePointer     int
	InstructionSize int
	ClosureInfo     ClosureInfo
}

// ClosureInfo contains closure details for display
type ClosureInfo struct {
	NumLocals     int
	NumParameters int
	NumFree       int
	FreeVars      []objects.Object
}

// GetState returns a snapshot of current VM state
func (vm *VM) GetState() *VMState {
	frame := vm.currentFrame()
	ins := frame.cl.Fn.Instructions
	ip := frame.ip

	// Get current opcode info
	var opCode code.Opcode
	var opName string
	var operands []int

	if ip >= 0 && ip < len(ins) {
		opCode = code.Opcode(ins[ip])
		if def, err := code.Lookup(byte(opCode)); err == nil {
			opName = def.Name
			operands, _ = code.ReadOperands(def, ins[ip+1:])
		}
	}

	// Copy stack (only active portion)
	stack := make([]objects.Object, vm.sp)
	copy(stack, vm.stack[:vm.sp])

	// Build frame info
	frames := make([]FrameInfo, vm.framesIndex)
	for i := 0; i < vm.framesIndex; i++ {
		f := vm.frames[i]
		freeVars := make([]objects.Object, len(f.cl.Free))
		for j, cell := range f.cl.Free {
			freeVars[j] = cell.Value
		}
		frames[i] = FrameInfo{
			IP:              f.ip,
			BasePointer:     f.basePointer,
			InstructionSize: len(f.cl.Fn.Instructions),
			ClosureInfo: ClosureInfo{
				NumLocals:     f.cl.Fn.NumLocals,
				NumParameters: f.cl.Fn.NumParameters,
				NumFree:       len(f.cl.Free),
				FreeVars:      freeVars,
			},
		}
	}

	// Collect globals from all modules
	moduleGlobals := make([][]objects.Object, len(vm.modules))
	for i, mod := range vm.modules {
		globals := make([]objects.Object, len(mod.Globals))
		copy(globals, mod.Globals)
		moduleGlobals[i] = globals
	}

	// Current module's globals for convenience
	var currentGlobals []objects.Object
	if vm.currentModule < len(vm.modules) {
		currentGlobals = moduleGlobals[vm.currentModule]
	}

	return &VMState{
		IP:            ip,
		OpCode:        opCode,
		OpName:        opName,
		Operands:      operands,
		Instructions:  ins,
		SP:            vm.sp,
		Stack:         stack,
		FrameIndex:    vm.framesIndex,
		Frames:        frames,
		CurrentModule: vm.currentModule,
		NumModules:    len(vm.modules),
		ModuleGlobals: moduleGlobals,
		Constants:     vm.constants,
		Globals:       currentGlobals,
		Output:        vm.output,
	}
}
