package types

// Equal reports whether two types are the same type.
//
// Primitives are singletons, so identity settles them. Classes are nominal:
// two classes are equal only when they are the same declaration, never because
// they happen to have the same shape. Everything else compares structurally.
func Equal(a, b Type) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch x := a.(type) {
	case *Array:
		y, ok := b.(*Array)
		return ok && Equal(x.Elem, y.Elem)

	case *Map:
		y, ok := b.(*Map)
		return ok && Equal(x.Value, y.Value)

	case *Function:
		y, ok := b.(*Function)
		if !ok || len(x.Params) != len(y.Params) {
			return false
		}
		for i := range x.Params {
			if !Equal(x.Params[i], y.Params[i]) {
				return false
			}
		}
		return Equal(x.Return, y.Return)
	}

	// Primitives and classes are settled by identity, checked above.
	return false
}

// AssignableTo reports whether a value of type src may be stored where dst is
// expected.
//
// With no nilable types there is no third case to worry about: it is exact
// equality, plus a walk up the superclass chain so a Square may be used as a
// Shape. Arrays, maps, and functions are invariant — []Square is not
// assignable to []Shape, because the array could then be written through with
// a plain Shape.
func AssignableTo(src, dst Type) bool {
	if src == nil || dst == nil {
		return false
	}
	// Invalid already produced a diagnostic; let it pass so one mistake does
	// not cascade.
	if src == Invalid || dst == Invalid {
		return true
	}
	// Void is not a value and cannot be stored anywhere.
	if src == Void || dst == Void {
		return false
	}
	if Equal(src, dst) {
		return true
	}

	if sub, ok := src.(*Class); ok {
		if super, ok := dst.(*Class); ok {
			for k := sub.Super; k != nil; k = k.Super {
				if k == super {
					return true
				}
			}
		}
	}
	return false
}

// Comparable reports whether two types may be compared with == or !=.
//
// The operands must be the same type, or two classes on one inheritance chain
// — comparing a string to a number is a mistake, not a false result.
func Comparable(a, b Type) bool {
	if a == Invalid || b == Invalid {
		return true
	}
	if a == Void || b == Void {
		return false
	}
	return AssignableTo(a, b) || AssignableTo(b, a)
}

// IsStorable reports whether a value of this type may be bound to a name,
// passed as an argument, or held in a collection. Only Void cannot.
func IsStorable(t Type) bool {
	return t != nil && t != Void
}
