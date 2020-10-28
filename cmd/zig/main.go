package main

import (
	"fmt"
	"github.com/gojekfarm/ziggurat-go/cmd"
	"html/template"
	"os"
)

type Data struct {
	AppName       string
	TopicEntity   string
	ConsumerGroup string
}

var modFile = `module {{.AppName}}

go 1.14

require (
	github.com/gojekfarm/ziggurat-go v0.4.3
	github.com/julienschmidt/httprouter v1.2.0
)`

var yamlFile = `
service-name: "{{.AppName}}"
stream-router:
  {{.TopicEntity}}:
    bootstrap-servers: "localhost:9092"
    instance-count: 1
    origin-topics: "my-topic"
    group-id: "{{.ConsumerGroup}}"
log-level: "debug"
retry:
  enabled: true
  count: 5
rabbitmq:
  host: "amqp://user:bitnami@localhost:5672/"
  delay-queue-expiration: "1000"
statsd:
  host: "localhost:8125"
http-server:
  port: 8080
`

var mainFile = `
package main

import (
	"github.com/gojekfarm/ziggurat-go/zig"
)

func main() {
	app := zig.NewApp()
	router := zig.NewRouter()

	router.HandlerFunc("{{.TopicEntity}}", func(messageEvent zig.MessageEvent, app *zig.App) zig.ProcessStatus {
		return zig.ProcessingSuccess
	}, zig.MessageLogger)

	app.Run(router, zig.RunOptions{
		StartCallback: func(a *zig.App) {

		},
		StopCallback: func() {
			
		},
	})

}
`

func main() {
	cli := cmd.NewCLI()
	cli.AddCommand("new", func(args []string) int {
		if len(args) < 2 {
			usage := `[USAGE]
zig new <app_name>`
			fmt.Println(usage)
			return 1
		}
		appName := args[1]
		d := Data{
			AppName:       appName,
			TopicEntity:   "my-topic-entity",
			ConsumerGroup: appName + "_" + "consumer",
		}
		if err := os.MkdirAll(appName+"/"+"config", 07770); err != nil {
			return 1
		}
		mainT, _ := template.New("mainTpl").Parse(mainFile)
		mainF, _ := os.Create(appName + "/main.go")
		mainT.Execute(mainF, d)
		yamlT, _ := template.New("yamlTpl").Parse(yamlFile)
		yamlF, _ := os.Create(appName + "/config" + "/config.yaml")
		yamlT.Execute(yamlF, d)
		modT, _ := template.New("modTpl").Parse(modFile)
		modF, _ := os.Create(appName + "/go.mod")
		modT.Execute(modF, d)
		fmt.Printf("successfully created app %s\n", appName)
		return 0
	})
	cli.Run(os.Args)
}
