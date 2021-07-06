package handlers

import (
	"fmt"

	"github.com/gojekfarm/ziggurat/cmd"
)

var NewHandler = func(args []string) int {
	if len(args) < 2 {
		usage := `[USAGE]
zig new <app_name>`
		fmt.Println(usage)
		return 1
	}
	appName := args[1]
	d := cmd.Data{
		AppName:          appName,
		TopicEntity:      "plain-text-log",
		ConsumerGroup:    appName + "_" + "consumer",
		OriginTopics:     "plain-text-log",
		BootstrapServers: "localhost:9092",
		ZigguratVersion:  "v1.2.3",
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
}
