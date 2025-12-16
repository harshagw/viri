import Link from "next/link";
import { Button } from "@/components/ui/button";

import { CodeSnippet } from "@/components/code-snippet";
import { FeatureList } from "@/components/feature-list";
import { SNIPPETS } from "@/lib/snippets";

const featureList = [
  { title: "Classes & inheritance", description: "define methods and reuse behavior via single inheritance" },
  { title: "Functions", description: "named functions with lexical scoping" },
  { title: "Closures", description: "functions capture surrounding lexical scope" },
  { title: "Module system", description: "file-based modules with explicit exports and alias-based imports" },
  { title: "Const bindings", description: "immutable names when you want them" },
  { title: "Arrays & hashes", description: "built-in indexed and key-value data structures" },
];

export default function Page() {
  return (
    <>
      <main className="flex-1">
        <section className="min-h-[60vh] flex items-center">
          <div className="container mx-auto px-6 max-w-4xl">
            <div className="max-w-xl">
              <h1 className="text-7xl font-bold font-mono text-primary mb-6">viri</h1>
              <p className="text-xl leading-relaxed text-foreground mb-8">A tiny language built for learning how languages work</p>
              <div className="flex gap-3 flex-wrap">
                <Link href="/playground">
                  <Button size="lg" variant="default">
                    try playground
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
              <a href="https://harshagw.github.io" target="_blank" rel="noopener noreferrer" className="underline hover:text-foreground transition-colors">
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
            </ul>
          </div>
        </div>
      </footer>
    </>
  );
}
