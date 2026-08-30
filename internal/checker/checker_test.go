package checker_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/harshagw/viri/internal/ast"

	"github.com/harshagw/viri/internal/checker"
	"github.com/harshagw/viri/internal/objects"
	"github.com/harshagw/viri/internal/parser"
	"github.com/harshagw/viri/internal/scanner"
	"github.com/harshagw/viri/internal/types"
)

// check runs the checker over one module of source and returns its
// diagnostics. Programs here are self-contained: cross-module checking is
// covered by the e2e suite, which exercises the real import graph.
func check(t *testing.T, src string) []objects.Diagnostic {
	t.Helper()
	tokens, err := scanner.New(bytes.NewBufferString(src), nil).Scan()
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	collector := &objects.DiagnosticCollector{}
	p := parser.NewParser(tokens, collector)
	p.SetFilePath("test.viri")
	mod, err := p.Parse()
	if err != nil {
		t.Fatalf("parse failed before checking could start: %v (%v)", err, collector.Errors)
	}

	checkCollector := &objects.DiagnosticCollector{}
	checker.New(checkCollector).CheckModule(mod, nil)
	return checkCollector.Errors
}

// wantError asserts the program is rejected with a diagnostic containing want.
func wantError(t *testing.T, src, want string) {
	t.Helper()
	errs := check(t, src)
	if len(errs) == 0 {
		t.Fatalf("expected an error containing %q, but the program checked cleanly\n%s", want, src)
	}
	for _, e := range errs {
		if strings.Contains(e.Message, want) {
			return
		}
	}
	msgs := make([]string, len(errs))
	for i, e := range errs {
		msgs[i] = e.Message
	}
	t.Errorf("want a diagnostic containing %q, got:\n  %s\nfor:\n%s", want, strings.Join(msgs, "\n  "), src)
}

// wantOK asserts the program checks cleanly.
func wantOK(t *testing.T, src string) {
	t.Helper()
	if errs := check(t, src); len(errs) > 0 {
		msgs := make([]string, len(errs))
		for i, e := range errs {
			msgs[i] = e.Message
		}
		t.Errorf("expected no errors, got:\n  %s\nfor:\n%s", strings.Join(msgs, "\n  "), src)
	}
}

func TestAssignability(t *testing.T) {
	t.Run("rejects mismatched initializer", func(t *testing.T) {
		wantError(t, `var x: number = "hello";`, "Cannot assign string to 'x' of type number")
	})
	t.Run("rejects mismatched assignment", func(t *testing.T) {
		wantError(t, `var x: number = 1;
x = "hello";`, "Cannot assign string to 'x' of type number")
	})
	t.Run("requires an initializer", func(t *testing.T) {
		// Every type is non-nilable, so there is no value to default to.
		wantError(t, `var x: number;`, "needs an initializer")
	})
	t.Run("rejects assignment to a constant", func(t *testing.T) {
		wantError(t, `const x: number = 1;
x = 2;`, "Cannot assign to constant 'x'")
	})
	t.Run("rejects an unknown type name", func(t *testing.T) {
		wantError(t, `var x: Widget = 1;`, "Unknown type 'Widget'")
	})
	t.Run("rejects an undefined variable", func(t *testing.T) {
		wantError(t, `var x: number = y;`, "Undefined variable 'y'")
	})
	t.Run("rejects a duplicate declaration", func(t *testing.T) {
		wantError(t, `var x: number = 1;
var x: number = 2;`, "'x' is already declared")
	})
}

// Builtin type names are predeclared identifiers, so the parser accepts them
// as names. The checker owns the builtin scope, so rejecting them is its job —
// which is also why the message can say what is actually wrong.
func TestBuiltinTypeNamesCannotBeRedeclared(t *testing.T) {
	cases := []struct{ name, src, want string }{
		{"variable", `var number: string = "x";`, "Cannot use builtin type name 'number' as a variable name."},
		{"constant", `const string: number = 1;`, "Cannot use builtin type name 'string' as a constant name."},
		{"function", `fun bool(): number { return 1; }`, "Cannot use builtin type name 'bool' as a function name."},
		{"class", `class number { }`, "Cannot use builtin type name 'number' as a class name."},
		{"parameter", `fun f(number: string) { }`, "Cannot use builtin type name 'number' as a parameter name."},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { wantError(t, tc.src, tc.want) })
	}
}

