package checker

import (
	"sort"
	"strings"

	"github.com/harshagw/viri/internal/ast"
	"github.com/harshagw/viri/internal/token"
	"github.com/harshagw/viri/internal/types"
)

// Definite assignment closes the last hole in "no implicit nil": a field whose
// type says string must never be observed before something put a string in it.
//
// The analysis needs no control-flow graph. It walks the statement tree
// carrying the set of things definitely assigned so far:
//
//	sequence    union       — a; b assigns whatever a or b assigns
//	if/else     intersect   — only what both branches assign is guaranteed
//	if          nothing     — the branch may not run
//	while/for   nothing     — the body may run zero times
//	return      everything  — the path ends, so it constrains nothing after it
//
// That is conservative and sound: it accepts
//
//	if (c) { this.x = 1; } else { this.x = 2; }
//
// and rejects
//
//	while (c) { this.x = 1; }
//
// which may assign nothing at all.

// assigned is what a path has definitely done by some point.
type assigned struct {
	fields    map[string]bool
	superInit bool
	// returned marks a path that has already left the function. Such a path
	// constrains nothing, so joining it with another yields the other.
	returned bool
}

func newAssigned() assigned {
	return assigned{fields: make(map[string]bool)}
}

func (a assigned) clone() assigned {
	out := assigned{fields: make(map[string]bool, len(a.fields)), superInit: a.superInit, returned: a.returned}
	for k := range a.fields {
		out.fields[k] = true
	}
	return out
}

// union adds everything b did to a. Used for statements in sequence.
func (a *assigned) union(b assigned) {
	for k := range b.fields {
		a.fields[k] = true
	}
	a.superInit = a.superInit || b.superInit
	a.returned = a.returned || b.returned
}

// intersect keeps only what both paths did. Used where branches rejoin.
func intersect(a, b assigned) assigned {
	// A path that returned never reaches the join, so it cannot weaken the
	// other. `if (c) { return; } this.x = 1;` still assigns x.
	if a.returned {
		return b
	}
	if b.returned {
		return a
	}
	out := newAssigned()
	for k := range a.fields {
		if b.fields[k] {
			out.fields[k] = true
		}
	}
	out.superInit = a.superInit && b.superInit
	return out
}

