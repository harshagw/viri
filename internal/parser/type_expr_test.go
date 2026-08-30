package parser

import (
	"strings"
	"testing"

	"github.com/harshagw/viri/internal/ast"
)

// typeOfFirstVar parses a program whose first statement is a var declaration
// and returns the declared type, rendered back to source form.
func typeOfFirstVar(t *testing.T, src string) string {
	t.Helper()
	mod := parseSource(t, src)
	decl, ok := mod.Statements[0].(*ast.VarDeclStmt)
	if !ok {
		t.Fatalf("expected a VarDeclStmt, got %T", mod.Statements[0])
	}
	if decl.Type == nil {
		t.Fatal("declared type is nil; annotations are mandatory")
	}
	return ast.TypeExprString(decl.Type)
}

func TestParseTypeForms(t *testing.T) {
	tests := []struct{ src, want string }{
		// builtins are ordinary identifiers to the parser
		{`var x: number = 1;`, "number"},
		{`var x: string = "a";`, "string"},
		{`var x: bool = true;`, "bool"},
		// a class is spelled exactly like a builtin
		{`var x: Shape = Shape();`, "Shape"},
		// a class imported from another module
		{`var x: shapes.Circle = shapes.Circle();`, "shapes.Circle"},
		// arrays
		{`var x: []number = [];`, "[]number"},
		{`var x: [][]string = [];`, "[][]string"},
		{`var x: []Shape = [];`, "[]Shape"},
		// maps
		{`var x: map[string]number = {};`, "map[string]number"},
		{`var x: map[string][]bool = {};`, "map[string][]bool"},
		{`var x: map[string]map[string]number = {};`, "map[string]map[string]number"},
		// function types
		{`var x: fun() = f;`, "fun()"},
		{`var x: fun(): number = f;`, "fun(): number"},
		{`var x: fun(number) = f;`, "fun(number)"},
		{`var x: fun(number, string): bool = f;`, "fun(number, string): bool"},
		{`var x: fun(fun(number): number): number = f;`, "fun(fun(number): number): number"},
		// the forms compose
		{`var x: []fun(number): number = [];`, "[]fun(number): number"},
		{`var x: map[string]fun(): []number = {};`, "map[string]fun(): []number"},
		{`var x: fun([]number, map[string]bool): []Shape = f;`, "fun([]number, map[string]bool): []Shape"},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			if got := typeOfFirstVar(t, tc.src); got != tc.want {
				t.Errorf("parsed %q as %q, want %q", tc.src, got, tc.want)
			}
		})
	}
}

// Builtin type names are predeclared identifiers, not keywords, so the parser
// happily accepts them as ordinary names. Rejecting that is the checker's job
// (Phase 2) — it owns the builtin scope, so it can give a real message. What
// matters here is that the parser does not need to know the builtin list.
func TestBuiltinTypeNamesAreOrdinaryIdentifiers(t *testing.T) {
	mod := parseSource(t, `var number: string = "x";`)
	decl := mod.Statements[0].(*ast.VarDeclStmt)
	if decl.Name.Lexeme != "number" {
		t.Errorf("variable name = %q, want %q", decl.Name.Lexeme, "number")
	}
	if got := ast.TypeExprString(decl.Type); got != "string" {
		t.Errorf("declared type = %q, want %q", got, "string")
	}
}

// 'map' is the one word this phase reserves, because it is structural syntax
// like 'fun' rather than a name the checker could resolve.
func TestMapIsReserved(t *testing.T) {
	_, errs := parseSourceExpectingErrors(t, `var map: number = 1;`)
	if len(errs) == 0 {
		t.Fatal("expected 'map' to be rejected as a variable name")
	}
}

func TestFunctionSignatureParsing(t *testing.T) {
	mod := parseSource(t, `fun add(a: number, b: []string): map[string]bool { return {}; }`)
	fn := mod.Statements[0].(*ast.FunctionStmt)

	if len(fn.Params) != 2 {
		t.Fatalf("params = %d, want 2", len(fn.Params))
	}
	if got := ast.TypeExprString(fn.Params[0].Type); got != "number" {
		t.Errorf("param 0 type = %q, want %q", got, "number")
	}
	if got := ast.TypeExprString(fn.Params[1].Type); got != "[]string" {
		t.Errorf("param 1 type = %q, want %q", got, "[]string")
	}
	if got := ast.TypeExprString(fn.ReturnType); got != "map[string]bool" {
		t.Errorf("return type = %q, want %q", got, "map[string]bool")
	}
}

