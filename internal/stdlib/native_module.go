package stdlib

import (
	"github.com/harshagw/viri/internal/objects"
)

// NativeModule represents a built-in module implemented in Go.
type NativeModule struct {
	Name    string                    // module name (e.g., "std:math")
	Exports map[string]objects.Object // exported symbols
}

// NewNativeModule creates a new native module.
func NewNativeModule(name string, exports map[string]objects.Object) *NativeModule {
	return &NativeModule{
		Name:    name,
		Exports: exports,
	}
}

// GetExportNames returns a list of all exported symbol names.
func (m *NativeModule) GetExportNames() []string {
	names := make([]string, 0, len(m.Exports))
	for name := range m.Exports {
		names = append(names, name)
	}
	return names
}
