//go:build e2e

package test

import (
	"bytes"
	"io"
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

// Field access is the change with the most to gain: reads and writes were a
// map[string]Object lookup and are now an index into []Object.
func BenchmarkFieldAccess(b *testing.B) {
	runBench(b, `class Point {
		x: number;
		y: number;
		init(x: number, y: number) { this.x = x; this.y = y; }
		shift(dx: number, dy: number) { this.x = this.x + dx; this.y = this.y + dy; }
		sum(): number { return this.x + this.y; }
	}
	var p: Point = Point(0, 0);
	var total: number = 0;
	for (var i: number = 0; i < 20000; i = i + 1) {
		p.shift(1, 2);
		total = total + p.sum();
	}
	print total;`)
}

// Inherited fields occupy the first slots, so a subclass access is the same
// single index as a direct one — no walk up the class chain.
func BenchmarkInheritedFieldAccess(b *testing.B) {
	runBench(b, `class Base {
		a: number;
		init(a: number) { this.a = a; }
	}
	class Mid < Base {
		b: number;
		init(a: number, b: number) { super.init(a); this.b = b; }
	}
	class Leaf < Mid {
		c: number;
		init(a: number, b: number, c: number) { super.init(a, b); this.c = c; }
		total(): number { return this.a + this.b + this.c; }
	}
	var leaf: Leaf = Leaf(1, 2, 3);
	var total: number = 0;
	for (var i: number = 0; i < 20000; i = i + 1) {
		total = total + leaf.total();
	}
	print total;`)
}

// '+' no longer inspects its operands: the checker decided which meaning
// applies and codegen emitted the specific instruction.
func BenchmarkNumberArithmetic(b *testing.B) {
	runBench(b, `var total: number = 0;
	for (var i: number = 0; i < 50000; i = i + 1) {
		total = total + i + 1;
	}
	print total;`)
}

func BenchmarkStringConcat(b *testing.B) {
	runBench(b, `var s: string = "";
	for (var i: number = 0; i < 2000; i = i + 1) {
		s = s + "x";
	}
	print len(s);`)
}

// Conditions are bool, so the jump unwraps rather than coercing.
func BenchmarkBranching(b *testing.B) {
	runBench(b, `var hits: number = 0;
	for (var i: number = 0; i < 50000; i = i + 1) {
		if (i > 25000) { hits = hits + 1; } else { hits = hits + 2; }
	}
	print hits;`)
}

func BenchmarkMethodCalls(b *testing.B) {
	runBench(b, `class Counter {
		n: number;
		init() { this.n = 0; }
		bump(): number { this.n = this.n + 1; return this.n; }
	}
	var c: Counter = Counter();
	for (var i: number = 0; i < 20000; i = i + 1) { c.bump(); }
	print c.n;`)
}
