package objects

import "github.com/harshagw/viri/internal/token"

// DiagnosticHandler receives parse/type/runtime diagnostics.
type DiagnosticHandler interface {
	Error(token.Token, string)
	Warn(token.Token, string)
}

// Diagnostic captures a diagnostic entry for collection/inspection.
type Diagnostic struct {
	Token   token.Token
	Message string
}

// DiagnosticCollector is a simple in-memory handler useful for tests and plumbing.
type DiagnosticCollector struct {
	Errors   []Diagnostic
	Warnings []Diagnostic
}

func (c *DiagnosticCollector) Error(tok token.Token, msg string) {
	c.Errors = append(c.Errors, Diagnostic{Token: tok, Message: msg})
}

func (c *DiagnosticCollector) Warn(tok token.Token, msg string) {
	c.Warnings = append(c.Warnings, Diagnostic{Token: tok, Message: msg})
}
