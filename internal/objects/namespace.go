package objects

import (
	"fmt"

	"github.com/harshagw/viri/internal/token"
)

// Namespace represents an imported module's exported symbols.
type Namespace struct {
	Name    string           // module name/alias
	Exports map[string]Object 
}

func NewNamespace(name string, exports map[string]Object) *Namespace {
	return &Namespace{
		Name:    name,
		Exports: exports,
	}
}

func (n *Namespace) Type() Type {
	return TypeNamespace
}

func (n *Namespace) Inspect() string {
	return fmt.Sprintf("<namespace %s>", n.Name)
}

func (n *Namespace) Get(name *token.Token) (Object, error) {
	if obj, ok := n.Exports[name.Lexeme]; ok {
		return obj, nil
	}
	return nil, fmt.Errorf("symbol '%s' is not exported from namespace '%s'", name.Lexeme, n.Name)
}

