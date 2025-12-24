.PHONY: viri repl tidy test wasm build e2e repl-compiler

viri:
	go run cmd/viri/main.go examples/demo.viri

repl:
	go run cmd/repl/main.go

repl-compiler:
	go run cmd/repl-compiler/main.go

tidy:
	go mod tidy

build:
	go build -o viri cmd/viri/main.go

test:
	go test ./...

e2e: build
	go test -tags=e2e ./test/...
	rm -f viri

build-plugin:
	cd /Users/harsh/code/viri/viri-syntax-plugin && vsce package

web:
	cd viri-web && npm run dev

build-playground:
	GOOS=js GOARCH=wasm go build -o viri-web/public/viri.wasm cmd/web-playground/main.go
	cp $$(go env GOROOT)/misc/wasm/wasm_exec.js viri-web/public/