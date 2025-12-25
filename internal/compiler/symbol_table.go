package compiler

type SymbolScope string

const (
	GlobalScope SymbolScope = "GLOBAL"
	LocalScope  SymbolScope = "LOCAL"
	NativeScope SymbolScope = "NATIVE"
	FreeScope   SymbolScope = "FREE"
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
	FreeSymbols    []Symbol
	store          map[string]Symbol
	numDefinitions int
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		store:       make(map[string]Symbol),
		FreeSymbols: []Symbol{},
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

func (s *SymbolTable) defineFree(symbol Symbol) Symbol {
	s.FreeSymbols = append(s.FreeSymbols, symbol)
	newSymbol := Symbol{
		Name:  symbol.Name,
		Scope: FreeScope,
		Index: len(s.FreeSymbols) - 1,
	}
	s.store[symbol.Name] = newSymbol
	return newSymbol
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

// Resolve looks up a symbol by name
func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := s.store[name]
	// If the symbol is not found in this scope, check the outer scope
	if !ok && s.Outer != nil {
		obj, ok = s.Outer.Resolve(name)
		if !ok {
			return obj, ok
		}
		// Global and native symbols don't need to be captured as free variables
		if obj.Scope == GlobalScope || obj.Scope == NativeScope {
			return obj, ok
		}
		// Local or free variables from outer scopes become free variables in this scope
		free := s.defineFree(obj)
		return free, true
	}
	return obj, ok
}

// NumDefinitions returns the number of definitions in this scope
func (s *SymbolTable) NumDefinitions() int {
	return s.numDefinitions
}
