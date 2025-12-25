package compiler

type SymbolScope string

const (
	GlobalScope SymbolScope = "GLOBAL"
	LocalScope  SymbolScope = "LOCAL"
	NativeScope SymbolScope = "NATIVE"
)

// Symbol represents a named binding in the symbol table
type Symbol struct {
	Name    string
	Scope   SymbolScope
	Index   int
	IsConst bool
}

type SymbolTable struct {
	Outer          *SymbolTable
	store          map[string]Symbol
	numDefinitions int
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		store: make(map[string]Symbol),
	}
}

// DefineNative defines a native function in the symbol table
func (s *SymbolTable) DefineNative(index int, name string) Symbol {
	symbol := Symbol{
		Name:    name,
		Scope:   NativeScope,
		Index:   index,
		IsConst: true,
	}
	s.store[name] = symbol
	return symbol
}

func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	return s
}

// Define creates a new symbol in the table
func (s *SymbolTable) Define(name string, isConst bool) Symbol {
	symbol := Symbol{
		Name:    name,
		Index:   s.numDefinitions,
		IsConst: isConst,
	}

	if s.Outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}

	s.store[name] = symbol
	s.numDefinitions++
	return symbol
}

// Resolve looks up a symbol by name, checking outer scopes if necessary
func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := s.store[name]
	if !ok && s.Outer != nil {
		return s.Outer.Resolve(name)
	}
	return obj, ok
}

// NumDefinitions returns the number of definitions in this scope
func (s *SymbolTable) NumDefinitions() int {
	return s.numDefinitions
}
