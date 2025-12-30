package compiler

type SymbolScope string

const (
	GlobalScope   SymbolScope = "GLOBAL"
	LocalScope    SymbolScope = "LOCAL"
	NativeScope   SymbolScope = "NATIVE"
	FreeScope     SymbolScope = "FREE"
	FunctionScope SymbolScope = "FUNCTION" // For recursive self-reference
)

// Symbol represents a named binding in the symbol table
type Symbol struct {
	Name       string
	Scope      SymbolScope
	Index      int
	IsConst    bool
	FrameDepth int // function nesting level when defined
}

// ImportInfo tracks an imported module's exports
type ImportInfo struct {
	ModuleIndex int
	Exports     map[string]int // export name -> export index
}

type SymbolTable struct {
	Outer          *SymbolTable
	FreeSymbols    []Symbol
	store          map[string]Symbol
	imports        map[string]*ImportInfo // import alias -> module info
	numDefinitions int
	functionName   string
	frameDepth     int // function nesting level (0 = global)
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		store:       make(map[string]Symbol),
		imports:     make(map[string]*ImportInfo),
		FreeSymbols: []Symbol{},
		frameDepth:  0,
	}
}

// NewFunctionScope creates a new scope for a function.
// This increments frameDepth and resets numDefinitions.
func NewFunctionScope(outer *SymbolTable, functionName string) *SymbolTable {
	s := &SymbolTable{
		store:          make(map[string]Symbol),
		imports:        outer.imports, // inherit imports from outer scope
		FreeSymbols:    []Symbol{},
		Outer:          outer,
		functionName:   functionName,
		frameDepth:     outer.frameDepth + 1,
		numDefinitions: 0,
	}
	return s
}

// NewBlockScope creates a new scope for a block within the same function.
// This keeps the same frameDepth and inherits numDefinitions.
func NewBlockScope(outer *SymbolTable) *SymbolTable {
	return &SymbolTable{
		store:          make(map[string]Symbol),
		imports:        outer.imports,     // share imports with parent
		FreeSymbols:    outer.FreeSymbols, // share free symbols with parent
		Outer:          outer,
		functionName:   outer.functionName,
		frameDepth:     outer.frameDepth,     // same frame
		numDefinitions: outer.numDefinitions, // inherit counter
	}
}

// DefineNative defines a native function in the symbol table
func (s *SymbolTable) DefineNative(index int, name string) Symbol {
	symbol := Symbol{
		Name:       name,
		Scope:      NativeScope,
		Index:      index,
		IsConst:    true,
		FrameDepth: 0,
	}
	s.store[name] = symbol
	return symbol
}

func (s *SymbolTable) defineFree(symbol Symbol) Symbol {
	s.FreeSymbols = append(s.FreeSymbols, symbol)
	newSymbol := Symbol{
		Name:       symbol.Name,
		Scope:      FreeScope,
		Index:      len(s.FreeSymbols) - 1,
		FrameDepth: s.frameDepth,
	}
	s.store[symbol.Name] = newSymbol
	return newSymbol
}

// Define creates a new symbol in the table.
func (s *SymbolTable) Define(name string, isConst bool) (Symbol, bool) {
	// Check if name conflicts with an import alias
	if s.IsImportAlias(name) {
		return Symbol{}, false
	}

	symbol := Symbol{
		Name:       name,
		Index:      s.numDefinitions,
		IsConst:    isConst,
		FrameDepth: s.frameDepth,
	}

	if s.frameDepth == 0 {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}

	s.store[name] = symbol
	s.numDefinitions++
	return symbol, true
}

// Resolve looks up a symbol by name
func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := s.store[name]
	// If the symbol is not found in this scope, check for recursive self-reference
	if !ok && s.functionName == name {
		// This is a recursive call - return a special FunctionScope symbol
		return Symbol{Name: name, Scope: FunctionScope, Index: 0, FrameDepth: s.frameDepth}, true
	}
	// If the symbol is not found in this scope, check the outer scope
	if !ok && s.Outer != nil {
		obj, ok = s.Outer.Resolve(name)
		if !ok {
			return obj, ok
		}
		// Global, native, and function symbols don't need to be captured as free variables
		if obj.Scope == GlobalScope || obj.Scope == NativeScope || obj.Scope == FunctionScope {
			return obj, ok
		}
		// If the resolved symbol is in the same frame, return it as-is (block scope)
		if obj.FrameDepth == s.frameDepth {
			return obj, ok
		}
		// Local or free variables from outer function scopes become free variables
		free := s.defineFree(obj)
		return free, true
	}
	return obj, ok
}

// NumDefinitions returns the number of definitions in this scope
func (s *SymbolTable) NumDefinitions() int {
	return s.numDefinitions
}

// DefineImport registers an import alias with its module index and exports
func (s *SymbolTable) DefineImport(alias string, moduleIndex int, exports map[string]int) {
	s.imports[alias] = &ImportInfo{
		ModuleIndex: moduleIndex,
		Exports:     exports,
	}
}

// ResolveImport looks up an import alias and export name, returning (moduleIdx, exportIdx, found)
func (s *SymbolTable) ResolveImport(alias string, exportName string) (int, int, bool) {
	importInfo, ok := s.imports[alias]
	if !ok {
		return 0, 0, false
	}
	exportIdx, ok := importInfo.Exports[exportName]
	if !ok {
		return 0, 0, false
	}
	return importInfo.ModuleIndex, exportIdx, true
}

// IsImportAlias checks if a name is a registered import alias
func (s *SymbolTable) IsImportAlias(name string) bool {
	_, ok := s.imports[name]
	return ok
}
