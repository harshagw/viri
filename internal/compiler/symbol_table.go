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
	FrameDepth int  // function nesting level when defined
	Pending    bool // hoisted but its declaration not yet compiled
}

// ImportInfo tracks an imported module's exports
type ImportInfo struct {
	ModuleIndex int
	Exports     map[string]int // export name -> export index
	IsStdlib    bool           // true if this is a stdlib import
	StdlibName  string         // stdlib module name (e.g., "std:math")
}

type SymbolTable struct {
	Outer          *SymbolTable
	FreeSymbols    []Symbol
	store          map[string]Symbol
	imports        map[string]*ImportInfo // import alias -> module info
	numDefinitions int
	numBlockLocals int // frameDepth 0 only: local slots used by block-scoped variables
	functionName   string
	frameDepth     int  // function nesting level (0 = global)
	isBlock        bool // true for block scopes (not function or module scopes)
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
// This keeps the same frameDepth and inherits the slot counters; the parent
// must call AbsorbBlockCounters when the block ends so slots allocated inside
// the block stay reserved in the enclosing frame.
func NewBlockScope(outer *SymbolTable) *SymbolTable {
	return &SymbolTable{
		store:          make(map[string]Symbol),
		imports:        outer.imports,     // share imports with parent
		FreeSymbols:    outer.FreeSymbols, // share free symbols with parent
		Outer:          outer,
		functionName:   outer.functionName,
		frameDepth:     outer.frameDepth,     // same frame
		numDefinitions: outer.numDefinitions, // inherit counter
		numBlockLocals: outer.numBlockLocals, // inherit counter
		isBlock:        true,
	}
}

// AbsorbBlockCounters propagates slot counters from an ended block scope back
// to its parent, so a later declaration in the parent (or a sibling block)
// cannot reuse a slot whose value is still live in the frame.
func (s *SymbolTable) AbsorbBlockCounters(block *SymbolTable) {
	if block.numDefinitions > s.numDefinitions {
		s.numDefinitions = block.numDefinitions
	}
	if block.numBlockLocals > s.numBlockLocals {
		s.numBlockLocals = block.numBlockLocals
	}
	// Free symbols are shared by slice value; re-sync in case the block appended.
	s.FreeSymbols = block.FreeSymbols
}

// NumBlockLocals returns the number of main-frame local slots used by
// block-scoped variables at module level.
func (s *SymbolTable) NumBlockLocals() int {
	return s.numBlockLocals
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
		IsConst:    symbol.IsConst, // a captured const stays const
		FrameDepth: s.frameDepth,
	}
	s.store[symbol.Name] = newSymbol
	return newSymbol
}

// Define creates a new symbol in the table.
func (s *SymbolTable) Define(name string, isConst bool) (Symbol, bool) {
	// An import alias cannot be redeclared at module scope, but locals and
	// parameters may shadow it (matching the interpreter).
	if s.IsImportAlias(name) && s.frameDepth == 0 && !s.isBlock {
		return Symbol{}, false
	}

	// Reject duplicate declarations in the same scope (shadowing a native
	// function is allowed, matching the interpreter).
	if existing, exists := s.store[name]; exists && existing.Scope != NativeScope {
		return Symbol{}, false
	}

	symbol := Symbol{
		Name:       name,
		IsConst:    isConst,
		FrameDepth: s.frameDepth,
	}

	switch {
	case s.frameDepth == 0 && !s.isBlock:
		symbol.Scope = GlobalScope
		symbol.Index = s.numDefinitions
		s.numDefinitions++
	case s.frameDepth == 0 && s.isBlock:
		// A block-scoped variable at module level lives in the module main
		// frame, not in a global slot: this gives it block lifetime and
		// makes it capturable per-iteration, like function locals.
		symbol.Scope = LocalScope
		symbol.Index = s.numBlockLocals
		s.numBlockLocals++
	default:
		symbol.Scope = LocalScope
		symbol.Index = s.numDefinitions
		s.numDefinitions++
	}

	s.store[name] = symbol
	return symbol, true
}