func TestOperators(t *testing.T) {
	t.Run("plus rejects mixed operands", func(t *testing.T) {
		wantError(t, `var x: number = 1 + "a";`, "'+' needs two numbers or two strings")
	})
	t.Run("plus concatenates strings", func(t *testing.T) {
		wantOK(t, `var x: string = "a" + "b";`)
	})
	t.Run("minus needs numbers", func(t *testing.T) {
		wantError(t, `var x: number = "a" - 1;`, "'-' needs two numbers")
	})
	t.Run("comparison needs numbers", func(t *testing.T) {
		wantError(t, `var x: bool = "a" < 1;`, "'<' needs two numbers")
	})
	t.Run("equality needs related operands", func(t *testing.T) {
		wantError(t, `var x: bool = 1 == "a";`, "'==' needs operands of the same type")
	})
	t.Run("unary minus needs a number", func(t *testing.T) {
		wantError(t, `var x: number = -"a";`, "Unary '-' needs a number")
	})
	t.Run("bang needs a bool", func(t *testing.T) {
		wantError(t, `var x: bool = !1;`, "'!' needs a bool")
	})
	t.Run("and needs bools", func(t *testing.T) {
		wantError(t, `var x: bool = 1 and true;`, "needs bool operands")
	})
}

// There is no truthiness: control flow branches on a real bool. This is what
// lets the VM drop its IsTruthy call in Phase 4.
func TestConditionsMustBeBool(t *testing.T) {
	cases := []struct{ name, src string }{
		{"if", `if (1) { print "x"; }`},
		{"while", `while ("a") { break; }`},
		{"for", `for (var i: number = 0; i; i = i + 1) { break; }`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) { wantError(t, tc.src, "condition must be bool") })
	}
}

func TestCalls(t *testing.T) {
	t.Run("rejects too few arguments", func(t *testing.T) {
		wantError(t, `fun add(a: number, b: number): number { return a + b; }
var x: number = add(1);`, "Expected 2 arguments but got 1")
	})
	t.Run("rejects a wrong argument type", func(t *testing.T) {
		wantError(t, `fun add(a: number, b: number): number { return a + b; }
var x: number = add(1, "two");`, "Argument 2 is string, expected number")
	})
	t.Run("rejects calling a non-function", func(t *testing.T) {
		wantError(t, `var x: number = 1;
var y: number = x();`, "it is not a function")
	})
	t.Run("accepts a correct call", func(t *testing.T) {
		wantOK(t, `fun add(a: number, b: number): number { return a + b; }
var x: number = add(1, 2);`)
	})
	t.Run("checks calls through a function value", func(t *testing.T) {
		wantError(t, `var f: fun(number): number = fun(n: number): number { return n; };
var x: number = f("a");`, "Argument 1 is string, expected number")
	})
}

func TestReturns(t *testing.T) {
	t.Run("rejects a wrong return type", func(t *testing.T) {
		wantError(t, `fun f(): number { return "a"; }`, "Cannot return string from a function returning number")
	})
	t.Run("rejects a bare return from a value function", func(t *testing.T) {
		wantError(t, `fun f(): number { return; }`, "This function must return number")
	})
	t.Run("rejects returning a value from a void function", func(t *testing.T) {
		wantError(t, `fun f() { return 1; }`, "returns nothing, so it cannot return a value")
	})
	t.Run("accepts a bare return from a void function", func(t *testing.T) {
		wantOK(t, `fun f() { return; }`)
	})
}

// A call to a function that returns nothing produces no value, so it is legal
// only as a statement.
func TestVoidIsNotAValue(t *testing.T) {
	t.Run("rejects binding a void call", func(t *testing.T) {
		wantError(t, `fun f() { }
var x: number = f();`, "Cannot assign void to 'x'")
	})
	t.Run("rejects a void argument", func(t *testing.T) {
		wantError(t, `fun f() { }
fun g(n: number) { }
g(f());`, "Argument 1 is void")
	})
	t.Run("accepts a void call as a statement", func(t *testing.T) {
		wantOK(t, `fun f() { }
f();`)
	})
}

