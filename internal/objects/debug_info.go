package objects

// LineEntry marks the bytecode offset where a new source line begins.
type LineEntry struct {
	Offset int // first bytecode offset attributed to Line
	Line   int
}

// DebugInfoEntry holds debug information for a single function or module.
// The line table is run-length encoded: one entry per line change, not one
// per bytecode byte.
type DebugInfoEntry struct {
	LineTable []LineEntry
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

func (d *DebugInfo) Add(lineTable []LineEntry, filePath string) int {
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
	if entry == nil || ip < 0 || len(entry.LineTable) == 0 {
		return 0
	}
	// Binary search: greatest entry with Offset <= ip
	lo, hi := 0, len(entry.LineTable)-1
	if ip < entry.LineTable[0].Offset {
		return 0
	}
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if entry.LineTable[mid].Offset <= ip {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return entry.LineTable[lo].Line
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
