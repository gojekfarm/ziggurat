package main

import (
	"fmt"
	"github.com/gojekfarm/ziggurat-go/cmd"
	"os"
)

func main() {
	cli := cmd.NewCLI("zig")
	cli.AddUsage(`[USAGE]
zig command_name <args>`)
	cli.AddCommand("new", func(args []string) int {
		if len(args) < 2 {
			usage := `[USAGE]
zig new <app_name>`
			fmt.Println(usage)
			return 1
		}
		appName := args[1]
		d := cmd.Data{
			AppName:       appName,
			TopicEntity:   "my-topic-entity",
			ConsumerGroup: appName + "_" + "consumer",
		}
		tplConfig := cmd.GetTemplateConfig()
		zts := cmd.NewZigTemplateSet(appName, tplConfig)

		if err := zts.Parse(); err != nil {
			fmt.Println("command failed with error:", err)
			return 1
		}
		if err := zts.CreateOutFiles(); err != nil {
			fmt.Println("command failed with error:", err)
			return 1
		}
		if err := zts.Render(d); err != nil {
			fmt.Println("command failed with error:", err)
			return 1
		}
		fmt.Println("created app ", appName)
		return 0
	})
	cli.Run(os.Args)
}
