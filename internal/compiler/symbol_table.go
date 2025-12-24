package compiler

type SymbolScope string

const (
	GlobalScope SymbolScope = "GLOBAL"
)

// Symbol represents a named binding in the symbol table
type Symbol struct {
	Name    string
	Scope   SymbolScope
	Index   int
	IsConst bool
}

type SymbolTable struct {
	store          map[string]Symbol
	numDefinitions int
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		store: make(map[string]Symbol),
	}
}

// Define creates a new symbol in the table
func (s *SymbolTable) Define(name string, isConst bool) Symbol {
	symbol := Symbol{
		Name:    name,
		Scope:   GlobalScope,
		Index:   s.numDefinitions,
		IsConst: isConst,
	}
	s.store[name] = symbol
	s.numDefinitions++
	return symbol
}

// Resolve looks up a symbol by name
func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := s.store[name]
	return obj, ok
}
