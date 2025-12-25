package objects

import (
	"errors"
	"fmt"
	"time"
)

type NativeFunctionFn func(args ...Object) (Object, error)

type NativeFunction struct {
	Name    string
	NumArgs int // -1 means variadic
	Fn      NativeFunctionFn
}

func (n *NativeFunction) Type() Type      { return TypeNativeFun }
func (n *NativeFunction) Inspect() string { return fmt.Sprintf("<native_fun %s>", n.Name) }

func (n *NativeFunction) Call(exec BlockExecutor, arguments []Object) (Object, error) {
	return n.Fn(arguments...)
}

func (n *NativeFunction) Arity() int {
	return n.NumArgs
}

func (n *NativeFunction) String() string {
	return n.Inspect()
}

var NativeFunctions = []*NativeFunction{
	{Name: "clock", NumArgs: 0, Fn: nativeClock},
	{Name: "len", NumArgs: 1, Fn: nativeLen},
}

func GetNativeFunctionByIndex(index int) *NativeFunction {
	if index < 0 || index >= len(NativeFunctions) {
		return nil
	}
	return NativeFunctions[index]
}

func nativeClock(args ...Object) (Object, error) {
	return NewNumber(float64(time.Now().Unix())), nil
}

func nativeLen(args ...Object) (Object, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("wrong number of arguments. got=%d, want=1", len(args))
	}

	value := args[0]
	switch value.Type() {
	case TypeString:
		return NewNumber(float64(len(value.(*String).Value))), nil
	case TypeArray:
		if arr, ok := value.(*Array); ok {
			return NewNumber(float64(len(arr.Elements))), nil
		}
	case TypeHash:
		if hash, ok := value.(*Hash); ok {
			return NewNumber(float64(len(hash.Pairs))), nil
		}
	}
	return nil, errors.New("invalid argument type for len function: " + string(value.Type()))
}