func TestCollections(t *testing.T) {
	t.Run("annotation types an empty array", func(t *testing.T) {
		wantOK(t, `var xs: []number = [];`)
	})
	t.Run("annotation types an empty map", func(t *testing.T) {
		wantOK(t, `var m: map[string]number = {};`)
	})
	t.Run("rejects a heterogeneous array", func(t *testing.T) {
		wantError(t, `var xs: []number = [1, "two"];`, "Cannot store string in []number")
	})
	t.Run("rejects a heterogeneous literal with no annotation to guide it", func(t *testing.T) {
		wantError(t, `fun f(xs: []number) { }
f([1, "two"]);`, "Cannot store string in []number")
	})
	t.Run("rejects a non-string map key", func(t *testing.T) {
		wantError(t, `var m: map[number]string = {};`, "Map keys must be string")
	})
	t.Run("rejects a wrong map value", func(t *testing.T) {
		wantError(t, `var m: map[string]number = {"a": "b"};`, "Cannot store string in map[string]number")
	})
	t.Run("rejects a non-number array index", func(t *testing.T) {
		wantError(t, `var xs: []number = [1];
var x: number = xs["a"];`, "Array indices must be number")
	})
	t.Run("rejects a non-string map index", func(t *testing.T) {
		wantError(t, `var m: map[string]number = {};
var x: number = m[1];`, "Map indices must be string")
	})
	t.Run("rejects indexing a non-collection", func(t *testing.T) {
		wantError(t, `var s: string = "abc";
var c: string = s[0];`, "Cannot index string")
	})
	t.Run("rejects storing a wrong element type", func(t *testing.T) {
		wantError(t, `var xs: []number = [1];
xs[0] = "a";`, "Cannot store string in a collection of number")
	})
}

func TestClasses(t *testing.T) {
	t.Run("rejects an undeclared field assignment", func(t *testing.T) {
		// Dynamic field creation is gone.
		wantError(t, `class P { x: number; init(x: number) { this.x = x; } }
var p: P = P(1);
p.y = 2;`, "has no field 'y'")
	})
	t.Run("rejects an unknown property read", func(t *testing.T) {
		wantError(t, `class P { x: number; init(x: number) { this.x = x; } }
var p: P = P(1);
var n: number = p.z;`, "has no field or method 'z'")
	})
	t.Run("rejects a wrong field type", func(t *testing.T) {
		wantError(t, `class P { x: number; init() { this.x = "a"; } }`,
			"Cannot assign string to field 'x' of type number")
	})
	t.Run("rejects wrong constructor arguments", func(t *testing.T) {
		wantError(t, `class P { x: number; init(x: number) { this.x = x; } }
var p: P = P("a");`, "Argument 1 is string, expected number")
	})
	t.Run("rejects a class used as a value where an instance is expected", func(t *testing.T) {
		// A class name in value position is its constructor, not an instance.
		wantError(t, `class P { }
var p: P = P;`, "Cannot assign fun(): P to 'p' of type P")
	})
	t.Run("rejects this outside a class", func(t *testing.T) {
		wantError(t, `var x: number = this.y;`, "'this' can only be used inside a class")
	})
	t.Run("rejects a duplicate field", func(t *testing.T) {
		wantError(t, `class P { x: number; x: string; }`, "Field 'x' is already declared")
	})
	t.Run("rejects a return type on init", func(t *testing.T) {
		wantError(t, `class P { init(): number { return 1; } }`, "'init' cannot declare a return type")
	})
	t.Run("accepts a well-typed class", func(t *testing.T) {
		wantOK(t, `class P {
			x: number;
			init(x: number) { this.x = x; }
			double(): number { return this.x * 2; }
		}
		var p: P = P(3);
		print p.double();`)
	})
}

