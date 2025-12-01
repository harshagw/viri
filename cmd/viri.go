package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/harshagw/viri/internal"
)

const FILE_EXTENSION = ".viri"

func main() {
	viri := internal.NewViriRuntime()

	if(len(os.Args) == 2) {
		fileName := os.Args[1]
		// check filename ends with .viri

		if !strings.HasSuffix(fileName, FILE_EXTENSION) {
			fmt.Println("File must end with " + FILE_EXTENSION)
			os.Exit(66) // cannot open input
		}

		code, err := os.ReadFile(fileName)
		if err != nil {
			fmt.Println("Error reading file:", err)
			os.Exit(66) // cannot open input
		}

		codeBytes := bytes.NewBuffer(code)
		viri.Run(codeBytes)

		if viri.HasErrors() {
			os.Exit(70) // syntax error
		}
	} else {
		fmt.Println("Usage: viri <file>")
		os.Exit(64) // usage error
	} 
}