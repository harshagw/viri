// Package types is the compile-time analogue of internal/objects: where
// objects describes values that exist while a program runs, types describes
// what a program's expressions will produce before it runs.
//
// It is deliberately separate from ast.TypeExpr. A TypeExpr is syntax — the
// word the programmer wrote, once per occurrence. A Type is meaning: every
// `number` annotation in a program resolves to the same Number singleton,
// which is what makes comparison cheap.
package types

import "strings"

// Type is what an expression produces.
type Type interface {
	String() string
	typeNode()
}

// primitive is a type with no structure: number, string, bool, and void.
type primitive struct{ name string }

func (p *primitive) String() string { return p.name }
func (*primitive) typeNode()        {}

// The primitive types. These are singletons, so identity comparison is enough
// for them and Equal only needs to do real work on composite types.
var (
	Number = &primitive{"number"}
	String = &primitive{"string"}
	Bool   = &primitive{"bool"}

	// Void is the absence of a value, produced by a call to a function that
	// declares no return type. It is not a value and cannot be stored: no
	// variable, parameter, field, or collection element may have this type.
	// It is deliberately not called Nil — Viri has no nil, and naming this
	// Nil would invite it back into value positions.
	Void = &primitive{"void"}

	// Invalid is the result of an expression the checker could not type,
	// because something about it was already reported. It is assignable to
	// and from everything, which stops one mistake from cascading into a
	// series of unrelated complaints.
	Invalid = &primitive{"invalid"}
)

// Array is []T.
type Array struct{ Elem Type }

func (a *Array) String() string { return "[]" + a.Elem.String() }
func (*Array) typeNode()        {}

// Map is map[string]V. Keys are always strings: the VM's hash accepts nothing
// else, so the checker rejects any other key type rather than promising a map
// the runtime cannot provide.
type Map struct{ Value Type }

func (m *Map) String() string { return "map[string]" + m.Value.String() }
func (*Map) typeNode()        {}

// Function is a callable signature. Return is Void for a function that
// declares no return type.
type Function struct {
	Params []Type
	Return Type
}

func (f *Function) String() string {
	var b strings.Builder
	b.WriteString("fun(")
	for i, p := range f.Params {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(p.String())
	}
	b.WriteString(")")
	if f.Return != Void {
		b.WriteString(": ")
		b.WriteString(f.Return.String())
	}
	return b.String()
}
func (*Function) typeNode() {}

// Field is one declared class field. Fields are ordered because Phase 3
// assigns slot indices from that order, and a subclass layout has to extend
// its superclass's rather than reshuffle it.
type Field struct {
	Name string
	Type Type
}

// Class is a declared class. Classes are nominal: two classes with identical
// fields and methods are still different types, and assignability follows the
// declared Super chain rather than structure.
type Class struct {
	Name    string
	Super   *Class
	Fields  []Field
	Methods map[string]*Function
	// Init is the constructor signature, or nil when the class declares no
	// init. A class with no init of its own inherits its superclass's.
	Init *Function
}

func (c *Class) String() string { return c.Name }
func (*Class) typeNode()        {}

// LookupField finds a field by name, walking up the superclass chain. It
// mirrors objects.CompiledClass.LookupMethod, which does the same at runtime.
func (c *Class) LookupField(name string) (Field, bool) {
	for k := c; k != nil; k = k.Super {
		for _, f := range k.Fields {
			if f.Name == name {
				return f, true
			}
		}
	}
	return Field{}, false
}

// LookupMethod finds a method by name, walking up the superclass chain.
func (c *Class) LookupMethod(name string) (*Function, bool) {
	for k := c; k != nil; k = k.Super {
		if m, ok := k.Methods[name]; ok {
			return m, true
		}
	}
	return nil, false
}

// Constructor returns the signature used to build an instance: the class's own
// init, or the nearest one it inherits. A class with none takes no arguments.
func (c *Class) Constructor() *Function {
	for k := c; k != nil; k = k.Super {
		if k.Init != nil {
			return k.Init
		}
	}
	return &Function{Params: nil, Return: c}
}

// AllFields returns every field visible on the class, superclass fields first.
// The order is the layout Phase 3 will assign slots from.
func (c *Class) AllFields() []Field {
	if c.Super == nil {
		return append([]Field(nil), c.Fields...)
	}
	return append(c.Super.AllFields(), c.Fields...)
}

// Module is what another module exports.
//
// A module has two namespaces, because an exported class is both. In value
// position `shapes.Circle` is a constructor — a function returning an
// instance — while in type position it is the class itself. Keeping them
// apart is what stops `var c: shapes.Circle = shapes.Circle;` from checking.
type Module struct {
	Path string
	// Exports is the value namespace: what `alias.name` evaluates to.
	Exports map[string]Type
	// Classes is the type namespace: what `alias.Name` means in an
	// annotation.
	Classes map[string]*Class
}

func (m *Module) String() string { return "module " + m.Path }
func (*Module) typeNode()        {}
