//go:build e2e

package test

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/harshagw/viri/internal/checker"
	"github.com/harshagw/viri/internal/compiler"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/parser"
	"github.com/harshagw/viri/internal/scanner"
	"github.com/harshagw/viri/internal/vm"
)

// These benchmarks live beside the end-to-end suite rather than inside
// internal/vm, because they exercise the whole pipeline rather than one
// package: the numbers only mean something with the checker's type
// information feeding codegen.
//
// Run with: go test -tags=e2e -run XXX -bench . ./test/...

// compileSource takes a program to bytecode once, so a benchmark measures the
// VM rather than the front end.
func compileSource(b *testing.B, src string) *objects.CompiledProgram {
	b.Helper()

	tokens, err := scanner.New(bytes.NewBufferString(src), nil).Scan()
	if err != nil {
		b.Fatalf("scan: %v", err)
	}
	collector := &objects.DiagnosticCollector{}
	p := parser.NewParser(tokens, collector)
	p.SetFilePath("bench.viri")
	mod, err := p.Parse()
	if err != nil {
		b.Fatalf("parse: %v (%v)", err, collector.Errors)
	}

	ck := checker.New(collector)
	if !ck.CheckIncremental(mod.GetAllStatements()) {
		b.Fatalf("type errors: %v", collector.Errors)
	}

	comp := compiler.New(collector)
	comp.SetTypeInfo(ck.ExprTypes(), ck.ClassTypes())
	for _, stmt := range mod.GetAllStatements() {
		if err := comp.Compile(stmt); err != nil {
			b.Fatalf("compile: %v", err)
		}
	}
	return comp.Result()
}

func runBench(b *testing.B, src string) {
	b.Helper()
	program := compileSource(b, src)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		machine := vm.New(program)
		machine.SetStdout(io.Discard)
		if err := machine.RunProgram(); err != nil {
			b.Fatalf("runtime error: %v", err)
		}
	}
}

// ---------------------------------------------------------------------------
// Comparison probes
//
// These ten mirror the probe set used for the original interpreter-vs-VM
// measurement, at the same workload sizes, so ns/op here is directly
// comparable to the wall-clock seconds recorded there. Two differences worth
// knowing when reading the numbers:
//
//   - Scanning, parsing, checking and codegen happen once in compileSource and
//     are outside the timer. The original figures timed the whole process, so
//     they also carried startup and compilation.
//   - allocs/op and B/op replace the instrumented runtime.MemStats counts.
//     They measure the same thing — churn — but per run rather than per
//     process, and they need no special build.
//
// The interpreter is deliberately absent. It is frozen and no longer runs the
// same language, so there is nothing left to compare against.
// ---------------------------------------------------------------------------

// fib(27) is 832k calls: call frames, argument passing, and a numeric compare
// per call, with almost nothing else in the loop.
func BenchmarkFib27(b *testing.B) {
	runBench(b, `fun fib(n: number): number {
		if (n < 2) { return n; }
		return fib(n - 1) + fib(n - 2);
	}
	print fib(27);`)
}

// 3M iterations of loop condition, increment and an add — the tightest
// arithmetic path there is, and the one typed opcodes target most directly.
func BenchmarkArithmeticLoop3M(b *testing.B) {
	runBench(b, `var total: number = 0;
	for (var i: number = 0; i < 3000000; i = i + 1) {
		total = total + i;
	}
	print total;`)
}

// 1M calls to a trivial top-level function: frame push/pop cost, isolated.
func BenchmarkFunctionCallLoop1M(b *testing.B) {
	runBench(b, `fun id(n: number): number { return n; }
	var total: number = 0;
	for (var i: number = 0; i < 1000000; i = i + 1) {
		total = total + id(i);
	}
	print total;`)
}

