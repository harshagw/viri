import Link from "next/link";
import { Button } from "@/components/ui/button";

import { CodeSnippet } from "@/components/code-snippet";
import { FeatureList } from "@/components/feature-list";

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
              <p className="text-xl leading-relaxed text-foreground mb-8">A tiny language for learning big ideas.</p>
              <div className="flex gap-3">
                <Link href="/grammar">
                  <Button size="lg">explore grammar →</Button>
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
              <CodeSnippet
                title="functions"
                code={`fun fibonacci(n) {
  if (n <= 1) return n;
  return fibonacci(n - 1) 
       + fibonacci(n - 2);
}

print fibonacci(10);`}
              />
              <CodeSnippet
                title="classes"
                code={`class Animal {
  init(name) {
    this.name = name;
  }
}

class Dog < Animal {
  init(name){
    super.init(name);
  }

  speak() {
    print this.name + " barks";
  } 
}

var rex = Dog("Rex");
rex.speak();`}
              />
              <CodeSnippet
                title="modules"
                code={`import "math" as m;

print m.pi;
print m.pow(2, 3);

// easy to use modules
// for code organization`}
              />
              <CodeSnippet
                title="data types"
                code={`var list = [1, 2, 3];
print list[0];

var dict = {
  "name": "Viri",
  "ver": 1
};
print dict["name"];`}
              />
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
          <p className="text-sm text-muted-foreground">viri — a learning language</p>
          
          <div className="text-sm text-muted-foreground">
             <p className="font-medium mb-2">Shoutout to</p>
             <ul className="space-y-1">
               <li>
                 <a href="https://craftinginterpreters.com/" target="_blank" rel="noopener noreferrer" className="underline hover:text-foreground transition-colors">Crafting Interpreters</a> 
                 <span className="opacity-75"> by Robert Nystrom</span>
               </li>
               <li>
                 <a href="https://interpreterbook.com/" target="_blank" rel="noopener noreferrer" className="underline hover:text-foreground transition-colors">Writing An Interpreter In Go</a> 
                 <span className="opacity-75"> by Thorsten Ball</span>
               </li>
             </ul>
          </div>
        </div>
      </footer>
    </>
  );
}
