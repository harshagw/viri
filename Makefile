.PHONY: viri repl tidy test wasm build e2e repl-compiler debugger

viri:
	go run cmd/viri/main.go examples/demo.viri

repl:
	go run cmd/repl/main.go

repl-compiler:
	go run cmd/repl-compiler/main.go

debugger:
	go build -o debugger cmd/debugger/*.go
	./debugger examples/demo.viri
	rm -f debugger

tidy:
	go mod tidy

build:
	go build -o viri cmd/viri/main.go

test:
	go test ./...

e2e: build
	@echo "========================================="
	@echo "Running E2E tests (Interpreter Engine)"
	@echo "========================================="
	go test -tags=e2e -run TestE2E$$ ./test/...
	@echo ""
	@echo "========================================="
	@echo "Running E2E tests (VM Engine)"
	@echo "========================================="
	go test -tags=e2e -run TestE2E_VM ./test/...
	rm -f viri

build-plugin:
	cd /Users/harsh/code/viri/viri-syntax-plugin && vsce package

web:
	cd viri-web && npm run dev

build-playground:
	GOOS=js GOARCH=wasm go build -o viri-web/public/viri.wasm cmd/web-playground/main.go
	cp $(shell go env GOROOT)/lib/wasm/wasm_exec.js viri-web/public/