// 2M index operations — 1M reads and 1M writes against a fixed-size array.
func BenchmarkArrayReadWrite2M(b *testing.B) {
	// Viri has no modulo operator, so the index cycles with an explicit
	// counter. That adds one compare and one branch per iteration to every
	// probe below that needs to wrap.
	runBench(b, `var a: []number = [0, 0, 0, 0, 0, 0, 0, 0];
	var total: number = 0;
	var k: number = 0;
	for (var i: number = 0; i < 1000000; i = i + 1) {
		a[k] = i;
		total = total + a[k];
		k = k + 1;
		if (k > 7) { k = 0; }
	}
	print total;`)
}

// 1M method calls. Since Phase 3 the receiver's fields are slots, so this is
// also the clearest read on that change at scale.
func BenchmarkMethodCallLoop1M(b *testing.B) {
	runBench(b, `class Counter {
		n: number;
		init() { this.n = 0; }
		inc() { this.n = this.n + 1; }
	}
	var c: Counter = Counter();
	for (var i: number = 0; i < 1000000; i = i + 1) {
		c.inc();
	}
	print c.n;`)
}

// 500k closure creations and calls: each iteration boxes a fresh cell and
// builds a closure over it.
func BenchmarkClosureCreateCall500k(b *testing.B) {
	runBench(b, `fun makeAdder(n: number): fun(number): number {
		return fun(x: number): number { return x + n; };
	}
	var total: number = 0;
	for (var i: number = 0; i < 500000; i = i + 1) {
		var add: fun(number): number = makeAdder(i);
		total = total + add(1);
	}
	print total;`)
}

// 300k global reads from three call frames down, so the read walks past two
// enclosing frames to reach module scope.
func BenchmarkGlobalReadThrough3Frames300k(b *testing.B) {
	runBench(b, `var g: number = 1;
	fun inner(): number { return g; }
	fun mid(): number { return inner(); }
	fun outer(): number { return mid(); }
	var total: number = 0;
	for (var i: number = 0; i < 300000; i = i + 1) {
		total = total + outer();
	}
	print total;`)
}

// 200k string-keyed map operations — 100k sets and 100k gets over a small key
// space, so it measures hashing and probing rather than growth.
func BenchmarkHashSetGet200k(b *testing.B) {
	runBench(b, `var h: map[string]number = {};
	var keys: []string = ["a", "b", "c", "d", "e", "f", "g", "h"];
	var total: number = 0;
	var idx: number = 0;
	for (var i: number = 0; i < 100000; i = i + 1) {
		var k: string = keys[idx];
		h[k] = i;
		total = total + h[k];
		idx = idx + 1;
		if (idx > 7) { idx = 0; }
	}
	print total;`)
}

// 20k concatenations. Every step allocates a new string, so this is a pure
// allocator probe — and the one place the two engines were level.
func BenchmarkStringConcat20k(b *testing.B) {
	runBench(b, `var s: string = "";
	for (var i: number = 0; i < 20000; i = i + 1) {
		s = s + "x";
	}
	print len(s);`)
}

// 500k distinct entries, so the map grows and rehashes repeatedly. This is the
// probe that dominated peak RSS in the original run.
//
// Viri has no number-to-string conversion — the only natives are clock and len
// — so the keys come from a nested walk over a generated array of literals,
// giving 708*708 = 501,264 distinct keys. That means one concatenation per
// entry is inside the measurement. Any real build of half a million keys pays
// something equivalent, but it does mean this probe is not purely the map.
func BenchmarkHashBuild500k(b *testing.B) {
	const side = 708

	var keys strings.Builder
	keys.WriteString("var keys: []string = [")
	for i := 0; i < side; i++ {
		if i > 0 {
			keys.WriteString(", ")
		}
		fmt.Fprintf(&keys, "%q", "k"+strconv.Itoa(i))
	}
	keys.WriteString("];")

	runBench(b, keys.String()+`
	var h: map[string]number = {};
	for (var i: number = 0; i < `+strconv.Itoa(side)+`; i = i + 1) {
		for (var j: number = 0; j < `+strconv.Itoa(side)+`; j = j + 1) {
			h[keys[i] + keys[j]] = j;
		}
	}
	print len(h);`)
}
