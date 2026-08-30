//go:build js && wasm

// Command web-playground is the WASM build behind the viri-web playground.
//
// This file is the JavaScript boundary; the pipeline it drives lives in
// run.go, which carries no build tag so it can be tested on any platform.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"syscall/js"
)

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

	handler := &playgroundHandler{errors: []string{}, warnings: []string{}}
	var outBuf bytes.Buffer

	finalResult := execute(args[0].String(), &outBuf, handler)

	jsonBytes, err := json.Marshal(Response{
		Result:   finalResult,
		Output:   outBuf.String(),
		Errors:   handler.errors,
		Warnings: handler.warnings,
	})
	if err != nil {
		return fmt.Sprintf(`{"errors": ["Internal error: %s"]}`, err.Error())
	}
	return string(jsonBytes)
}
