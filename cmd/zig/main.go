package main

import (
	"github.com/gojekfarm/ziggurat-go/cmd"
	"github.com/gojekfarm/ziggurat-go/cmd/handlers"
	"os"
)

func main() {
	cli := cmd.NewCLI("zig")
	cli.AddUsage(`[USAGE]
zig command_name <args>`)
	cli.AddCommand("new", handlers.NewHandler)
	cli.Run(os.Args)
}