func TestInheritance(t *testing.T) {
	const shapes = `class Shape {
		name: string;
		init(name: string) { this.name = name; }
		area(): number { return 0; }
	}
	`

	t.Run("a subclass is assignable to its superclass", func(t *testing.T) {
		wantOK(t, shapes+`class Square < Shape {
			side: number;
			init(side: number) { super.init("square"); this.side = side; }
			area(): number { return this.side * this.side; }
		}
		var s: Shape = Square(4);
		print s.area();`)
	})
	t.Run("a superclass is not assignable to a subclass", func(t *testing.T) {
		wantError(t, shapes+`class Square < Shape {
			side: number;
			init(side: number) { super.init("sq"); this.side = side; }
		}
		var sq: Square = Shape("s");`, "Cannot assign Shape to 'sq' of type Square")
	})
	t.Run("rejects an override with a different signature", func(t *testing.T) {
		wantError(t, shapes+`class Square < Shape {
			area(): string { return "big"; }
		}`, "does not match the one it overrides")
	})
	t.Run("rejects redeclaring an inherited field", func(t *testing.T) {
		wantError(t, shapes+`class Square < Shape {
			name: string;
		}`, "already declared on superclass 'Shape'")
	})
	t.Run("a subclass sees inherited fields and methods", func(t *testing.T) {
		wantOK(t, shapes+`class Square < Shape {
			side: number;
			init(side: number) { super.init("square"); this.side = side; }
			describe(): string { return this.name; }
		}
		var sq: Square = Square(2);
		print sq.area();`)
	})
	t.Run("rejects super in a class with no superclass", func(t *testing.T) {
		wantError(t, `class A { m() { super.m(); } }`, "class with a superclass")
	})
	t.Run("rejects an unknown superclass method", func(t *testing.T) {
		wantError(t, shapes+`class Square < Shape {
			m() { super.nope(); }
		}`, "has no method 'nope'")
	})
	t.Run("rejects self-inheritance", func(t *testing.T) {
		wantError(t, `class A < A { }`, "cannot inherit from itself")
	})
}

func TestNativeFunctions(t *testing.T) {
	t.Run("checks clock arity", func(t *testing.T) {
		wantError(t, `var t: number = clock(1);`, "Expected 0 arguments but got 1")
	})
	t.Run("len accepts a string, array, or map", func(t *testing.T) {
		wantOK(t, `var a: number = len("abc");
var b: number = len([1, 2]);
var c: number = len({"k": 1});`)
	})
	t.Run("len rejects a number", func(t *testing.T) {
		wantError(t, `var n: number = len(1);`, "len needs a string, array or map")
	})
}

// One bad expression should produce one diagnostic, not a cascade of unrelated
// complaints about everything downstream of it.
func TestErrorsDoNotCascade(t *testing.T) {
	errs := check(t, `var x: number = "a" - 1;
var y: number = x + 1;
var z: number = x * 2;`)
	if len(errs) != 1 {
		msgs := make([]string, len(errs))
		for i, e := range errs {
			msgs[i] = e.Message
		}
		t.Errorf("expected exactly 1 diagnostic, got %d:\n  %s", len(errs), strings.Join(msgs, "\n  "))
	}
}

// Diagnostics must carry a usable source position.
func TestDiagnosticsCarryAPosition(t *testing.T) {
	errs := check(t, "var a: number = 1;\nvar b: number = \"two\";\n")
	if len(errs) == 0 {
		t.Fatal("expected an error")
	}
	if errs[0].Token.Line != 2 {
		t.Errorf("error reported on line %d, want 2", errs[0].Token.Line)
	}
}

// Signatures are collected before any body is checked, so declarations may
// refer to one another in any order.
func TestForwardReferences(t *testing.T) {
	t.Run("mutual recursion", func(t *testing.T) {
		wantOK(t, `fun isEven(n: number): bool {
			if (n == 0) { return true; }
			return isOdd(n - 1);
		}
		fun isOdd(n: number): bool {
			if (n == 0) { return false; }
			return isEven(n - 1);
		}
		print isEven(10);`)
	})
	t.Run("a function may use a global declared later", func(t *testing.T) {
		wantOK(t, `fun later(): number { return declaredLater * 2; }
var declaredLater: number = 21;
print later();`)
	})
	t.Run("a class may refer to one declared later", func(t *testing.T) {
		wantOK(t, `class A { b: B; init(b: B) { this.b = b; } }
class B { n: number; init(n: number) { this.n = n; } }`)
	})
}

