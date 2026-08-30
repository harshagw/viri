package stdlib

import (
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/types"
)

// NativeModule represents a built-in module implemented in Go.
type NativeModule struct {
	Name    string                    // module name (e.g., "std:math")
	Exports map[string]objects.Object // exported symbols
	// Signatures gives each export its compile-time type, so calls into the
	// module are checked like any other call. Every name in Exports must
	// appear here; ModuleType is what the checker reads.
	Signatures map[string]types.Type
}

// NewNativeModule creates a new native module.
func NewNativeModule(name string, exports map[string]objects.Object, signatures map[string]types.Type) *NativeModule {
	return &NativeModule{
		Name:       name,
		Exports:    exports,
		Signatures: signatures,
	}
}

// ModuleType returns the compile-time view of a stdlib module, or nil if the
// module is unknown.
func ModuleType(name string) *types.Module {
	mod, ok := Get(name)
	if !ok {
		return nil
	}
	return &types.Module{Path: name, Exports: mod.Signatures}
}

// GetExportNames returns a list of all exported symbol names.
func (m *NativeModule) GetExportNames() []string {
	names := make([]string, 0, len(m.Exports))
	for name := range m.Exports {
		names = append(names, name)
	}
	return names
}
