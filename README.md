# Viri

Viri is a small, statically typed programming language, built to learn how
languages work.

Every variable, parameter, class field and return type is declared, and the
whole program is type-checked before any of it runs. There is no `any`, no
untyped escape hatch, and no `nil` — every type holds a real value.

```viri
class Shape {
    name: string;
    init(name: string) { this.name = name; }
    area(): number { return 0; }
}

class Square < Shape {
    side: number;
    init(side: number) {
        super.init("square");
        this.side = side;
    }
    area(): number { return this.side * this.side; }
}

var s: Shape = Square(4);
print s.area();
```

Try it in the browser at the [Viri playground](https://harshagw.github.io/viri/),
or read the [grammar reference](https://harshagw.github.io/viri/grammar).

## Two engines, one of them frozen

Viri has two implementations, and they are **no longer the same language**:

| Engine | Flag | Status |
| --- | --- | --- |
| Compiler + bytecode VM | `--engine=vm` (default) | the live language — statically typed |
| Tree-walking interpreter | `--engine=interpreter` | frozen; the older untyped language |

The interpreter came first. When Viri became statically typed, it was
vendor-frozen into `internal/interpreter/` rather than dragged along: it keeps
its own copy of the scanner, parser and AST, imports nothing from the live tree,
and is never edited. It still runs the untyped programs it always did, and it
will be deleted once it has nothing left to teach.

So a typed program will not run under `--engine=interpreter`, and an untyped one
will not compile under the VM. That divergence is deliberate.

Each tree has its own README:

- [internal/README.md](internal/README.md) — the live compiler and VM
- [internal/interpreter/README.md](internal/interpreter/README.md) — the frozen interpreter

## Installation

```bash
go build -o viri cmd/viri/main.go
```

## Usage

```bash
./viri <file.viri>
```

Useful flags: `--debug` prints the compiled bytecode, `--stats` prints timing,
and `--no-warning` silences warnings.

## Example

```viri
var greeting: string = "Hello, World!";
print greeting;

fun sumTo(n: number): number {
    var total: number = 0;
    for (var i: number = 1; i <= n; i = i + 1) {
        total = total + i;
    }
    return total;
}

print sumTo(10);

var counts: map[string]number = {"a": 1, "b": 2};
print counts["a"];

var names: []string = ["viri"];
print names[0];

var double: fun(number): number = fun(n: number): number { return n * 2; };
print double(21);
```

An ill-typed program is rejected before it runs, with a line number:

```
$ ./viri broken.viri
Error in broken.viri at line 3: Cannot assign string to 'count' of type number.
Error in broken.viri at line 5: An if condition must be bool, got number.
```

## Development

```bash
make test    # unit tests
make e2e     # end-to-end suites, both engines
make bench   # VM benchmarks
```

The website and its WebAssembly playground live in
[viri-web/](viri-web/README.md), which documents how to run it locally and how
to pull compiler changes into it.

## Reference

1. [Crafting Interpreters](https://craftinginterpreters.com/) by Robert Nystrom
2. [Writing an Interpreter in Go](https://interpreterbook.com/) by Thorsten Ball
3. [Writing a Compiler in Go](https://compilerbook.com/) by Thorsten Ball
