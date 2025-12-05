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
	var fileName string
	var debugMode bool
	showWarning := true

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--debug" {
			debugMode = true
		} else if arg == "--no-warning" {
			showWarning = false
		} else if strings.HasSuffix(arg, FILE_EXTENSION) {
			fileName = arg
		}
	}

	if fileName == "" {
		fmt.Println("Usage: viri [--debug] <file>")
		os.Exit(64) // usage error
	}

	config := &internal.ViriRuntimeConfig{
		DebugMode:      debugMode,
		DisableWarning: !showWarning,
	}
	viri := internal.NewViriRuntime(config)

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
}
