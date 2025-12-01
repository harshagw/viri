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
}