// checkDefiniteAssignment verifies that constructing an instance of the class
// leaves every one of its fields holding a value, and that nothing reads one
// before it has been given that value.
//
// The read half matters as much as the write half: an unassigned slot holds no
// value at all, so observing one inside init is the one way nil could still be
// seen in a language that has no nil.
func (c *Checker) checkDefiniteAssignment(decl *ast.ClassStmt, class *types.Class) {
	previous := c.initClass
	c.initClass = class
	defer func() { c.initClass = previous }()

	var init *ast.FunctionStmt
	for _, method := range decl.Methods {
		if method.Name.Lexeme == "init" {
			init = method
			break
		}
	}

	superHasFields := class.Super != nil && len(class.Super.AllFields()) > 0

	if init == nil {
		// A class with no init of its own inherits its superclass's, which
		// already assigns the superclass's fields — so only the class's own
		// fields are a problem, because nothing would assign those.
		if len(class.Fields) > 0 {
			c.errorAt(decl.Name, "Class '"+class.Name+"' declares "+fieldList(class.Fields)+
				" but has no 'init' to assign "+plural(len(class.Fields), "it", "them")+".")
		}
		return
	}

	state := newAssigned()
	if init.Body != nil {
		state = c.analyzeStatements(init.Body.Statements, state)
	}

	var missing []string
	for _, field := range class.Fields {
		if !state.fields[field.Name] {
			missing = append(missing, "'"+field.Name+"'")
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		c.errorAt(init.Name, "'init' must assign "+strings.Join(missing, ", ")+
			" on every path; "+plural(len(missing), "this field has", "these fields have")+
			" no value otherwise.")
	}

	if superHasFields && !state.superInit {
		c.errorAt(init.Name, "'init' must call super.init(...) on every path, "+
			"because superclass '"+class.Super.Name+"' has fields to assign.")
	}
}

func (c *Checker) analyzeStatements(statements []ast.Stmt, state assigned) assigned {
	for _, stmt := range statements {
		state = c.analyzeStatement(stmt, state)
	}
	return state
}

func (c *Checker) analyzeStatement(stmt ast.Stmt, state assigned) assigned {
	switch s := stmt.(type) {
	case nil:
		return state

	case *ast.ExprStmt:
		return c.analyzeExpr(s.Expr, state)

	case *ast.PrintStmt:
		return c.analyzeExpr(s.Expr, state)

	case *ast.VarDeclStmt:
		return c.analyzeExpr(s.Initializer, state)

	case *ast.BlockStmt:
		return c.analyzeStatements(s.Statements, state)

	case *ast.IfStmt:
		after := c.analyzeExpr(s.Condition, state)
		thenState := c.analyzeStatement(s.ThenBranch, after.clone())
		if s.ElseBranch == nil {
			// The branch may not run, so it guarantees nothing.
			return after
		}
		return intersect(thenState, c.analyzeStatement(s.ElseBranch, after.clone()))

	case *ast.WhileStmt:
		after := c.analyzeExpr(s.Condition, state)
		// The body may run zero times, so it guarantees no assignment — but
		// if it does run, its reads happen, so they still have to be checked.
		c.analyzeDiscarding(s.Body, after)
		return after

	case *ast.ForStmt:
		after := c.analyzeStatement(s.Initializer, state)
		if s.Condition != nil {
			after = c.analyzeExpr(s.Condition, after)
		}
		c.analyzeDiscarding(s.Body, after)
		if s.Increment != nil {
			c.analyzeExprDiscarding(s.Increment, after)
		}
		return after

	case *ast.ReturnStmt:
		out := c.analyzeExpr(s.Value, state)
		out.returned = true
		return out
	}
	return state
}

func (c *Checker) analyzeExpr(expr ast.Expr, state assigned) assigned {
	switch e := expr.(type) {
	case nil:
		return state

	case *ast.SetExpr:
		// `this.x = v` is the assignment that counts. Evaluate the value
		// first: `this.a = this.b` reads b before assigning a.
		state = c.analyzeExpr(e.Value, state)
		if _, isThis := e.Object.(*ast.ThisExpr); isThis {
			state.fields[e.Name.Lexeme] = true
		}
		return state

	case *ast.CallExpr:
		state = c.analyzeExpr(e.Callee, state)
		for _, arg := range e.Arguments {
			state = c.analyzeExpr(arg, state)
			c.requireFullyBuilt(arg, "Cannot pass 'this'", state)
		}
		// super.init(...) is what assigns the superclass's fields.
		if super, ok := e.Callee.(*ast.SuperExpr); ok && super.Method.Lexeme == "init" {
			state.superInit = true
		}
		return state

	case *ast.GetExpr:
		state = c.analyzeExpr(e.Object, state)
		if _, isThis := e.Object.(*ast.ThisExpr); isThis {
			c.requireReadable(e.Name, e.Name.Lexeme, state)
		}
		return state

	case *ast.SuperExpr:
		// Any superclass method may read the superclass's fields, so it can
		// only be reached once super.init has assigned them.
		if e.Method.Lexeme != "init" && c.superNeedsInit() && !state.superInit {
			c.errorAt(e.Method, "Cannot call super."+e.Method.Lexeme+
				"(...) before super.init(...); the superclass's fields have no value yet.")
		}
		return state

	case *ast.BinaryExpr:
		return c.analyzeExpr(e.Right, c.analyzeExpr(e.Left, state))
	case *ast.LogicalExpr:
		// The right operand is short-circuited, so it guarantees nothing.
		return c.analyzeExpr(e.Left, state)
	case *ast.GroupingExpr:
		return c.analyzeExpr(e.Expr, state)
	case *ast.UnaryExpr:
		return c.analyzeExpr(e.Expr, state)
	case *ast.AssignExpr:
		return c.analyzeExpr(e.Value, state)
	case *ast.IndexExpr:
		return c.analyzeExpr(e.Index, c.analyzeExpr(e.Object, state))
	case *ast.SetIndexExpr:
		return c.analyzeExpr(e.Value, c.analyzeExpr(e.Index, c.analyzeExpr(e.Object, state)))
	case *ast.ArrayLiteralExpr:
		for _, el := range e.Elements {
			state = c.analyzeExpr(el, state)
		}
		return state
	case *ast.HashLiteralExpr:
		for _, pair := range e.Pairs {
			state = c.analyzeExpr(pair.Value, c.analyzeExpr(pair.Key, state))
		}
		return state

	case *ast.FunctionExpr:
		// A lambda may never be called, so nothing it assigns is guaranteed.
		// But it may also be called immediately, and if it closes over `this`
		// its reads happen now — so check them against the state here.
		if e.Body != nil {
			c.analyzeDiscarding(e.Body, state)
		}
		return state
	}

	// Literals, variables, this and super assign nothing on their own.
	return state
}

// analyzeDiscarding checks a statement for reads that are too early, then
// throws away whatever it assigned. It is for code that may or may not run:
// a loop body, or a lambda body that may never be called.
func (c *Checker) analyzeDiscarding(stmt ast.Stmt, state assigned) {
	c.analyzeStatement(stmt, state.clone())
}

func (c *Checker) analyzeExprDiscarding(expr ast.Expr, state assigned) {
	c.analyzeExpr(expr, state.clone())
}

func fieldList(fields []types.Field) string {
	names := make([]string, len(fields))
	for i, f := range fields {
		names[i] = "'" + f.Name + "'"
	}
	sort.Strings(names)
	return "field" + plural(len(fields), "", "s") + " " + strings.Join(names, ", ")
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

// --- read checks -----------------------------------------------------------
//
// These only apply inside init. Once init returns, definite assignment
// guarantees every field holds a value, so no later read can be too early.

// superNeedsInit reports whether the superclass has fields that only
// super.init can assign.
func (c *Checker) superNeedsInit() bool {
	return c.initClass != nil && c.initClass.Super != nil && len(c.initClass.Super.AllFields()) > 0
}

// isOwnField reports whether name is declared on the class itself rather than
// inherited. Own fields are assigned by this init; inherited ones by super's.
func (c *Checker) isOwnField(name string) bool {
	if c.initClass == nil {
		return false
	}
	for _, f := range c.initClass.Fields {
		if f.Name == name {
			return true
		}
	}
	return false
}

// requireReadable reports an error if `this.name` is read before it holds a
// value. A method reference counts as a read of everything, because the method
// body may touch any field.
func (c *Checker) requireReadable(tok *token.Token, name string, state assigned) {
	if c.initClass == nil {
		return
	}

	if c.isOwnField(name) {
		if !state.fields[name] {
			c.errorAt(tok, "Cannot read 'this."+name+"' before it is assigned.")
		}
		return
	}

	if _, inherited := c.initClass.LookupField(name); inherited {
		if c.superNeedsInit() && !state.superInit {
			c.errorAt(tok, "Cannot read inherited field 'this."+name+
				"' before super.init(...) assigns it.")
		}
		return
	}

	// A method: its body may read any field, so every field must already hold
	// a value before it can be called.
	if _, isMethod := c.initClass.LookupMethod(name); isMethod {
		if missing := c.unassigned(state); missing != "" {
			c.errorAt(tok, "Cannot call 'this."+name+"' before "+missing+
				" assigned; the method may read "+plural2(missing)+".")
		}
	}
}

// requireFullyBuilt reports an error if `this` escapes before the instance is
// complete — passing it somewhere means the receiver can read any field.
func (c *Checker) requireFullyBuilt(expr ast.Expr, what string, state assigned) {
	if c.initClass == nil {
		return
	}
	this, ok := expr.(*ast.ThisExpr)
	if !ok {
		return
	}
	if missing := c.unassigned(state); missing != "" {
		c.errorAt(this.Keyword, what+" before "+missing+" assigned.")
	}
}

// unassigned describes what the instance is still missing, or "" when it is
// fully built.
func (c *Checker) unassigned(state assigned) string {
	var missing []string
	for _, f := range c.initClass.Fields {
		if !state.fields[f.Name] {
			missing = append(missing, "'"+f.Name+"'")
		}
	}
	if c.superNeedsInit() && !state.superInit {
		missing = append(missing, "the superclass's fields")
	}
	if len(missing) == 0 {
		return ""
	}
	sort.Strings(missing)
	return strings.Join(missing, ", ") + " " + plural(len(missing), "is", "are")
}

func plural2(missing string) string {
	if strings.Contains(missing, ",") {
		return "them"
	}
	return "it"
}
