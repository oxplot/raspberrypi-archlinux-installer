package main

import (
	"flag"
	"log"
)

const (
	title = "Raspberry Pi Arch Linux Installer"
)

var (
	cliMode   = flag.Bool("cli", false, "run in interactive command line mode")
	batchMode = flag.Bool("batch", false, "run in batch mode - good for scripting")
)

func main() {
	log.SetFlags(0)
	flag.Parse()
	if *cliMode && *batchMode {
		log.Fatal("only one of --cli and --batch can be specified, not both")
	}
	var err error
	if *cliMode {
		err = runInCLIMode()
	} else if *batchMode {
		err = runInBatchMode()
	} else {
		err = runInGUIMode()
	}
	if err != nil {
		log.Fatalf("error: %s", err)
	}
}
