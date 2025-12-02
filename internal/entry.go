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
	scanner := NewScanner(bytes);
	tokens, err := scanner.scan()
	if err != nil {
		fmt.Println("Error parsing tokens:", err)
		v.hasErrors = true
		return
	}

	parser := NewParser(tokens, v);
	statements := parser.parse();
	
	if v.hasErrors{
		return;
	}

	interpreter := NewInterpreter(v)
	err = interpreter.Interpret(statements)
	if err != nil {
		fmt.Println("Error interpreting expression:", err)
		v.hasErrors = true
		return
	}
}
