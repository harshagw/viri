package objects

import "fmt"

// CompiledClass represents a class in the VM.
type CompiledClass struct {
	Name       string
	Methods    map[string]*Closure // Method name -> closure
	SuperClass *CompiledClass      // nil if no superclass
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
type CompiledInstance struct {
	Class  *CompiledClass
	Fields map[string]Object
}

func NewCompiledInstance(class *CompiledClass) *CompiledInstance {
	return &CompiledInstance{
		Class:  class,
		Fields: make(map[string]Object),
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
	return fmt.Sprintf("<bound_method %s>", b.Method.Fn.Name)
}
