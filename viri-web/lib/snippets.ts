export const SNIPPETS = [
  {
    id: "functions",
    title: "functions",
    code: `fun fibonacci(n: number): number {
  if (n <= 1) return n;
  return fibonacci(n - 1)
       + fibonacci(n - 2);
}

print fibonacci(10);

var multiply: fun(number, number): number =
  fun(a: number, b: number): number {
    return a * b;
  };
print multiply(3, 4);

// IIFE
print fun(x: number): number { return x * x; }(5);`,
  },
  {
    id: "classes",
    title: "classes",
    code: `class Animal {
  name: string;

  init(name: string) {
    this.name = name;
  }
}

class Dog < Animal {
  init(name: string) {
    super.init(name);
  }

  speak() {
    print this.name + " barks";
  }
}

var rex: Dog = Dog("Rex");
rex.speak();

// every field is declared, and
// init must assign all of them`,
  },
  {
    id: "types",
    title: "types",
    code: `// Viri checks the whole program
// before any of it runs.

var count: number = 10;
var name: string = "Viri";

// Uncomment a line to see the
// checker reject it:

// var bad: number = "ten";
// print count + name;
// if (count) { print "hi"; }
// count.missing;

print name + " has " + "types";`,
  },
  {
    id: "modules",
    title: "modules",
    code: `import "std:math" as m;

print m.PI;
print m.pow(2, 3);
print m.sqrt(144);

// standard library calls are
// type-checked like any other:
// m.sqrt("144") is an error`,
  },
  {
    id: "datatypes",
    title: "data types",
    code: `var list: []number = [1, 2, 3];
print list[0];

var dict: map[string]string = {
  "name": "Viri",
  "ver": "1"
};
print dict["name"];

// an annotation is what gives an
// empty literal its type
var empty: []string = [];
print len(empty);

for (var i: number = 0; i < 10000; i = i + 1){
  print i;
}
  `,
  },
] as const;

export type SnippetId = (typeof SNIPPETS)[number]["id"];

export function getSnippet(id: string) {
  return SNIPPETS.find((s) => s.id === id);
}

export const DEFAULT_CODE = `print "Hello, Viri!";
var x: number = 10;
var y: number = 20;
print x + y;`;
