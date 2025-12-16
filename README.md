# Viri

A tiny language interpreter written in Go. Viri supports variables, expressions, control flow, and more.
 Check out the [Viri Website](https://harshagw.github.io/viri/) for grammar.
## Installation

```bash
go build -o viri cmd/viri/main.go
```

## Usage

```bash
./viri <file.viri>
```



## Example

```viri
var greeting = "Hello, World!";

print greeting;

var count = 0;

print "Counting from 0 to 5:";

while (count <= 5) {
    print count;
    count = count + 1;
}

print "Multiples of 2 up to 10:";

for (var i = 2; i <= 10; i = i + 2) {
    print i;
}

var sum = 0;
for (var j = 1; j <= 10; j = j + 1) {
    sum = sum + j;
}

print "Sum of 1 to 10:";
print sum;

var result = (10 + 5) * 2;
if (result > 20) {
    print "Result is greater than 20!";
} else {
    print "Result is 20 or less.";
}
```

## Reference

1. [Crafting Interpreters](https://craftinginterpreters.com/) by Robert Nystrom
2. [Writing an Interpreter in Go](https://interpreterbook.com/) by Thorsten Ball
