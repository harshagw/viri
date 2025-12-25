package vm

import (
	"github.com/harshagw/viri/internal/code"
	"github.com/harshagw/viri/internal/objects"
)

// Frame represents a call frame for function execution
type Frame struct {
	cl          *objects.Closure
	ip          int // instruction pointer within this frame
	basePointer int // points to the bottom of the stack for this frame
}

func NewFrame(cl *objects.Closure, basePointer int) *Frame {
	return &Frame{
		cl:          cl,
		ip:          -1,
		basePointer: basePointer,
	}
}

func (f *Frame) Instructions() code.Instructions {
	return f.cl.Fn.Instructions
}
