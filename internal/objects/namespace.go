package objects

import (
	"fmt"

	"github.com/harshagw/viri/internal/token"
)

// Namespace represents an imported module's exported symbols.
// A namespace backed by Env reads the module's environment live, so the
// importer observes assignments made after the module finished loading
// (e.g. an exported counter mutated by an exported function). Stdlib
// namespaces have no environment and use the static Exports map.
type Namespace struct {
	Name          string            // module name/alias
	Exports       map[string]Object // static exports (stdlib)
	Env           *Environment      // module environment (user modules)
	ExportedNames map[string]bool   // names exported by the module
}

func NewNamespace(name string, exports map[string]Object) *Namespace {
	return &Namespace{
		Name:    name,
		Exports: exports,
	}
}

// NewLiveNamespace creates a namespace that resolves exported names against
// the module's environment at access time.
func NewLiveNamespace(name string, env *Environment, exportedNames map[string]bool) *Namespace {
	return &Namespace{
		Name:          name,
		Env:           env,
		ExportedNames: exportedNames,
	}
}

func (n *Namespace) Type() Type {
	return TypeNamespace
}

func (n *Namespace) Inspect() string {
	return fmt.Sprintf("<namespace %s>", n.Name)
}

func (n *Namespace) Get(name *token.Token) (Object, error) {
	if n.Env != nil {
		if !n.ExportedNames[name.Lexeme] {
			return nil, fmt.Errorf("symbol '%s' is not exported from namespace '%s'", name.Lexeme, n.Name)
		}
		value, err := n.Env.Get(name.Lexeme)
		if err != nil {
			return nil, fmt.Errorf("symbol '%s' is not exported from namespace '%s'", name.Lexeme, n.Name)
		}
		return value, nil
	}
	if obj, ok := n.Exports[name.Lexeme]; ok {
		return obj, nil
	}
	return nil, fmt.Errorf("symbol '%s' is not exported from namespace '%s'", name.Lexeme, n.Name)
}
