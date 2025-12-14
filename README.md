# Viri

A scripting language interpreter written in Go. Viri supports variables, expressions, control flow, and more.

## Installation

```bash
go build -o viri cmd/viri/main.go
```

## Usage

```bash
./viri <file.viri>
```

## Grammar

```
program       ::= { importDecl } { declaration } "EOF" ;

importDecl    ::= "import" STRING "as" IDENTIFIER ";" ;

declaration   ::= classDecl
                | funDecl
                | varDecl
                | constDecl
                | statement ;

varDecl       ::= [ "export" ] "var" IDENTIFIER [ "=" expression ] ";" ;
constDecl     ::= [ "export" ] "const" IDENTIFIER "=" expression ";" ;
funDecl       ::= [ "export" ] "fun" function ;
classDecl     ::= [ "export" ] "class" IDENTIFIER [ "<" IDENTIFIER ]
                  "{" { function } "}" ;
function      ::= IDENTIFIER "(" [ parameters ] ")" block ;
parameters    ::= IDENTIFIER { "," IDENTIFIER } ;

statement     ::= exprStmt
                | forStmt
                | ifStmt
                | printStmt
                | returnStmt
                | whileStmt
                | breakStmt
                | continueStmt
                | block ;
returnStmt    ::= "return" [ expression ] ";" ;
forStmt       ::= "for" "(" ( varDecl | constDecl | exprStmt | ";" )
                  [ expression ] ";"
                  [ expression ] ")" statement ;
whileStmt     ::= "while" "(" expression ")" statement ;
ifStmt        ::= "if" "(" expression ")" statement
                  [ "else" statement ] ;
breakStmt     ::= "break" ";" ;
continueStmt  ::= "continue" ";" ;
block         ::= "{" { declaration } "}" ;
exprStmt      ::= expression ";" ;
printStmt     ::= "print" expression ";" ;

expression    ::= assignment ;
assignment    ::= ( call "." IDENTIFIER
                 | call "[" expression "]"
                 | IDENTIFIER ) "=" assignment
                 | logic_or ;
logic_or      ::= logic_and { "or" logic_and } ;
logic_and     ::= equality { "and" equality } ;
equality      ::= comparison { ( "!=" | "==" ) comparison } ;
comparison    ::= term { ( ">" | ">=" | "<" | "<=" ) term } ;
term          ::= factor { ( "-" | "+" ) factor } ;
factor        ::= unary { ( "/" | "*" ) unary } ;
unary         ::= ( "!" | "-" ) unary | call ;
call          ::= primary { "(" [ arguments ] ")"
                         | "." IDENTIFIER
                         | "[" expression "]" } ;
primary       ::= "true" | "false" | "nil" | "this"
                | NUMBER | STRING | IDENTIFIER | "(" expression ")"
                | "super" "." IDENTIFIER
                | arrayLiteral | hashLiteral ;

arguments     ::= expression { "," expression } ;

arrayLiteral  ::= "[" [ expression { "," expression } ] "]" ;
hashLiteral   ::= "{" [ hashEntry { "," hashEntry } ] "}" ;
hashEntry     ::= expression ":" expression ;
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

This interpreter is inspired by [Crafting Interpreters](https://craftinginterpreters.com/) by Robert Nystrom, an excellent resource for learning how to build programming languages.
