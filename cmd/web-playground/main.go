//go:build js && wasm

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/interp"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/parser"
	"github.com/harshagw/viri/internal/scanner"
	"github.com/harshagw/viri/internal/token"
)

type Response struct {
	Result   string   `json:"result"`
	Output   string   `json:"output"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

func main() {
	c := make(chan struct{}, 0)
	js.Global().Set("runViri", js.FuncOf(runViri))
	<-c
}

func runViri(this js.Value, args []js.Value) (ret interface{}) {
	defer func() {
		if r := recover(); r != nil {
			errResp := Response{
				Errors: []string{fmt.Sprintf("Internal Panic: %v", r)},
			}
			jsonBytes, _ := json.Marshal(errResp)
			ret = string(jsonBytes)
		}
	}()

	if len(args) == 0 {
		return "Error: No input provided"
	}
	input := args[0].String()

	handler := &replHandler{errors: []string{}, warnings: []string{}}
	interpreter := interp.NewInterpreter(nil)

	var outBuf bytes.Buffer
	interpreter.SetStdout(&outBuf)

	finalResult := execute(input, interpreter, handler)

	resp := Response{
		Result:   finalResult,
		Output:   outBuf.String(),
		Errors:   handler.errors,
		Warnings: handler.warnings,
	}

	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		return fmt.Sprintf(`{"errors": ["Internal error: %s"]}`, err.Error())
	}

	return string(jsonBytes)
}

func execute(source string, interpreter *interp.Interpreter, handler *replHandler) string {
	if source == "" {
		return ""
	}

	code := bytes.NewBufferString(source)

	sc := scanner.New(code, nil)
	tokens, err := sc.Scan()
	if err != nil {
		handler.errors = append(handler.errors, fmt.Sprintf("Scanning error: %v", err))
		return ""
	}

	p := parser.NewParser(tokens, handler)
	p.SetFilePath("<playground>")
	module, err := p.Parse()
	if err != nil || handler.hasErrors {
		return ""
	}

	stmts := module.GetAllStatements()
	mod := ast.NewModule("<playground>", nil, stmts)

	res := parser.NewResolver(handler)
	locals, err := res.Resolve(mod)
	if err != nil || handler.hasErrors {
		return ""
	}

	interpreter.SetLocals(locals)
	_, err = interpreter.Interpret(stmts)
	if err != nil {
		handler.errors = append(handler.errors, fmt.Sprintf("Runtime error: %v", err))
		return ""
	}

	return ""
}

type replHandler struct {
	hasErrors bool
	errors    []string
	warnings  []string
}

var _ objects.DiagnosticHandler = (*replHandler)(nil)

func (h *replHandler) Error(tok token.Token, msg string) {
	h.errors = append(h.errors, fmt.Sprintf("Line %d: %s", tok.Line, msg))
	h.hasErrors = true
}

func (h *replHandler) Warn(tok token.Token, msg string) {
	h.warnings = append(h.warnings, fmt.Sprintf("Line %d: %s", tok.Line, msg))
}
