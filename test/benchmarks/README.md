# Benchmarks

Ten probes over the compiler + VM. The interpreter is not measured: it is
frozen and no longer runs the same language, so there is nothing to compare it
against.

## Running

```bash
make bench                 # quick look, single run
make bench-save            # record test/benchmarks/baseline.txt (count=6)
make bench-cmp             # measure the working tree, diff against baseline
```

`BENCH_FILTER` narrows the set and `BENCH_COUNT` changes repetitions:

```bash
make bench-cmp BENCH_FILTER='Arithmetic|MethodCall' BENCH_COUNT=10
```

`bench-cmp` runs the result through `benchstat`, which reports each delta with
a p-value. Anything without one is noise. Six repetitions is the floor for a
number worth acting on; use ten before claiming a change under 5%.

## Reading the numbers

**`allocs/op` is the trustworthy signal.** It is exactly reproducible and
independent of machine, thermal state and Go version, so a change there is
always real. Every profile taken of this VM has come back the same way: the
cost is allocation churn, not retention — live heap stays under a few MB even
in the 500k-entry probe.

**`sec/op` is indicative.** It moves with the host. Only compare timings
produced on the same machine in the same session, which is what `bench-cmp`
does and what a stored baseline from another machine does not.

Scanning, parsing, checking and codegen run once in `compileSource`, outside
the timer. These measure the VM, not the front end.

## The probes

Sizes match the original interpreter-vs-VM probe set so the two are broadly
comparable. Two of them had to change shape, because the language did:

- Viri has no `%`, so `ArrayReadWrite2M` and `HashSetGet200k` wrap their index
  with an explicit counter, adding a compare and a branch per iteration.
- Viri has no number-to-string conversion — the only natives are `clock` and
  `len` — so `HashBuild500k` builds keys by walking a generated array of
  literals, 708×708 = 501,264 entries, with one concatenation per entry inside
  the measurement.

## The baseline refreshes itself

`baseline.txt` is committed, and the pre-commit hook (`.husky/pre-commit`)
keeps it current. On any commit touching `cmd/`, `internal/` or `test/` it runs
`make test`, then `make e2e`, then `make bench-save`, and stages the result
into that same commit.

The benchmarks run last on purpose: the baseline should only move for a commit
that is actually going to happen. If the tests fail the hook stops before
reaching them, and `bench-save` writes through a temporary file, so a failed or
interrupted run leaves the previous baseline byte-identical rather than
truncated.

That leaves one gap worth knowing about. The hook cannot see what happens after
it returns, so aborting at the commit-message editor leaves `baseline.txt`
rewritten and staged with no commit behind it. Nothing breaks — the next commit
picks it up — but `git commit -m` sidesteps the editor entirely.

Six repetitions costs roughly 90 seconds. For a tighter loop:

```bash
BENCH_COUNT=1 git commit -m "..."   # faster, no variance estimate
git commit --no-verify -m "..."     # skip the hook altogether
```

A baseline is only meaningful against the machine that produced it, so expect
it to move when you commit from different hardware.
