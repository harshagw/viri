package objects

// Type describes the concrete runtime type of a value.
type Type string

const (
	TypeNumber    Type = "NUMBER"
	TypeString    Type = "STRING"
	TypeBool      Type = "BOOL"
	TypeNil       Type = "NIL"
	TypeFunction  Type = "FUNCTION"
	TypeClass     Type = "CLASS"
	TypeInstance  Type = "INSTANCE"
	TypeNativeFun Type = "NATIVE_FUNCTION"
	TypeArray Type = "ARRAY"
)

// Object is a runtime value.
type Object interface {
	Type() Type
	Inspect() string
}

// IsTruthy implements language truthiness rules.
func IsTruthy(v Object) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case *Bool:
		return val.Value
	case *Nil:
		return false
	default:
		return true
	}
}

// IsEqual implements language equality semantics.
func IsEqual(a, b Object) bool {
	switch av := a.(type) {
	case *Nil:
		_, isNil := b.(*Nil)
		return isNil
	case *Number:
		bv, ok := b.(*Number)
		return ok && av.Value == bv.Value
	case *String:
		bv, ok := b.(*String)
		return ok && av.Value == bv.Value
	case *Bool:
		bv, ok := b.(*Bool)
		return ok && av.Value == bv.Value
	default:
		return a == b
	}
}

// Stringify renders a value to the user.
func Stringify(v Object) string {
	if v == nil {
		return "nil"
	}
	return v.Inspect()
}
