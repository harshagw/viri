package internal

import (
	"bytes"
	"fmt"
)

type Viri struct {
	hasErrors bool
}

func NewViriRuntime() *Viri {
	return &Viri{
		hasErrors: false,
	}
}

func (v *Viri) HasErrors() bool {
	return v.hasErrors
}

func (v *Viri) Error(token Token, message string) {
	fmt.Printf("Error at line %d: %s\n", token.Line, message)
	v.hasErrors = true
}

func (v *Viri) Run(bytes *bytes.Buffer) {
	fmt.Println("------- source code ---------")
	fmt.Println(bytes.String())
	fmt.Println("------- source code ---------")

	scanner := NewScanner(bytes);
	tokens, err := scanner.scan()
	if err != nil {
		fmt.Println("Error parsing tokens:", err)
		v.hasErrors = true
		return
	}

	fmt.Println("------- tokens ---------")
	for _, token := range tokens {
		fmt.Println(token.ToString())
	}
	fmt.Println("------- tokens ---------")

	parser := NewParser(tokens, v);
	expr, err := parser.parse();
	if err != nil {
		fmt.Println("Error parsing expression:", err)
		v.hasErrors = true
		return
	}
	
	astPrinter := NewAstPrinter()
	fmt.Println("------- AST tree ---------")
	fmt.Print(astPrinter.PrintTree(expr))
	fmt.Println("------- AST tree ---------")

	interpreter := NewInterpreter(v)
	result, err := interpreter.Interpret(expr)
	if err != nil {
		fmt.Println("Error interpreting expression:", err)
		v.hasErrors = true
		return
	}
	fmt.Println("------- result ---------")
	fmt.Println(result)
	fmt.Println("------- result ---------")
}
