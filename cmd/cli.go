package cmd

import "os"

type Runner func(args []string) int

type CLI struct {
	cmdMapping map[string]Runner
}

func NewCLI() *CLI {
	return &CLI{
		cmdMapping: map[string]Runner{},
	}
}

func (c *CLI) AddCommand(cmd string, runner Runner) {
	c.cmdMapping[cmd] = runner
}

func (c *CLI) Run(args []string) {
	cmd := args[1]

	status := c.cmdMapping[cmd](args[1:])
	os.Exit(status)

}
