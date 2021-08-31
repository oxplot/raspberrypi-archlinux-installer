package main

import (
	"flag"
	"log"
)

const (
	title = "Raspberry Pi Arch Linux Installer"
)

var (
	interactiveMode = flag.Bool("interactive", false, "run in interactive command line mode")
	batchMode       = flag.Bool("batch", false, "run in batch mode - good for scripting")
)

func main() {
	log.SetFlags(0)
	flag.Parse()
	if *interactiveMode && *batchMode {
		log.Fatal("only one of --interactive and --batch can be specified, not both")
	}
	var err error
	if *interactiveMode {
		err = runInInteractiveMode()
	} else if *batchMode {
		err = runInBatchMode()
	} else {
		err = runInGUIMode()
	}
	if err != nil {
		log.Fatalf("error: %s", err)
	}
}
