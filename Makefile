.PHONY: viri repl tidy test wasm build e2e bench repl-compiler debugger

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
	@echo "Running E2E tests (Interpreter Engine, frozen)"
	@echo "========================================="
	go test -tags=e2e -run TestE2E$$ ./test/...
	@echo ""
	@echo "========================================="
	@echo "Running E2E tests (VM Engine)"
	@echo "========================================="
	go test -tags=e2e -run 'TestE2E_VM_' ./test/...
	rm -f viri

# Benchmarks run the VM only; the interpreter is frozen and no longer runs the
# same language. COUNT repetitions feed benchstat's variance estimate, so keep
# it at 6+ for any number you intend to act on.
BENCH_COUNT ?= 6
BENCH_FILTER ?= .
BENCH_DIR := test/benchmarks
BASELINE := $(BENCH_DIR)/baseline.txt
CURRENT := $(BENCH_DIR)/current.txt

bench:
	go test -tags=e2e -run XXX -bench '$(BENCH_FILTER)' -benchmem ./test/...

# Record the current tree as the reference to compare future runs against.
# The run writes to a temporary file and only replaces the baseline once it has
# succeeded, so a failed or interrupted run leaves the old baseline intact
# rather than a truncated one. Output still streams: the exit status travels
# out of the pipeline in a side file, since 'set -o pipefail' is not portable.
bench-save:
	@mkdir -p $(BENCH_DIR)
	@tmp=$$(mktemp); rc=$$(mktemp); \
	{ go test -tags=e2e -run XXX -bench '$(BENCH_FILTER)' -benchmem \
		-count=$(BENCH_COUNT) ./test/...; echo $$? >$$rc; } | tee $$tmp; \
	status=$$(cat $$rc); rm -f $$rc; \
	if [ "$$status" -eq 0 ]; then \
		mv $$tmp $(BASELINE); \
		echo ""; echo "baseline written to $(BASELINE)"; \
	else \
		rm -f $$tmp; \
		echo ""; echo "benchmarks failed; $(BASELINE) left unchanged"; \
		exit 1; \
	fi

# Measure the working tree and diff it against the saved baseline. benchstat
# reports the delta with a p-value, so noise does not read as a change.
bench-cmp:
	@test -f $(BASELINE) || { echo "no baseline: run 'make bench-save' first"; exit 1; }
	@mkdir -p $(BENCH_DIR)
	@tmp=$$(mktemp); rc=$$(mktemp); \
	{ go test -tags=e2e -run XXX -bench '$(BENCH_FILTER)' -benchmem \
		-count=$(BENCH_COUNT) ./test/...; echo $$? >$$rc; } | tee $$tmp; \
	status=$$(cat $$rc); rm -f $$rc; \
	if [ "$$status" -ne 0 ]; then rm -f $$tmp; exit 1; fi; \
	mv $$tmp $(CURRENT)
	@echo ""
	go run golang.org/x/perf/cmd/benchstat@latest $(BASELINE) $(CURRENT)

build-plugin:
	cd /Users/harsh/code/viri/viri-syntax-plugin && vsce package

web:
	cd viri-web && npm run dev

build-playground:
	GOOS=js GOARCH=wasm go build -o viri-web/public/viri.wasm ./cmd/web-playground/
	cp $(shell go env GOROOT)/lib/wasm/wasm_exec.js viri-web/public/