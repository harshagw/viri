package main

import (
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

	viri.Run(fileName)

	if viri.HasErrors() {
		os.Exit(70) // syntax error
	}
}
