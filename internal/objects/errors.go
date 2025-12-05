package objects

// ReturnError is used for function return control flow.
type ReturnError struct {
	Value interface{}
}

func (e *ReturnError) Error() string { return "return" }

// BreakError is used for loop control flow.
type BreakError struct{}

func (e *BreakError) Error() string { return "break" }