// The side table is the point of the pass: Phase 3 and 4 read it.
func TestExprTypesArePopulated(t *testing.T) {
	tokens, err := scanner.New(bytes.NewBufferString(`var x: number = 1 + 2;`), nil).Scan()
	if err != nil {
		t.Fatal(err)
	}
	p := parser.NewParser(tokens, &objects.DiagnosticCollector{})
	mod, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	ck := checker.New(&objects.DiagnosticCollector{})
	if _, ok := ck.CheckModule(mod, nil); !ok {
		t.Fatal("expected the program to check")
	}

	found := 0
	for _, typ := range ck.ExprTypes() {
		if typ == types.Number {
			found++
		}
	}
	// 1, 2, and (1 + 2)
	if found != 3 {
		t.Errorf("expected 3 number-typed expressions in the side table, got %d", found)
	}
}

// Definite assignment is what makes "no implicit nil" true for fields: a field
// typed string must never be readable before something put a string in it.
func TestDefiniteAssignment(t *testing.T) {
	t.Run("rejects a field init never assigns", func(t *testing.T) {
		wantError(t, `class P { x: number; y: number; init(x: number) { this.x = x; } }`,
			"'init' must assign 'y' on every path")
	})
	t.Run("rejects a class with fields and no init", func(t *testing.T) {
		wantError(t, `class P { x: number; }`, "has no 'init' to assign it")
	})
	t.Run("accepts a class with no fields and no init", func(t *testing.T) {
		wantOK(t, `class P { m(): number { return 1; } }`)
	})
	t.Run("accepts assignment on both branches", func(t *testing.T) {
		wantOK(t, `class P {
			x: number;
			init(c: bool) { if (c) { this.x = 1; } else { this.x = 2; } }
		}`)
	})
	t.Run("rejects assignment on only one branch", func(t *testing.T) {
		wantError(t, `class P {
			x: number;
			init(c: bool) { if (c) { this.x = 1; } }
		}`, "must assign 'x' on every path")
	})
	t.Run("rejects assignment only inside a loop", func(t *testing.T) {
		// The body may run zero times.
		wantError(t, `class P {
			x: number;
			init(c: bool) { while (c) { this.x = 1; } }
		}`, "must assign 'x' on every path")
	})
	t.Run("an early return does not weaken the other path", func(t *testing.T) {
		wantOK(t, `class P {
			x: number;
			init(c: bool) { if (c) { return; } this.x = 1; }
		}`)
	})
	t.Run("accepts assignment inside a nested block", func(t *testing.T) {
		wantOK(t, `class P { x: number; init() { { this.x = 1; } } }`)
	})
}

func TestSuperInitIsRequired(t *testing.T) {
	const base = `class Base { n: number; init(n: number) { this.n = n; } }
	`
	t.Run("rejects a subclass init that omits super.init", func(t *testing.T) {
		wantError(t, base+`class Sub < Base {
			s: number;
			init(s: number) { this.s = s; }
		}`, "must call super.init(...) on every path")
	})
	t.Run("accepts a subclass init that calls super.init", func(t *testing.T) {
		wantOK(t, base+`class Sub < Base {
			s: number;
			init(s: number) { super.init(1); this.s = s; }
		}`)
	})
	t.Run("a subclass with no init inherits its superclass's", func(t *testing.T) {
		// Base's init already assigns Base's fields, so there is nothing for
		// Sub to do — and the VM looks up the inherited init the same way.
		wantOK(t, base+`class Sub < Base { }`)
	})
	t.Run("but a subclass with its own fields still needs an init", func(t *testing.T) {
		wantError(t, base+`class Sub < Base { s: number; }`, "has no 'init' to assign it")
	})
	t.Run("rejects super.init on only one branch", func(t *testing.T) {
		wantError(t, base+`class Sub < Base {
			s: number;
			init(c: bool, s: number) { if (c) { super.init(1); } this.s = s; }
		}`, "must call super.init(...) on every path")
	})
	t.Run("a superclass with no fields needs no super.init", func(t *testing.T) {
		wantOK(t, `class Base { m(): number { return 1; } }
		class Sub < Base { s: number; init(s: number) { this.s = s; } }`)
	})
}

