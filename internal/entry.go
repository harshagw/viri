package internal

import (
	"bytes"
	"fmt"
)

type ViriRuntimeConfig struct {
	DebugMode bool
	DisableWarning bool
}

type Viri struct {
	hasErrors bool
	config *ViriRuntimeConfig
}

func NewViriRuntime(config *ViriRuntimeConfig) *Viri {
	if config == nil {
		config = &ViriRuntimeConfig{
			DebugMode: false,
			DisableWarning: false,
		}
	}
	return &Viri{
		hasErrors: false,
		config: config,
	}
}

func (v *Viri) HasErrors() bool {
	return v.hasErrors
}

func (v *Viri) Error(token Token, message string) {
	fmt.Printf("Error at line %d: %s\n", token.Line, message)
	v.hasErrors = true
}

func (v *Viri) Warn(token Token, message string) {
	if v.config.DisableWarning {
		return
	}
	fmt.Printf("Warning at line %d: %s\n", token.Line, message)
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

	if v.config.DebugMode {
		astPrinter := NewAstPrinter()
		tree := astPrinter.PrintStatements(statements)
		fmt.Println(tree)
	}

	interpreter := NewInterpreter(v)

	resolver := NewResolver(v, interpreter)
	resolver.Resolve(statements)
	if v.hasErrors {
		return
	}

	err = interpreter.Interpret(statements)
	if err != nil {
		fmt.Println("Error interpreting expression:", err)
		v.hasErrors = true
		return
	}
}