// An omitted return type means the function returns nothing. This is
// unambiguous because a body always starts with '{' and no type does.
func TestOmittedReturnTypeIsVoid(t *testing.T) {
	mod := parseSource(t, `fun log(msg: string) { print msg; }`)
	fn := mod.Statements[0].(*ast.FunctionStmt)
	if fn.ReturnType != nil {
		t.Errorf("return type = %q, want nil (void)", ast.TypeExprString(fn.ReturnType))
	}
}

func TestParseClassFields(t *testing.T) {
	mod := parseSource(t, `
		class Shape {
			name: string;
			sides: []number;
			init(name: string) { this.name = name; }
			area(): number { return 0; }
		}
	`)
	class := mod.Statements[0].(*ast.ClassStmt)

	if len(class.Fields) != 2 {
		t.Fatalf("fields = %d, want 2", len(class.Fields))
	}
	if class.Fields[0].Name.Lexeme != "name" || ast.TypeExprString(class.Fields[0].Type) != "string" {
		t.Errorf("field 0 = %s: %s", class.Fields[0].Name.Lexeme, ast.TypeExprString(class.Fields[0].Type))
	}
	if class.Fields[1].Name.Lexeme != "sides" || ast.TypeExprString(class.Fields[1].Type) != "[]number" {
		t.Errorf("field 1 = %s: %s", class.Fields[1].Name.Lexeme, ast.TypeExprString(class.Fields[1].Type))
	}
	if len(class.Methods) != 2 {
		t.Errorf("methods = %d, want 2", len(class.Methods))
	}
}

// A class member is a field when its name is followed by ':' and a method when
// it is followed by '('. Interleaving them must not confuse the parser.
func TestClassFieldsAndMethodsInterleave(t *testing.T) {
	mod := parseSource(t, `
		class C {
			a: number;
			m1() {}
			b: string;
			m2(): bool { return true; }
			c: []C;
		}
	`)
	class := mod.Statements[0].(*ast.ClassStmt)
	if len(class.Fields) != 3 {
		t.Errorf("fields = %d, want 3", len(class.Fields))
	}
	if len(class.Methods) != 2 {
		t.Errorf("methods = %d, want 2", len(class.Methods))
	}
}

// Annotation colons only ever follow a declaration name or ')', so they never
// collide with the ':' in a hash literal.
func TestHashLiteralColonIsUnambiguous(t *testing.T) {
	mod := parseSource(t, `var m: map[string]number = {"a": 1, "b": 2};`)
	decl := mod.Statements[0].(*ast.VarDeclStmt)
	hash, ok := decl.Initializer.(*ast.HashLiteralExpr)
	if !ok {
		t.Fatalf("expected a HashLiteralExpr, got %T", decl.Initializer)
	}
	if len(hash.Pairs) != 2 {
		t.Errorf("pairs = %d, want 2", len(hash.Pairs))
	}
}

func TestMissingAnnotationsAreParseErrors(t *testing.T) {
	tests := []struct{ name, src, wantMessage string }{
		{"var without type", `var x = 1;`, "Expect ':' and a type after variable name."},
		{"const without type", `const x = 1;`, "Expect ':' and a type after variable name."},
		{"param without type", `fun f(a) { }`, "Expect ':' and a type after parameter name."},
		{"second param without type", `fun f(a: number, b) { }`, "Expect ':' and a type after parameter name."},
		{"lambda param without type", `var f: fun(number) = fun(a) { };`, "Expect ':' and a type after parameter name."},
		{"method param without type", `class C { m(a) { } }`, "Expect ':' and a type after parameter name."},
		{"field without type", `class C { a; }`, "Expect '(' after named."},
		{"empty annotation", `var x: = 1;`, "Expect a type."},
		{"array type missing elem", `var x: [] = 1;`, "Expect a type."},
		{"map missing bracket", `var x: map string = 1;`, "Expect '[' after 'map'."},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, errs := parseSourceExpectingErrors(t, tc.src)
			if len(errs) == 0 {
				t.Fatalf("expected %q to be rejected", tc.src)
			}
			found := false
			for _, e := range errs {
				if strings.Contains(e.Message, tc.wantMessage) {
					found = true
				}
			}
			if !found {
				t.Errorf("got %v, want a diagnostic containing %q", errs, tc.wantMessage)
			}
		})
	}
}

// Diagnostics must carry a real line number so they can be printed with a
// location.
func TestAnnotationErrorsCarryAPosition(t *testing.T) {
	_, errs := parseSourceExpectingErrors(t, "var a: number = 1;\nvar b = 2;\n")
	if len(errs) == 0 {
		t.Fatal("expected an error for the unannotated declaration")
	}
	if errs[0].Token.Line != 2 {
		t.Errorf("error reported on line %d, want line 2", errs[0].Token.Line)
	}
}