// The REPL checks one line at a time against a scope the earlier lines built.
// Codegen depends on this: field access is resolved to slots, so an unchecked
// line would emit instructions the VM cannot run.
func TestCheckIncremental(t *testing.T) {
	parseLine := func(t *testing.T, src string) []ast.Stmt {
		t.Helper()
		tokens, err := scanner.New(bytes.NewBufferString(src), nil).Scan()
		if err != nil {
			t.Fatalf("scan %q: %v", src, err)
		}
		mod, err := parser.NewParser(tokens, &objects.DiagnosticCollector{}).Parse()
		if err != nil {
			t.Fatalf("parse %q: %v", src, err)
		}
		return mod.GetAllStatements()
	}

	collector := &objects.DiagnosticCollector{}
	ck := checker.New(collector)

	lines := []struct {
		src     string
		wantOK  bool
		comment string
	}{
		{`var x: number = 1;`, true, "declares x"},
		{`print x + 1;`, true, "x is still in scope on the next line"},
		{`var y: number = "a";`, false, "type errors are still caught"},
		{`class P { n: number; init(n: number) { this.n = n; } }`, true, "declares a class"},
		{`var p: P = P(7);`, true, "the class is usable on a later line"},
		{`print p.n;`, true, "its fields are known"},
		{`print p.missing;`, false, "unknown fields are still rejected"},
	}

	for _, line := range lines {
		if got := ck.CheckIncremental(parseLine(t, line.src)); got != line.wantOK {
			t.Errorf("%s: CheckIncremental(%q) = %v, want %v", line.comment, line.src, got, line.wantOK)
		}
	}

	// Field slots must be resolvable across lines, which is what lets codegen
	// emit OpGetField for `p.n` typed on an earlier line.
	classes := ck.ClassTypes()
	if len(classes) != 1 {
		t.Fatalf("expected 1 class type, got %d", len(classes))
	}
	for _, class := range classes {
		if slot, ok := class.FieldSlot("n"); !ok || slot != 0 {
			t.Errorf("field 'n' slot = (%d, %v), want (0, true)", slot, ok)
		}
	}
}

// Reading a field before it is assigned is the one way nil could still be
// observed: an unassigned slot holds no value at all. Every case here printed
// "nil" before these checks existed.
func TestNoReadBeforeAssignment(t *testing.T) {
	t.Run("own field", func(t *testing.T) {
		wantError(t, `class P { x: number; init() { print this.x; this.x = 1; } }`,
			"Cannot read 'this.x' before it is assigned")
	})
	t.Run("assigning a field to itself", func(t *testing.T) {
		wantError(t, `class P { x: number; init() { this.x = this.x; } }`,
			"Cannot read 'this.x' before it is assigned")
	})
	t.Run("inside a for body", func(t *testing.T) {
		// The body may not run, so it assigns nothing — but if it does run,
		// its reads happen, so they still have to be checked.
		wantError(t, `class P {
			x: number;
			init() { for (var i: number = 0; i < 1; i = i + 1) { print this.x; } this.x = 1; }
		}`, "Cannot read 'this.x' before it is assigned")
	})
	t.Run("inside a while body", func(t *testing.T) {
		wantError(t, `class P {
			x: number;
			init() { while (false) { print this.x; } this.x = 1; }
		}`, "Cannot read 'this.x' before it is assigned")
	})
	t.Run("inside a lambda that closes over this", func(t *testing.T) {
		wantError(t, `class P {
			x: number;
			init() { var f: fun(): number = fun(): number { return this.x; }; print f(); this.x = 1; }
		}`, "Cannot read 'this.x' before it is assigned")
	})
	t.Run("calling a method that may read a field", func(t *testing.T) {
		wantError(t, `class P {
			x: number;
			init() { this.show(); this.x = 1; }
			show() { print this.x; }
		}`, "Cannot call 'this.show' before 'x' is assigned")
	})
	t.Run("passing this out before it is built", func(t *testing.T) {
		wantError(t, `fun take(p: P) { print p.x; }
		class P { x: number; init() { take(this); this.x = 1; } }`,
			"Cannot pass 'this' before 'x' is assigned")
	})
	t.Run("inherited field before super.init", func(t *testing.T) {
		wantError(t, `class A { a: number; init(a: number) { this.a = a; } }
		class B < A { b: number; init() { print this.a; super.init(1); this.b = 2; } }`,
			"before super.init(...) assigns it")
	})
	t.Run("superclass method before super.init", func(t *testing.T) {
		wantError(t, `class A { a: number; init(a: number) { this.a = a; } get(): number { return this.a; } }
		class B < A { b: number; init() { print super.get(); super.init(1); this.b = 2; } }`,
			"before super.init(...)")
	})
}

