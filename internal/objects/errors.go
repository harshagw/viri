package objects

import "github.com/harshagw/viri/internal/token"

// ReturnError is used for function return control flow.
type ReturnError struct {
	Value Object
}

func (e *ReturnError) Error() string { return "return" }

// BreakError is used for loop control flow.
type BreakError struct{}

func (e *BreakError) Error() string { return "break" }

// ContinueError is used for continue control flow in loops.
type ContinueError struct{}

func (e *ContinueError) Error() string { return "continue" }

// RuntimeError is used for runtime errors.
type RuntimeError struct {
	Token   *token.Token
	Message string
}

func (e *RuntimeError) Error() string { return e.Message }
