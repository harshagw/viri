package stdlib

import (
	"strings"
)

const (
	// StdPrefix is the prefix used for standard library modules.
	StdPrefix = "std:"
)

// registry holds all registered native modules.
var registry = make(map[string]*NativeModule)

// Register registers a native module in the registry.
func Register(module *NativeModule) {
	registry[module.Name] = module
}

// Get retrieves a native module by name.
func Get(name string) (*NativeModule, bool) {
	module, ok := registry[name]
	return module, ok
}

// IsStdLib checks if an import path refers to a standard library module.
func IsStdLib(importPath string) bool {
	return strings.HasPrefix(importPath, StdPrefix)
}

// GetExportMap returns a map of export names to indices for a stdlib module.
// This is used by the compiler for module import resolution.
func GetExportMap(name string) (map[string]int, bool) {
	module, ok := Get(name)
	if !ok {
		return nil, false
	}

	exportMap := make(map[string]int)
	idx := 0
	for exportName := range module.Exports {
		exportMap[exportName] = idx
		idx++
	}
	return exportMap, true
}

// init registers all standard library modules.
func init() {
	Register(MathModule)
}
