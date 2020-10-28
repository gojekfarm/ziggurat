package cmd

import (
	"fmt"
	"os"
)

type Runner func(args []string) int

type CLI struct {
	cmdMapping map[string]Runner
	binName    string
	usage      string
}

func NewCLI(binName string) *CLI {
	return &CLI{
		cmdMapping: map[string]Runner{},
		binName:    binName,
	}
}

func (c *CLI) AddUsage(usage string) {
	c.usage = usage
}

func (c *CLI) AddCommand(cmd string, runner Runner) {
	c.cmdMapping[cmd] = runner
}

func (c *CLI) Run(args []string) {
	if len(args) < 2 {
		fmt.Println(c.usage)
		os.Exit(127)
	}

	cmd := args[1]

	status := c.cmdMapping[cmd](args[1:])
	os.Exit(status)

}
