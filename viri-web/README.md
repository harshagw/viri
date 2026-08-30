# viri-web

The Viri website: the landing page, the grammar reference, and the playground.

The playground is not a simulation. It is the real compiler and VM
(`internal/compiler`, `internal/vm`) built for WebAssembly, so a program behaves
in the browser exactly as it does under the `viri` binary. Nothing is sent to a
server.

## Requirements

- Node 20+
- Go (matching `go.mod` at the repo root) — needed to build the WASM bundle

## Running locally

From the repository root:

```bash
make build-playground
```

That compiles `cmd/web-playground` to `viri-web/public/viri.wasm` and copies
Go's `wasm_exec.js` next to it. Both are gitignored build outputs, so this step
is required on a fresh clone — without it the playground loads but never
becomes ready.

Then, in `viri-web`:

```bash
npm install
npm run dev
```

The site is at http://localhost:3000.

## After changing the compiler

The browser runs a **prebuilt** `viri.wasm`. Editing Go source has no effect on
the running site until that bundle is rebuilt — the dev server watches
TypeScript, not Go.

So after any change under `internal/` or `cmd/web-playground`, from the repo
root:

```bash
make build-playground
```

Then hard-reload the browser (⌘⇧R / Ctrl-Shift-R). A normal reload often serves
the cached `.wasm`, which looks exactly like your change not working.

Worth running before you rebuild, since it catches most problems faster than
clicking around:

```bash
go test ./cmd/web-playground/
```

Those tests drive the same pipeline the browser does, and include
`TestSiteSnippetsRun`, which reads the snippets out of `lib/snippets.ts` and
runs every one. If you change a snippet and it no longer compiles, that test
fails rather than the front page.

### When a language change lands

Few places on the site describe the language and can drift out of date:

| File | What it holds |
| --- | --- |
| `lib/snippets.ts` | the homepage examples and the playground's starting code |
| `app/grammar/page.tsx` | the grammar reference |

`viri-syntax-plugin/` at the repo root is the VS Code grammar and needs the same
treatment; it is not built or checked by this project.

## Deploying

Deployment is a **manual trigger**, on purpose — the site should update when you
decide it should, not on every push to `main`.

Go to **Actions → Deploy Viri Web to Pages → Run workflow** on GitHub. The
workflow (`.github/workflows/deploy-viri-web.yml`) builds the WASM bundle from
the current commit, builds the Next.js static export, and publishes to GitHub
Pages.

The WASM bundle is built in CI from source, so a deploy always ships the
compiler at that commit — a stale local `viri.wasm` cannot be published by
accident.

## Layout

```
app/
  page.tsx            landing page, with the playground embedded
  playground/         the full-window playground
  grammar/            grammar reference
components/
  playground.tsx      the playground itself, shared by both pages
hooks/
  use-viri-playground.ts   loads viri.wasm in a worker, with a 10s timeout
lib/
  snippets.ts         example programs
  prism-viri.ts       syntax highlighting grammar
public/
  viri.wasm           build output (gitignored)
  wasm_exec.js        build output (gitignored)
```

The playground runs in a Web Worker with a 10-second timeout, so a runaway
program in the editor cannot lock up the page.
