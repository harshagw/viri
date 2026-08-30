package main

import (
	"bytes"
	"strings"
	"testing"
)

// run drives the playground exactly as the WASM entry point does.
func run(src string) (result string, out string, errs []string, warns []string) {
	handler := &playgroundHandler{errors: []string{}, warnings: []string{}}
	var buf bytes.Buffer
	result = execute(src, &buf, handler)
	return result, buf.String(), handler.errors, handler.warnings
}

func TestPlaygroundRunsPrograms(t *testing.T) {
	tests := []struct {
		name, src, wantOut string
	}{
		{"print", `print 1 + 2;`, "3\n"},
		{"strings", `print "a" + "b";`, "ab\n"},
		{
			"functions",
			`fun add(a: number, b: number): number { return a + b; }
			print add(2, 3);`,
			"5\n",
		},
		{
			"classes",
			`class P {
				x: number;
				init(x: number) { this.x = x; }
				double(): number { return this.x * 2; }
			}
			print P(21).double();`,
			"42\n",
		},
		{
			"inheritance",
			`class A { a: number; init(a: number) { this.a = a; } }
			class B < A { b: number; init(a: number, b: number) { super.init(a); this.b = b; } }
			var x: B = B(1, 2);
			print x.a + x.b;`,
			"3\n",
		},
		{
			"loops and closures",
			`fun counter(): fun(): number {
				var n: number = 0;
				fun bump(): number { n = n + 1; return n; }
				return bump;
			}
			var c: fun(): number = counter();
			c();
			print c();`,
			"2\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, out, errs, _ := run(tc.src)
			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}
			if out != tc.wantOut {
				t.Errorf("output = %q, want %q", out, tc.wantOut)
			}
		})
	}
}

// The stdlib is reachable, and its calls are type-checked like any other.
func TestPlaygroundStdlib(t *testing.T) {
	_, out, errs, _ := run(`import "std:math" as math;
	print math.sqrt(16);`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if !strings.Contains(out, "4") {
		t.Errorf("output = %q, want it to contain 4", out)
	}
}

// Type errors must come back as diagnostics, not as a crash or partial output.
func TestPlaygroundReportsTypeErrors(t *testing.T) {
	tests := []struct{ name, src, want string }{
		{"assignment", `var x: number = "hello";`, "Cannot assign string"},
		{"missing annotation", `var x = 1;`, "Expect ':' and a type"},
		{"bad operands", `print 1 + "a";`, "needs two numbers or two strings"},
		{"non-bool condition", `if (1) { print "x"; }`, "condition must be bool"},
		{"unknown field", `class P { } var p: P = P(); print p.q;`, "no field or method 'q'"},
		{"unassigned field", `class P { x: number; init() { } }`, "must assign 'x'"},
		{"nil is not a value", `var x: number = nil;`, "'nil' is not a value"},
		{"arity", `fun f(a: number) { } f();`, "Expected 1 arguments but got 0"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, out, errs, _ := run(tc.src)
			if len(errs) == 0 {
				t.Fatalf("expected an error containing %q, got none", tc.want)
			}
			if !strings.Contains(strings.Join(errs, "\n"), tc.want) {
				t.Errorf("errors = %v, want one containing %q", errs, tc.want)
			}
			if out != "" {
				t.Errorf("a rejected program produced output %q; nothing should have run", out)
			}
		})
	}
}

// Diagnostics carry the line so the editor can point at it.
func TestPlaygroundErrorsCarryLines(t *testing.T) {
	_, _, errs, _ := run("var a: number = 1;\nvar b: number = \"two\";\n")
	if len(errs) == 0 {
		t.Fatal("expected an error")
	}
	if !strings.HasPrefix(errs[0], "Line 2:") {
		t.Errorf("error = %q, want it to start with \"Line 2:\"", errs[0])
	}
}

// A program that ends in a bare expression shows its value, the way the old
// interpreter-backed playground did.
func TestPlaygroundFinalValue(t *testing.T) {
	t.Run("trailing expression", func(t *testing.T) {
		result, _, errs, _ := run(`var x: number = 20;
		x + 22;`)
		if len(errs) > 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if result != "42" {
			t.Errorf("result = %q, want %q", result, "42")
		}
	})

	t.Run("a print is not a result", func(t *testing.T) {
		// print pops too, so the last popped value is not a result unless the
		// final statement was an expression.
		result, out, _, _ := run(`print 42;`)
		if result != "" {
			t.Errorf("result = %q, want empty", result)
		}
		if out != "42\n" {
			t.Errorf("output = %q, want %q", out, "42\n")
		}
	})

	t.Run("a void call is not a result", func(t *testing.T) {
		result, _, errs, _ := run(`fun f() { } f();`)
		if len(errs) > 0 {
			t.Fatalf("unexpected errors: %v", errs)
		}
		if result != "" {
			t.Errorf("result = %q, want empty", result)
		}
	})

	t.Run("a declaration is not a result", func(t *testing.T) {
		result, _, _, _ := run(`var x: number = 1;`)
		if result != "" {
			t.Errorf("result = %q, want empty", result)
		}
	})
}

// Runtime errors still reach the user, with a line.
func TestPlaygroundRuntimeErrors(t *testing.T) {
	_, _, errs, _ := run(`print 1 / 0;`)
	if len(errs) == 0 {
		t.Fatal("expected a runtime error")
	}
	if !strings.Contains(errs[0], "Division by zero") {
		t.Errorf("errors = %v, want one about division by zero", errs)
	}
}

// std: modules are built in and work here. A file import has nowhere to
// resolve to — there is no filesystem — so it is reported rather than silently
// ignored, which is what this used to do.
func TestPlaygroundRejectsFileImports(t *testing.T) {
	_, _, errs, _ := run(`import "other.viri" as other;
	print 1;`)
	if len(errs) == 0 {
		t.Fatal("expected an error for a file import")
	}
	if !strings.Contains(errs[0], "Only std: imports are available") {
		t.Errorf("errors = %v, want one about imports", errs)
	}
}

func TestPlaygroundRejectsUnknownStdlibModule(t *testing.T) {
	_, _, errs, _ := run(`import "std:nope" as n;
	print 1;`)
	if len(errs) == 0 {
		t.Fatal("expected an error for an unknown std: module")
	}
	if !strings.Contains(errs[0], "Unknown standard library module") {
		t.Errorf("errors = %v, want one about an unknown module", errs)
	}
}

func TestPlaygroundEmptyInput(t *testing.T) {
	result, out, errs, _ := run("")
	if result != "" || out != "" || len(errs) != 0 {
		t.Errorf("empty input produced result=%q out=%q errs=%v", result, out, errs)
	}
}

// A parse error must not be followed by checker noise about the same code.
func TestPlaygroundStopsAtParseErrors(t *testing.T) {
	_, out, errs, _ := run(`var x: number = ;`)
	if len(errs) == 0 {
		t.Fatal("expected a parse error")
	}
	if out != "" {
		t.Errorf("output = %q, want empty", out)
	}
}
