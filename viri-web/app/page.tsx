import Link from "next/link";
import { Button } from "@/components/ui/button";

import { CodeSnippet } from "@/components/code-snippet";
import { FeatureList } from "@/components/feature-list";
import { Playground } from "@/components/playground";
import { SNIPPETS, DEFAULT_CODE } from "@/lib/snippets";

const featureList = [
  { title: "Static types, checked ahead of time", description: "every variable, parameter, field and return type is declared; nothing runs until the whole program checks" },
  { title: "No nil", description: "every type holds a real value, so there is no null to guard against" },
  { title: "Classes & inheritance", description: "declared fields, single inheritance, and constructors that must leave every field assigned" },
  { title: "Functions & closures", description: "first-class functions with typed signatures that capture their surrounding scope" },
  { title: "Module system", description: "file-based modules with explicit exports and alias-based imports" },
  { title: "Arrays & maps", description: "typed collections — []number, map[string]bool — with no untyped escape hatch" },
];

export default function Page() {
  return (
    <>
      <main className="flex-1">
        <section className="pt-20 pb-12">
          <div className="container mx-auto px-6 max-w-4xl">
            <div className="max-w-2xl">
              <h1 className="text-7xl font-bold font-mono text-primary mb-6">viri</h1>
              <p className="text-xl leading-relaxed text-foreground mb-4">A small, statically typed programming language</p>
              <p className="text-base leading-relaxed text-muted-foreground mb-8">
                Built from scratch to learn how languages work — a hand-written scanner, parser, type checker, compiler and bytecode VM, all in Go.
              </p>
              <div className="flex gap-3 flex-wrap">
                <Link href="/playground">
                  <Button size="lg" variant="default">
                    open playground
                  </Button>
                </Link>
                <Link href="/grammar">
                  <Button variant="secondary" size="lg">
                    grammar
                  </Button>
                </Link>
                <Link href="https://github.com/harshagw/viri" target="_blank">
                  <Button variant="secondary" size="lg">
                    github
                  </Button>
                </Link>
              </div>
            </div>
          </div>
        </section>

        {/* The playground, right up front — the fastest way to see what the
            language is actually like. */}
        <section className="pb-20">
          <div className="container mx-auto px-6 max-w-6xl">
            <p className="text-sm text-muted-foreground mb-4 uppercase tracking-wider">Try it — runs entirely in your browser</p>
            <div className="border border-border rounded-lg overflow-hidden shadow-sm">
              <Playground initialCode={DEFAULT_CODE} className="h-[520px]" allowFullscreen />
            </div>
            <p className="text-xs text-muted-foreground mt-3">
              Viri compiles to bytecode and runs on its own virtual machine, built here for WebAssembly — the same compiler the{" "}
              <code className="font-mono">viri</code> binary uses. Nothing is sent to a server.
            </p>
          </div>
        </section>

        <section className="py-20 border-t border-border">
          <div className="container mx-auto px-6 max-w-4xl">
            <p className="text-sm text-muted-foreground mb-6 uppercase tracking-wider">A taste</p>
            <div className="grid md:grid-cols-2 gap-6">
              {SNIPPETS.map((snippet) => (
                <CodeSnippet key={snippet.id} id={snippet.id} title={snippet.title} code={snippet.code} />
              ))}
            </div>
          </div>
        </section>

        <section className="py-20 border-t border-border">
          <div className="container mx-auto px-6 max-w-4xl">
            <p className="text-sm text-muted-foreground mb-8 uppercase tracking-wider">What you get</p>
            <FeatureList features={featureList} />
          </div>
        </section>
      </main>

      {/* Footer - minimal */}
      <footer className="border-t border-border py-8">
        <div className="container mx-auto px-6 max-w-4xl flex flex-col md:flex-row justify-between items-start gap-6">
          <div className="flex flex-col gap-2">
            <p className="text-sm text-muted-foreground">viri — a learning language</p>
            <p className="text-sm text-muted-foreground">
              Made by{" "}
              <a href="https://harshagw.dev" target="_blank" rel="noopener noreferrer" className="underline hover:text-foreground transition-colors">
                Harsh Agarwal
              </a>
            </p>
          </div>

          <div className="text-sm text-muted-foreground">
            <p className="font-medium mb-2">Shoutout to</p>
            <ul className="space-y-1">
              <li>
                <a
                  href="https://craftinginterpreters.com/"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="underline hover:text-foreground transition-colors"
                >
                  Crafting Interpreters
                </a>
                <span className="opacity-75"> by Robert Nystrom</span>
              </li>
              <li>
                <a href="https://interpreterbook.com/" target="_blank" rel="noopener noreferrer" className="underline hover:text-foreground transition-colors">
                  Writing An Interpreter In Go
                </a>
                <span className="opacity-75"> by Thorsten Ball</span>
              </li>
              <li>
                <a href="https://compilerbook.com/" target="_blank" rel="noopener noreferrer" className="underline hover:text-foreground transition-colors">
                  Writing An Compiler In Go
                </a>
                <span className="opacity-75"> by Thorsten Ball</span>
              </li>
            </ul>
          </div>
        </div>
      </footer>
    </>
  );
}
