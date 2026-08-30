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
	var statsMode bool
	var engine string = "vm" // default to the compiler + VM pipeline
	showWarning := true

	usage := func() {
		fmt.Println("Usage: viri [--debug] [--stats] [--no-warning] [--engine=vm|interpreter] <file.viri>")
		os.Exit(64) // EX_USAGE
	}

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--debug":
			debugMode = true
		case arg == "--no-warning":
			showWarning = false
		case arg == "--stats":
			statsMode = true
		case strings.HasPrefix(arg, "--engine="):
			engine = strings.TrimPrefix(arg, "--engine=")
		case strings.HasPrefix(arg, "-"):
			fmt.Printf("Unknown flag: %s\n", arg)
			usage()
		case fileName == "":
			fileName = arg
		default:
			fmt.Printf("Unexpected argument: %s\n", arg)
			usage()
		}
	}

	if fileName == "" || !strings.HasSuffix(fileName, FILE_EXTENSION) {
		usage()
	}

	if engine != "interpreter" && engine != "vm" {
		fmt.Println("Invalid engine. Use --engine=interpreter or --engine=vm")
		usage()
	}

	config := &internal.ViriRuntimeConfig{
		DebugMode:      debugMode,
		StatsMode:      statsMode,
		DisableWarning: !showWarning,
		Engine:         engine,
	}
	viri := internal.NewViriRuntime(config)

	viri.Run(fileName)

	if viri.HasErrors() {
		os.Exit(65) // EX_DATAERR: the program failed to compile or run
	}
}
