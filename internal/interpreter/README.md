# internal/interpreter

The tree-walking interpreter: a self-contained implementation of Viri that
executes the AST directly, without producing bytecode. It is one of two engines
(`--engine=interpreter`); the other is the compiler + VM in `internal/compiler`
and `internal/vm`.

Everything the interpreter needs lives under this directory — it imports
nothing from the rest of the tree.

## The flow

```
source text
    │
    ├─ scanner ──────────►  []token.Token
    │
    ├─ parser ───────────►  *ast.Module          imports + statements
    │                           │
    │                           └─ recursively parses imported files
    │
    ├─ resolver ─────────►  map[ast.Expr]int     scope distance per variable
    │                                             (+ compile-time diagnostics)
    │
    └─ interp ───────────►  walks the AST, evaluating as it goes
                            produces objects.Object values
```

The entry point is `parser.LoadModuleFile` → `parser.NewResolver().Resolve(mod)`
→ `interp.NewInterpreter(nil).Interpret(mod.GetAllStatements())`. See
`runWithInterpreter` in `internal/entry.go`.

## The packages

### `token`

Token types and the reserved-word table. `LookupKeyword` maps an identifier
lexeme to its keyword token type, or returns `IDENTIFIER`. A `Token` carries its
lexeme, literal value, line, and originating file path — the file path is what
lets runtime errors name the right file across module boundaries.

### `scanner`

Characters → tokens. Single pass, no backtracking. Handles numbers, strings,
comments, and multi-character operators. Unlike the later stages it does not use
the diagnostic handler — it accumulates messages (unexpected character, invalid
escape sequence, unterminated string, exponent with no digits) and `Scan`
returns them joined as a single error alongside the tokens it did manage to
produce.

### `ast`

The node types. `Expr` and `Stmt` are the two interfaces; every node implements
`GetPrimaryToken()` so diagnostics can point at a source location.

- `expr.go` — `BinaryExpr`, `CallExpr`, `GetExpr`/`SetExpr` (property access),
  `IndexExpr`/`SetIndexExpr`, `ArrayLiteralExpr`, `HashLiteralExpr`,
  `FunctionExpr`, `ThisExpr`, `SuperExpr`, …
- `stmt.go` — `VarDeclStmt`, `FunctionStmt`, `ClassStmt`, `IfStmt`, `WhileStmt`,
  `ForStmt`, `ImportStmt`, …
- `module.go` — `Module`, which separates `Imports` from `Statements`;
  `GetAllStatements()` returns imports first, so dependencies load before use.
- `printer.go` — renders a tree for `--debug`.

### `parser`

Tokens → AST, plus module loading and resolution.

- `parser.go` — recursive-descent for statements and declarations.
- `pratt_expression.go` — Pratt (precedence-climbing) parsing for expressions,
  which keeps operator precedence in one table rather than in the grammar shape.
- `module.go` — `ResolveModulePath` normalises an import path against the
  importing file's directory; `LoadModuleFile` reads, scans, and parses a file
  into an `*ast.Module`.
- `resolver.go` — a separate pass between parsing and execution, described below.

### `resolver`

The resolver walks the AST once before execution and answers a single question
for each variable reference: **how many scopes up is it?** The answer goes into
`map[ast.Expr]int`, handed to the interpreter as `locals`.

This exists because environments are chained at runtime. Without it, looking up
a variable means walking the chain and comparing names at every level, and a
closure that outlives its defining scope can silently resolve to the wrong
binding. With a precomputed distance, `Environment.GetAt(distance, name)` jumps
straight to the right environment. An expression missing from the map is a
global, looked up by name in `globals`.

While walking, it also reports what it can see statically, before any code runs:

- self-reference in an initializer (`var a = a;`)
- assignment to a `const`, and a `const` with no initializer
- duplicate declaration in the same scope
- `return` outside a function; `return` with a value inside `init`
- `break`/`continue` outside a loop (the loop counter resets inside a function
  body, so a `break` in a function nested in a loop is still rejected)
- `this` outside a class; `super` outside a class or in a class with no superclass
- a class inheriting from itself
- `export` on a non-global declaration
- circular imports, unresolvable import paths, unknown `std:` modules
- unused-variable warnings (emitted in `endScope`)

