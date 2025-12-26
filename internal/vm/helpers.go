package vm

import (
	"github.com/harshagw/viri/internal/code"
	"github.com/harshagw/viri/internal/objects"
)

// unwrapCell returns the value inside a Cell, or the object itself if not a Cell
func unwrapCell(o objects.Object) objects.Object {
	if cell, ok := o.(*objects.Cell); ok {
		return cell.Value
	}
	return o
}

// readUint16 reads a 2-byte big-endian operand at ip+1
func readUint16(ins code.Instructions, ip int) int {
	return int(ins[ip+1])<<8 | int(ins[ip+2])
}

// readUint8 reads a 1-byte operand at ip+1
func readUint8(ins code.Instructions, ip int) int {
	return int(ins[ip+1])
}
