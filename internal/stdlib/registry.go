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

// GetExportNameSet returns the set of names a stdlib module exports.
//
// Only membership matters. Stdlib exports are resolved by name into the
// constants pool, not by index — an earlier version handed back indices
// assigned by ranging over a Go map, which made them nondeterministic across
// runs. They were never read, but returning a set says so rather than
// inviting someone to start trusting them.
func GetExportNameSet(name string) (map[string]struct{}, bool) {
	module, ok := Get(name)
	if !ok {
		return nil, false
	}

	names := make(map[string]struct{}, len(module.Exports))
	for exportName := range module.Exports {
		names[exportName] = struct{}{}
	}
	return names, true
}

// init registers all standard library modules.
func init() {
	Register(MathModule)
}
