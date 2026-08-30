# internal

Viri: a statically typed language compiled to bytecode and executed by a
stack VM. This is the live implementation — the one the language evolves in.

`internal/interpreter` holds a second, frozen implementation of the older
untyped language, reached only via `--engine=interpreter`. It has its own README
and shares no code with this tree.

## The flow

```
source text
    │
    ├─ scanner ────────►  []token.Token
    │
    ├─ parser ─────────►  *ast.Module           imports + statements
    │                         │
    │                         └─ recursively parses imported files
    │
    ├─ checker ────────►  verdict + map[ast.Expr]types.Type
    │                         │                  (nothing is emitted here)
    │                         │
    ├─ compiler ───────►  *objects.CompiledProgram
    │                         │                  bytecode, constants, debug info
    │
    └─ vm ─────────────►  runs it
```

The entry point is `Viri.runWithVM` in [entry.go](entry.go):
`compiler.CompileProgram(path)` — which loads, checks and compiles everything —
then `vm.New(program).RunProgram()`.

**The checker emits nothing.** It is a gate plus a lookup table: either the
program is rejected before any code exists, or every expression has a recorded
type that codegen uses to pick narrower instructions. Nothing runs unless the
whole program type-checks.

## The packages

| Package | Role |
| --- | --- |
| `token` | token types and the reserved-word table |
| `scanner` | characters → tokens |
| `ast` | syntax tree, including `TypeExpr` for annotations |
| `types` | compile-time types (`Number`, `Array`, `Class`, …) |
| `checker` | type-checks a module, records expression types |
| `compiler` | AST → bytecode, plus scope and slot allocation |
| `code` | opcode definitions and instruction encoding |
| `objects` | runtime values |
| `vm` | executes bytecode |
| `stdlib` | built-in modules (`std:math`) and their signatures |

### `ast` vs `types` — syntax and meaning

Both describe types, and the split matters. `ast.TypeExpr` is **syntax**: the
word the programmer wrote, one node per occurrence, unresolved. `types.Type` is
**meaning**: every `number` annotation in a program resolves to the same
`types.Number` singleton, which is what makes comparison cheap.

`ast.NamedType{Name: "number"}` and `ast.NamedType{Name: "Shape"}` are identical
in shape. Deciding that one names a builtin and the other a class is the
checker's job, not the parser's — one resolution path for every name in type
position.

`TypeExpr` is also kept apart from `Expr`. `x[y]` is an index operation as an
expression and part of a map type as a type; `[]number` is not an expression at
all. Separate interfaces mean a declaration cannot hold `3 + 4` where a type
belongs.

## Scopes

Two separate scope systems, deliberately not shared.

**The checker** keeps a stack of `map[string]binding` ([scope.go](checker/scope.go)),
pushed on entering a block and popped on leaving. It answers "what type does
this name have, and is it const?"

Builtin type names live in that stack's outermost level. They are not keywords:
the parser treats `number` as an ordinary identifier, and this map is the only
place the distinction exists. One rule closes the resulting hole — declaring
anything named `number`, `string` or `bool` is an error, so a word cannot mean a
type in one position and a variable in another.

**The compiler** keeps a `SymbolTable` ([symbol_table.go](compiler/symbol_table.go))
answering a different question: "where does this name live at runtime?" Each
symbol gets a scope and an index:

| Scope | Storage | Opcodes |
| --- | --- | --- |
| `GlobalScope` | module globals array | `OpGetGlobal` / `OpSetGlobal` |
| `LocalScope` | function stack frame | `OpGetLocal` / `OpSetLocal` |
| `FreeScope` | captured, boxed in a `Cell` | `OpGetFree` / `OpSetFree` |
| `NativeScope` | the built-in function table | `OpGetNative` |
| `FunctionScope` | the running closure, for self-reference | `OpGetCurrentClosure` |

Keeping the two apart is deliberate: typing and slot allocation are different
problems, and entangling them would mean neither could change alone.

## Functions

