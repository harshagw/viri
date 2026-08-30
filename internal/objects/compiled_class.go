package objects

import "fmt"

// CompiledClass represents a class in the VM.
type CompiledClass struct {
	Name       string
	Methods    map[string]*Closure // Method name -> closure
	SuperClass *CompiledClass      // nil if no superclass
	// NumFields is the size of an instance's field storage. The checker
	// computed the layout; the VM only needs its size, because every access
	// is already resolved to a slot at compile time.
	NumFields int
}

func (c *CompiledClass) Type() Type {
	return TypeCompiledClass
}

func (c *CompiledClass) Inspect() string {
	return fmt.Sprintf("<class %s>", c.Name)
}

// LookupMethod finds a method in the class hierarchy.
func (c *CompiledClass) LookupMethod(name string) (*Closure, bool) {
	if method, ok := c.Methods[name]; ok {
		return method, true
	}
	if c.SuperClass != nil {
		return c.SuperClass.LookupMethod(name)
	}
	return nil, false
}

// CompiledInstance represents an instance of a CompiledClass.
//
// Fields is indexed, not named: the checker proved every access refers to a
// declared field and resolved it to a slot, so the runtime needs no hash
// lookup and no "undefined property" case. Superclass fields occupy the first
// slots, so a subclass instance can be read through a superclass type without
// remapping.
type CompiledInstance struct {
	Class  *CompiledClass
	Fields []Object
}

func NewCompiledInstance(class *CompiledClass) *CompiledInstance {
	return &CompiledInstance{
		Class:  class,
		Fields: make([]Object, class.NumFields),
	}
}

func (i *CompiledInstance) Type() Type {
	return TypeCompiledInstance
}

func (i *CompiledInstance) Inspect() string {
	return fmt.Sprintf("<instance %s>", i.Class.Name)
}

// BoundMethod wraps a closure with its receiver instance.
// When called, the receiver becomes 'this'.
type BoundMethod struct {
	Receiver *CompiledInstance
	Method   *Closure
}

func NewBoundMethod(receiver *CompiledInstance, method *Closure) *BoundMethod {
	return &BoundMethod{
		Receiver: receiver,
		Method:   method,
	}
}

func (b *BoundMethod) Type() Type {
	return TypeBoundMethod
}

func (b *BoundMethod) Inspect() string {
	return fmt.Sprintf("<fun %s>", b.Method.Fn.Name)
}
