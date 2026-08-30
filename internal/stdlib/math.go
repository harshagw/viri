package stdlib

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/types"
)

// std:math standard library module.
var MathModule = NewNativeModule("std:math", mathExports, mathSignatures)

var mathExports = map[string]objects.Object{
	"PI":     objects.NewNumber(math.Pi),
	"E":      objects.NewNumber(math.E),
	"abs":    &objects.NativeFunction{Name: "abs", NumArgs: 1, Fn: mathAbs},
	"sqrt":   &objects.NativeFunction{Name: "sqrt", NumArgs: 1, Fn: mathSqrt},
	"pow":    &objects.NativeFunction{Name: "pow", NumArgs: 2, Fn: mathPow},
	"floor":  &objects.NativeFunction{Name: "floor", NumArgs: 1, Fn: mathFloor},
	"ceil":   &objects.NativeFunction{Name: "ceil", NumArgs: 1, Fn: mathCeil},
	"round":  &objects.NativeFunction{Name: "round", NumArgs: 1, Fn: mathRound},
	"sin":    &objects.NativeFunction{Name: "sin", NumArgs: 1, Fn: mathSin},
	"cos":    &objects.NativeFunction{Name: "cos", NumArgs: 1, Fn: mathCos},
	"tan":    &objects.NativeFunction{Name: "tan", NumArgs: 1, Fn: mathTan},
	"log":    &objects.NativeFunction{Name: "log", NumArgs: 1, Fn: mathLog},
	"exp":    &objects.NativeFunction{Name: "exp", NumArgs: 1, Fn: mathExp},
	"min":    &objects.NativeFunction{Name: "min", NumArgs: 2, Fn: mathMin},
	"max":    &objects.NativeFunction{Name: "max", NumArgs: 2, Fn: mathMax},
	"random": &objects.NativeFunction{Name: "random", NumArgs: 0, Fn: mathRandom},
}

// mathSignatures is the compile-time view of the module. Every export needs an
// entry: the checker resolves `math.sqrt` through this, and a missing name
// reads as "not exported".
var mathSignatures = map[string]types.Type{
	"PI": types.Number,
	"E":  types.Number,

	"abs":   unaryNumber,
	"sqrt":  unaryNumber,
	"floor": unaryNumber,
	"ceil":  unaryNumber,
	"round": unaryNumber,
	"sin":   unaryNumber,
	"cos":   unaryNumber,
	"tan":   unaryNumber,
	"log":   unaryNumber,
	"exp":   unaryNumber,

	"pow": binaryNumber,
	"min": binaryNumber,
	"max": binaryNumber,

	"random": &types.Function{Params: nil, Return: types.Number},
}

var (
	unaryNumber  = &types.Function{Params: []types.Type{types.Number}, Return: types.Number}
	binaryNumber = &types.Function{Params: []types.Type{types.Number, types.Number}, Return: types.Number}
)

// Helper to extract a number from an argument
func requireNumber(args []objects.Object, index int, fnName string) (float64, error) {
	if index >= len(args) {
		return 0, fmt.Errorf("%s: missing argument %d", fnName, index+1)
	}
	num, ok := args[index].(*objects.Number)
	if !ok {
		return 0, fmt.Errorf("%s: argument %d must be a number, got %s", fnName, index+1, args[index].Type())
	}
	return num.Value, nil
}

func mathAbs(args ...objects.Object) (objects.Object, error) {
	n, err := requireNumber(args, 0, "abs")
	if err != nil {
		return nil, err
	}
	return objects.NewNumber(math.Abs(n)), nil
}

func mathSqrt(args ...objects.Object) (objects.Object, error) {
	n, err := requireNumber(args, 0, "sqrt")
	if err != nil {
		return nil, err
	}
	if n < 0 {
		return nil, fmt.Errorf("sqrt: cannot take square root of negative number")
	}
	return objects.NewNumber(math.Sqrt(n)), nil
}

func mathPow(args ...objects.Object) (objects.Object, error) {
	base, err := requireNumber(args, 0, "pow")
	if err != nil {
		return nil, err
	}
	exp, err := requireNumber(args, 1, "pow")
	if err != nil {
		return nil, err
	}
	return objects.NewNumber(math.Pow(base, exp)), nil
}

func mathFloor(args ...objects.Object) (objects.Object, error) {
	n, err := requireNumber(args, 0, "floor")
	if err != nil {
		return nil, err
	}
	return objects.NewNumber(math.Floor(n)), nil
}

func mathCeil(args ...objects.Object) (objects.Object, error) {
	n, err := requireNumber(args, 0, "ceil")
	if err != nil {
		return nil, err
	}
	return objects.NewNumber(math.Ceil(n)), nil
}

func mathRound(args ...objects.Object) (objects.Object, error) {
	n, err := requireNumber(args, 0, "round")
	if err != nil {
		return nil, err
	}
	return objects.NewNumber(math.Round(n)), nil
}

func mathSin(args ...objects.Object) (objects.Object, error) {
	n, err := requireNumber(args, 0, "sin")
	if err != nil {
		return nil, err
	}
	return objects.NewNumber(math.Sin(n)), nil
}

func mathCos(args ...objects.Object) (objects.Object, error) {
	n, err := requireNumber(args, 0, "cos")
	if err != nil {
		return nil, err
	}
	return objects.NewNumber(math.Cos(n)), nil
}

func mathTan(args ...objects.Object) (objects.Object, error) {
	n, err := requireNumber(args, 0, "tan")
	if err != nil {
		return nil, err
	}
	return objects.NewNumber(math.Tan(n)), nil
}

func mathLog(args ...objects.Object) (objects.Object, error) {
	n, err := requireNumber(args, 0, "log")
	if err != nil {
		return nil, err
	}
	if n <= 0 {
		return nil, fmt.Errorf("log: argument must be positive")
	}
	return objects.NewNumber(math.Log(n)), nil
}

func mathExp(args ...objects.Object) (objects.Object, error) {
	n, err := requireNumber(args, 0, "exp")
	if err != nil {
		return nil, err
	}
	return objects.NewNumber(math.Exp(n)), nil
}

func mathMin(args ...objects.Object) (objects.Object, error) {
	a, err := requireNumber(args, 0, "min")
	if err != nil {
		return nil, err
	}
	b, err := requireNumber(args, 1, "min")
	if err != nil {
		return nil, err
	}
	return objects.NewNumber(math.Min(a, b)), nil
}

func mathMax(args ...objects.Object) (objects.Object, error) {
	a, err := requireNumber(args, 0, "max")
	if err != nil {
		return nil, err
	}
	b, err := requireNumber(args, 1, "max")
	if err != nil {
		return nil, err
	}
	return objects.NewNumber(math.Max(a, b)), nil
}

func mathRandom(args ...objects.Object) (objects.Object, error) {
	return objects.NewNumber(rand.Float64()), nil
}
