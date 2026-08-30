//go:build !(js && wasm)

package main

import (
	"fmt"
	"os"
)

// This command is meant to be built for WebAssembly:
//
//	GOOS=js GOARCH=wasm go build -o viri-web/public/viri.wasm cmd/web-playground/main.go
//
// A native build exists only so `go build ./...` and `go test ./...` cover
// run.go, which carries no build tag precisely so it can be tested without a
// browser.
func main() {
	fmt.Fprintln(os.Stderr, "web-playground is a WebAssembly target; build it with GOOS=js GOARCH=wasm (see `make wasm`).")
	os.Exit(1)
}
