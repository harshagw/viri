package objects

// DebugInfoEntry holds debug information for a single function or module.
type DebugInfoEntry struct {
	LineTable []int  // maps bytecode offset -> source line number
	FilePath  string // source file path
}

// DebugInfo holds all debug information for a compiled program.
type DebugInfo struct {
	Entries []DebugInfoEntry
}

func NewDebugInfo() *DebugInfo {
	return &DebugInfo{
		Entries: []DebugInfoEntry{},
	}
}

func (d *DebugInfo) Add(lineTable []int, filePath string) int {
	idx := len(d.Entries)
	d.Entries = append(d.Entries, DebugInfoEntry{
		LineTable: lineTable,
		FilePath:  filePath,
	})
	return idx
}

// Get returns the debug entry at the given index.
func (d *DebugInfo) Get(idx int) *DebugInfoEntry {
	if d == nil || idx < 0 || idx >= len(d.Entries) {
		return nil
	}
	return &d.Entries[idx]
}

// GetLine returns the line number for a given debug index and instruction pointer.
// Returns 0 if not found.
func (d *DebugInfo) GetLine(idx int, ip int) int {
	entry := d.Get(idx)
	if entry == nil || ip < 0 || ip >= len(entry.LineTable) {
		return 0
	}
	return entry.LineTable[ip]
}

// GetFilePath returns the file path for a given debug index.
// Returns empty string if not found.
func (d *DebugInfo) GetFilePath(idx int) string {
	entry := d.Get(idx)
	if entry == nil {
		return ""
	}
	return entry.FilePath
}
