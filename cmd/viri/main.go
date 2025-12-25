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
	var engine string = "interpreter" // default to interpreter
	showWarning := true

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--debug" {
			debugMode = true
		} else if arg == "--no-warning" {
			showWarning = false
		} else if val, found := strings.CutPrefix(arg, "--engine="); found {
			engine = val
		} else if strings.HasSuffix(arg, FILE_EXTENSION) {
			fileName = arg
		}
	}

	if fileName == "" {
		fmt.Println("Usage: viri [--debug] [--engine=interpreter|vm] <file>")
		os.Exit(64) // usage error
	}

	if engine != "interpreter" && engine != "vm" {
		fmt.Println("Invalid engine. Use --engine=interpreter or --engine=vm")
		os.Exit(64) // usage error
	}

	config := &internal.ViriRuntimeConfig{
		DebugMode:      debugMode,
		DisableWarning: !showWarning,
		Engine:         engine,
	}
	viri := internal.NewViriRuntime(config)

	viri.Run(fileName)

	if viri.HasErrors() {
		os.Exit(70) // syntax error
	}
}
