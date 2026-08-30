package checker

import (
	"github.com/harshagw/viri/internal/types"
)

// binding is one name in scope.
type binding struct {
	typ     types.Type
	isConst bool
}

// scope is one lexical level. The stack shape follows the interpreter's
// resolver: push on entering a block, pop on leaving, look up from the
// innermost outward.
type scope struct {
	names map[string]*binding
}

func newScope() *scope {
	return &scope{names: make(map[string]*binding)}
}

// builtinTypes are the predeclared type names. They are not keywords: the
// parser emits a plain NamedType for `number` exactly as it does for `Shape`,
// and this map is the only place that distinguishes them. Adding a builtin
// type later is an entry here, not a scanner change.
var builtinTypes = map[string]types.Type{
	"number": types.Number,
	"string": types.String,
	"bool":   types.Bool,
}

// IsBuiltinTypeName reports whether a name is predeclared. Declaring anything
// with one of these names is an error — it would shadow the type in value
// position while leaving it reachable in type position, so the same word would
// mean two different things depending on where it appeared.
func IsBuiltinTypeName(name string) bool {
	_, ok := builtinTypes[name]
	return ok
}

// declare binds a name in the innermost scope. It reports false if the name is
// already bound at this level.
func (c *Checker) declare(name string, t types.Type, isConst bool) bool {
	s := c.scopes[len(c.scopes)-1]
	if _, exists := s.names[name]; exists {
		return false
	}
	s.names[name] = &binding{typ: t, isConst: isConst}
	return true
}

// lookup finds a name, innermost scope first.
func (c *Checker) lookup(name string) (*binding, bool) {
	for i := len(c.scopes) - 1; i >= 0; i-- {
		if b, ok := c.scopes[i].names[name]; ok {
			return b, true
		}
	}
	return nil, false
}

func (c *Checker) beginScope() {
	c.scopes = append(c.scopes, newScope())
}

func (c *Checker) endScope() {
	c.scopes = c.scopes[:len(c.scopes)-1]
}
