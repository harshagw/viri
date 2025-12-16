export const SNIPPETS = [
  {
    id: "functions",
    title: "functions",
    code: `fun fibonacci(n) {
  if (n <= 1) return n;
  return fibonacci(n - 1) 
       + fibonacci(n - 2);
}

print fibonacci(10);`,
  },
  {
    id: "classes",
    title: "classes",
    code: `class Animal {
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
rex.speak();`,
  },
  {
    id: "modules",
    title: "modules",
    code: `import "math" as m;

print m.pi;
print m.pow(2, 3);

// easy to use modules
// for code organization`,
  },
  {
    id: "datatypes",
    title: "data types",
    code: `var list = [1, 2, 3];
print list[0];

var dict = {
  "name": "Viri",
  "ver": 1
};
print dict["name"];

for (var i = 0; i < 10000; i = i + 1){
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
var x = 10;
var y = 20;
print x + y;`;
