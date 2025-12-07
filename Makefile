.PHONY: viri repl tidy test

viri:
	go run cmd/viri/main.go test.viri

repl:
	go run cmd/repl/main.go

tidy:
	go mod tidy

test:
	go test ./...

build-plugin:
	cd /Users/harsh/code/viri/viri-syntax-plugin && vsce package
