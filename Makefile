.PHONY: viri repl tidy test wasm

viri:
	go run cmd/viri/main.go examples/demo.viri

repl:
	go run cmd/repl/main.go

tidy:
	go mod tidy

test:
	go test ./...

build-plugin:
	cd /Users/harsh/code/viri/viri-syntax-plugin && vsce package

web:
	cd viri-web && npm run dev

wasm:
	GOOS=js GOARCH=wasm go build -o viri-web/public/viri.wasm cmd/web-playground/main.go
	cp $$(go env GOROOT)/misc/wasm/wasm_exec.js viri-web/public/