// Hoist pre-declares a module-level name before its statement is compiled,
// so function bodies can reference globals declared later in the file
// (mutual recursion between top-level functions). The symbol stays Pending
// until DefineOrActivate is called for its declaration.
func (s *SymbolTable) Hoist(name string, isConst bool) (Symbol, bool) {
	symbol, ok := s.Define(name, isConst)
	if !ok {
		return symbol, false
	}
	symbol.Pending = true
	s.store[name] = symbol
	return symbol, true
}

// DefineOrActivate defines a new symbol, or activates a hoisted one when its
// declaration is reached.
func (s *SymbolTable) DefineOrActivate(name string, isConst bool) (Symbol, bool) {
	if symbol, ok := s.store[name]; ok && symbol.Pending {
		symbol.Pending = false
		symbol.IsConst = isConst
		s.store[name] = symbol
		return symbol, true
	}
	return s.Define(name, isConst)
}

// lookupChain finds a binding by walking the scope chain without any side
// effects (no free-variable capture, no self-reference shortcut).
func (s *SymbolTable) lookupChain(name string) (Symbol, bool) {
	for t := s; t != nil; t = t.Outer {
		if symbol, ok := t.store[name]; ok {
			return symbol, true
		}
	}
	return Symbol{}, false
}

// Resolve looks up a symbol by name
func (s *SymbolTable) Resolve(name string) (Symbol, bool) {
	return s.resolve(name, s.frameDepth)
}

// resolve walks outward from the requesting scope. fromDepth is the frame
// depth of the scope the lookup started in: the self-reference shortcut
// (OpGetCurrentClosure) is only valid inside the named function's own frame —
// a nested function that mentions the enclosing function's name must resolve
// the ordinary binding instead of compiling to a call to itself.
func (s *SymbolTable) resolve(name string, fromDepth int) (Symbol, bool) {
	obj, ok := s.store[name]
	// A hoisted global whose declaration has not been compiled yet is not
	// readable from code that executes immediately (module level), only
	// from function bodies, which run later.
	if ok && obj.Pending && fromDepth == 0 {
		ok = false
	}
	// Recursive self-reference, valid only within the function's own frame.
	// When the name also has a global (or native) binding, prefer that:
	// it is late-bound, matching the interpreter. The closure shortcut is
	// for local function declarations, whose capture cell would not be
	// filled in yet.
	if !ok && s.functionName == name && s.frameDepth == fromDepth {
		if outer, found := s.lookupChain(name); !found || (outer.Scope != GlobalScope && outer.Scope != NativeScope) {
			return Symbol{Name: name, Scope: FunctionScope, Index: 0, FrameDepth: s.frameDepth}, true
		}
	}
	// If the symbol is not found in this scope, check the outer scope
	if !ok && s.Outer != nil {
		obj, ok = s.Outer.resolve(name, fromDepth)
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
		IsStdlib:    false,
	}
}

// DefineStdlibImport registers a stdlib import alias
func (s *SymbolTable) DefineStdlibImport(alias string, stdlibName string, exports map[string]int) {
	s.imports[alias] = &ImportInfo{
		ModuleIndex: -1, // not used for stdlib
		Exports:     exports,
		IsStdlib:    true,
		StdlibName:  stdlibName,
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

// ResolveStdlibImport looks up a stdlib import, returning (stdlibName, exportName, found)
func (s *SymbolTable) ResolveStdlibImport(alias string, exportName string) (string, string, bool) {
	importInfo, ok := s.imports[alias]
	if !ok || !importInfo.IsStdlib {
		return "", "", false
	}
	if _, ok := importInfo.Exports[exportName]; !ok {
		return "", "", false
	}
	return importInfo.StdlibName, exportName, true
}

// IsStdlibImport checks if an import alias refers to a stdlib module
func (s *SymbolTable) IsStdlibImport(alias string) bool {
	importInfo, ok := s.imports[alias]
	return ok && importInfo.IsStdlib
}

// IsImportAlias checks if a name is a registered import alias
func (s *SymbolTable) IsImportAlias(name string) bool {
	_, ok := s.imports[name]
	return ok
}
