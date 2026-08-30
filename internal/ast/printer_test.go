package ast_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/parser"
	"github.com/harshagw/viri/internal/scanner"
)

func printSource(t *testing.T, src string) string {
	t.Helper()
	toks, err := scanner.New(bytes.NewBufferString(src), nil).Scan()
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	collector := &objects.DiagnosticCollector{}
	mod, err := parser.NewParser(toks, collector).Parse()
	if err != nil {
		t.Fatalf("parse: %v (%v)", err, collector.Errors)
	}
	return ast.NewPrinter().PrintStatements(mod.GetAllStatements())
}

// The tree is what `--debug` shows in the compiler REPL, so annotations have to
// survive into it — a dump that hides the declared types is not much use while
// building the checker.
func TestPrinterRendersAnnotations(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []string
	}{
		{
			"var declaration",
			`var count: number = 1;`,
			[]string{"VarDecl (count: number)"},
		},
		{
			"const with array type",
			`const names: []string = [];`,
			[]string{"ConstDecl (names: []string)"},
		},
		{
			"function signature",
			`fun add(a: number, b: string): map[string]bool { return {}; }`,
			[]string{"Function (add): map[string]bool", "a: number", "b: string"},
		},
		{
			"void function has no return type",
			`fun log(msg: string) { print msg; }`,
			[]string{"Function (log)", "msg: string"},
		},
		{
			"class fields",
			`class Shape { name: string; sides: []number; area(): number { return 0; } }`,
			[]string{"fields", "name: string", "sides: []number", "methods"},
		},
		{
			"function type and lambda",
			`var f: fun(number): number = fun(n: number): number { return n; };`,
			[]string{"VarDecl (f: fun(number): number)", "Function (anonymous): number", "n: number"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := printSource(t, tc.src)
			for _, want := range tc.want {
				if !strings.Contains(got, want) {
					t.Errorf("tree is missing %q:\n%s", want, got)
				}
			}
		})
	}
}

// A void function must not render a stray ':' with nothing after it.
func TestPrinterOmitsVoidReturnType(t *testing.T) {
	got := printSource(t, `fun log(msg: string) { print msg; }`)
	if strings.Contains(got, "Function (log):") {
		t.Errorf("void function rendered a return type:\n%s", got)
	}
}

// A class with no declared fields should not grow an empty "fields" branch.
func TestPrinterOmitsEmptyFields(t *testing.T) {
	got := printSource(t, `class C { m(): number { return 1; } }`)
	if strings.Contains(got, "fields") {
		t.Errorf("class with no fields rendered a fields branch:\n%s", got)
	}
}