A function compiles to a `CompiledFunction` — its own instruction stream, its
local count, and its parameter count — which is stored in the constants table.

At runtime it becomes a `Closure`: the compiled function plus the free variables
it captured. `emitClosure` walks the free symbols, emits `OpMakeCell` or
`OpGetFree` for each, and then `OpGetClosure <constIdx> <numFree>`.

A captured variable is boxed in a `Cell` — a one-field mutable container. That
box is why two closures capturing the same variable see each other's writes; a
plain copy would not.

Calling pushes a `Frame`. `OpReturnValue` returns the value on the stack;
`OpReturn` returns from a function that declares no return type, and the checker
guarantees nothing consumes that result.

Signatures are collected before any body is checked, so functions may refer to
each other in any order — mutual recursion works, as does calling a function
declared later in the file.

## Classes

A class declares its fields, and that declaration is load-bearing.

```viri
class Shape {
    name: string;
    init(name: string) { this.name = name; }
    area(): number { return 0; }
}
```

**Fields are slots, not names.** The checker builds an ordered layout —
superclass fields first, then the class's own — and records the index of every
field access. `CompiledInstance.Fields` is a `[]Object`, and `p.x` compiles to
`OpGetField <slot>`: an array index, no hash lookup, no "undefined property"
case. Because superclass fields come first, a subclass instance reads correctly
through a superclass type with no remapping.

Methods still resolve by name (`OpGetProperty`), because they live on the class
rather than in the instance, and `LookupMethod` walks the superclass chain.

**Dynamic field creation is gone.** Assigning to an undeclared field is a
compile error; there is no name-based store to fall back on.

**Definite assignment** is what makes the field types honest. Every field must
hold a value once `init` returns, and nothing may read one before it does. The
analysis ([definite.go](checker/definite.go)) needs no control-flow graph — it
carries an assigned-set over the statement tree, unioning in sequence and
intersecting where branches rejoin. A lone `if` or any loop body guarantees
nothing, because it may not run; but its *reads* are still checked, because if
it does run they happen.

That covers the ways a half-built instance could otherwise escape: reading a
field before assigning it, calling a method that might, passing `this` out, a
lambda closing over `this`, and reading an inherited field before `super.init`.

Inheritance is single and nominal: `Square` is assignable to `Shape` because it
declares `Shape` as its superclass, never because the shapes happen to match. An
overriding method must have exactly the signature it replaces; `init` is exempt,
since constructors are not virtual.

## Modules

Three steps, all in [module.go](compiler/module.go).

**Load.** `loadModule` recurses into imports and appends each module to
`moduleOrder` *after* its dependencies. The list is therefore already
topologically sorted — there is no separate sort — and cycles are caught on the
way down.

**Check.** `checkProgram` walks `moduleOrder`, so a module's imports are always
checked before it. Each module yields a `types.Module` its importers consume.

A module has two namespaces, because an exported class is both. In value
position `shapes.Circle` is a constructor — a function returning an instance —
while in a type annotation it is the class. `types.Module` keeps `Exports` and
`Classes` apart, which is what stops `var c: shapes.Circle = shapes.Circle;`
from checking.

**Compile.** Each module gets its own globals, numbered from zero. Exported
names are collected in declaration order into `Exports []int`, mapping export
index to global slot. `math.add` compiles to
`OpGetModuleExport <moduleIdx> <exportIdx>` — two array lookups at runtime, no
name resolution.

Stdlib is different: `std:math` is not compiled. The export object is copied
into the constants table by name and emitted as `OpGetStdlibExport <constIdx>`.
Its signatures live in `stdlib.ModuleType`, so `math.sqrt("hi")` is a compile
error like any other bad call.

`RunProgram` executes modules in `moduleOrder`, which is why post-order loading
matters: a module's dependencies have already run and filled their globals
before anything reads them. Globals are rebound on every call and return, keyed
by `frame.cl.Fn.ModuleIdx` — a function defined in module A but *called* from B
still reads A's globals.