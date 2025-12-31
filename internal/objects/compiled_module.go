package objects

import "github.com/harshagw/viri/internal/code"

// CompiledProgram represents a fully compiled program with all modules.
type CompiledProgram struct {
	Modules   []CompiledModule // in topological order (dependencies first)
	Constants []Object         // global constants table (merged from all modules)
	DebugInfo *DebugInfo       // debug information (line tables, file paths)
}

// CompiledModule represents a single compiled module.
type CompiledModule struct {
	Instructions code.Instructions
	NumGlobals   int   // slots needed for this module's globals
	Exports      []int // export index -> global slot mapping
	DebugInfoIdx int   // index into DebugInfo.Entries for line table and file path
}