It tracks `currentFunction`, `currentClass`, and `loopDepth` to make those
context-sensitive checks, and follows imports so a program's whole module graph
is resolved in one pass.

Note what is *not* here: an undeclared variable is not a resolve error. Names
that miss every scope are assumed global, and the failure surfaces at runtime as
`'x' variable not found` when `globals.Get` misses.

### `objects`

Runtime values. `Object` is the interface — `Type() Type` and `Inspect() string`.

| | |
| --- | --- |
| `value.go` | `Number` (float64), `String`, `Bool`, `Nil` (singleton) |
| `array.go`, `hash.go` | `Array`, `Hash` (string keys only) |
| `functions.go` | `Function` — params, body, and the closure environment |
| `classes.go` | `Class`, `ClassInstance` |
| `callable.go` | `Callable` (`Call`/`Arity`) and `BlockExecutor` |
| `native_functions.go` | `NativeFunction` plus the built-ins `clock` and `len` |
| `environment.go` | `Environment` — the scope chain |
| `namespace.go` | `Namespace` — an imported module's exported names |
| `module.go` | `Module`, `ModuleCache` |
| `errors.go` | `RuntimeError`, plus `ReturnError`/`BreakError`/`ContinueError` |
| `diagnostic.go` | `DiagnosticHandler`, `DiagnosticCollector` |

Two details worth knowing:

**Control flow travels as errors.** `return`, `break`, and `continue` are
implemented by returning a sentinel error that unwinds the Go call stack until
the construct that handles it. That's why `ReturnError` carries a `Value`.

**`Callable` is the call protocol.** `Function`, `Class`, and `NativeFunction`
all implement it, so `visitCallExpr` doesn't care which it is. Calling a `Class`
constructs an instance and runs `init`; calling a `Function` binds parameters in
a fresh environment and executes the body.

### `stdlib`

Built-in modules, imported as `std:name`. A `NativeModule` is a name plus a map
of exported `objects.Object`, registered in `registry.go` at init. `std:math`
is the only one today. `ToNamespace` wraps a module's exports so `math.sqrt`
resolves as ordinary property access.

### `interp`

The evaluator. `Interpreter` holds the current `environment`, the `globals`
environment, the `locals` distance map from the resolver, and module state.

`evalStmt` and `evalExpr` are type switches over the AST that dispatch to a
`visitX` method per node. Execution is direct: `visitBinaryExpr` evaluates both
sides and applies the operator, `visitIfStmt` evaluates the condition and
executes one branch, and so on. `callDepth` is checked against `MaxCallDepth`
(1024) to turn runaway recursion into a "Stack overflow" error instead of
crashing the process.

**Scopes.** `ExecuteBlock` swaps in a new `Environment` and restores the
previous one on the way out, so a block's bindings disappear when it ends.
A `Function` captures the environment it was defined in, which is what makes
closures work.

**Classes.** `visitClass` builds a `Class` from its methods. If it has a
superclass, methods are defined in an extra environment holding `super`, so
`super.m()` resolves lexically. `ClassInstance.Get` checks fields first, then
walks the class chain via `LookupMethod` and binds the method to the receiver by
defining `this` in a new environment. Fields are created on assignment — there
are no declared fields.

**Modules.** `visitImportStmt` resolves the path, and for a stdlib import
defines a namespace directly. For a file import it runs the module body in a
fresh environment (`ExecuteModule`), collects its exports, and caches the result
so each module executes exactly once. The namespace reads that environment
live, so later writes to an exported variable stay visible to importers.
## Wiring

`run.go` at the top of this directory is the entry point: `Run(filePath,
Options)` parses, resolves, executes, and prints its own diagnostics and
runtime errors, reporting only whether the run failed.

It formats diagnostics itself rather than sharing the runtime's formatting.
That is deliberate: `*Viri` cannot implement both `objects.DiagnosticHandler`
interfaces at once — this package's and the compiler's are structurally
identical but parameterised on different `token` packages — and the two engines
are free to drift anyway. The ~20 duplicated lines of colour formatting are the
price of the engine being self-contained, and they get deleted along with it.

`internal/entry.go` names the interpreter in exactly one place, the branch in
`Run`.
