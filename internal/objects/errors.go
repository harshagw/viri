package objects

// VMRuntimeError is used for runtime errors in the VM.
type VMRuntimeError struct {
	Message  string
	Line     int
	FilePath string
}

func (e *VMRuntimeError) Error() string { return e.Message }