// The read checks must not reject ordinary constructors.
func TestReadsAfterAssignmentAreFine(t *testing.T) {
	t.Run("reading a field once assigned", func(t *testing.T) {
		wantOK(t, `class P {
			x: number;
			y: number;
			init() { this.x = 1; this.y = this.x + 1; }
		}`)
	})
	t.Run("calling a method once every field is assigned", func(t *testing.T) {
		wantOK(t, `class P {
			x: number;
			init() { this.x = 1; this.show(); }
			show() { print this.x; }
		}`)
	})
	t.Run("looping over a field after assigning it", func(t *testing.T) {
		wantOK(t, `class P {
			x: number;
			init() { this.x = 0; for (var i: number = 0; i < 3; i = i + 1) { this.x = this.x + i; } }
		}`)
	})
	t.Run("a lambda not touching this", func(t *testing.T) {
		wantOK(t, `class P {
			x: number;
			init() { var f: fun(number): number = fun(v: number): number { return v * 2; }; this.x = f(21); }
		}`)
	})
	t.Run("reads outside init need no analysis", func(t *testing.T) {
		// Once init returns, definite assignment guarantees every field holds
		// a value, so no later read can be too early.
		wantOK(t, `class P {
			x: number;
			init() { this.x = 1; }
			get(): number { return this.x; }
		}`)
	})
}

// A module the checker cannot resolve must not leave a nil in the import map:
// dereferencing one crashed the compiler rather than reporting anything.
func TestUnresolvedImportDoesNotCrash(t *testing.T) {
	t.Run("qualified read from an unbound alias", func(t *testing.T) {
		wantError(t, `var x: number = missing.thing();`, "Undefined variable 'missing'")
	})
	t.Run("qualified type from an unbound alias", func(t *testing.T) {
		wantError(t, `var x: missing.Thing = 1;`, "Unknown module 'missing'")
	})
}

// Loop placement and export placement moved out of the compiler in Phase 4.
// They are user-facing rules about a program that parses, so the checker owns
// them; the compiler keeps only an invariant guard, since it needs a jump
// target and a module scope to emit against.
func TestLoopPlacement(t *testing.T) {
	t.Run("rejects break outside a loop", func(t *testing.T) {
		wantError(t, `break;`, "'break' can only be used inside a loop")
	})
	t.Run("rejects continue outside a loop", func(t *testing.T) {
		wantError(t, `continue;`, "'continue' can only be used inside a loop")
	})
	t.Run("rejects break in a function nested in a loop", func(t *testing.T) {
		// The loop does not reach into the function, so there is nothing to
		// break out of.
		wantError(t, `while (true) {
  fun g() { break; }
}`, "'break' can only be used inside a loop")
	})
	t.Run("accepts break and continue inside their loop", func(t *testing.T) {
		wantOK(t, `while (true) { break; }
for (var i: number = 0; i < 3; i = i + 1) { continue; }`)
	})
	t.Run("accepts break inside a nested loop", func(t *testing.T) {
		wantOK(t, `while (true) {
  for (var i: number = 0; i < 3; i = i + 1) { break; }
  break;
}`)
	})
	t.Run("rejects break after its loop closes", func(t *testing.T) {
		wantError(t, `while (true) { print 1; }
break;`, "'break' can only be used inside a loop")
	})
}

func TestExportPlacement(t *testing.T) {
	t.Run("rejects an exported local variable", func(t *testing.T) {
		wantError(t, `fun f() { export var a: number = 1; }`,
			"'a' cannot be exported")
	})
	t.Run("rejects an exported nested function", func(t *testing.T) {
		wantError(t, `fun f() { export fun g() { print 1; } }`,
			"'g' cannot be exported")
	})
	t.Run("rejects an exported class inside a block", func(t *testing.T) {
		wantError(t, `{ export class C {} }`, "'C' cannot be exported")
	})
	t.Run("accepts exports at module level", func(t *testing.T) {
		wantOK(t, `export var a: number = 1;
export fun f() { print a; }
export class C {}`)
	})
}
