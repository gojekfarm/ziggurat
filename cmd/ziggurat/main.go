package main

import (
	"github.com/gojekfarm/ziggurat/cmd"
	"github.com/gojekfarm/ziggurat/cmd/handlers"
	"os"
)

func main() {
	cli := cmd.NewCLI("ziggurat")
	cli.AddUsage(`[USAGE]
ziggurat command_name <args>`)
	cli.AddCommand("new", handlers.NewHandler)
	cli.Run(os.Args)
}
