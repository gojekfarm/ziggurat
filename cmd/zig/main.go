package main

import (
	"fmt"
	"github.com/gojekfarm/ziggurat-go/cmd"
	"os"
	"text/template"
)

type Data struct {
	AppName       string
	TopicEntity   string
	ConsumerGroup string
}

var modFile = `module {{.AppName}}

go 1.14

require (
	github.com/gojekfarm/ziggurat-go v0.4.4
	github.com/julienschmidt/httprouter v1.2.0
)`

var yamlFile = `service-name: "{{.AppName}}"
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

var kafkaComposeFile = `version: '3.3'

services:
  zookeeper:
    image: 'bitnami/zookeeper:latest'
    ports:
      - '2181:2181'
    container_name: '{{.AppName}}_zookeeper'
    environment:
      - ALLOW_ANONYMOUS_LOGIN=yes
  kafka:
    image: 'bitnami/kafka:2'
    ports:
      - '9092:9092'
    container_name: '{{.AppName}}_kafka'
    environment:
      - KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181
      - KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092
      - ALLOW_PLAINTEXT_LISTENER=yes
  rabbitmq:
    image: 'bitnami/rabbitmq:latest'
    ports:
      - '15672:15672'
      - '5672:5672'
`
var metricsComposeFile = `influxdb:
  image: influxdb:latest
  container_name: influxdb
  ports:
    - "8083:8083"
    - "8086:8086"
    - "8090:8090"

telegraf:
  image: telegraf:latest
  container_name: telegraf
  links:
    - influxdb
  ports:
    - "8125:8125/udp"
    - "9273:9273"
  volumes:
    - ./telegraf.conf:/etc/telegraf/telegraf.conf:ro

grafana:
  image: grafana/grafana:latest
  container_name: grafana
  ports:
    - "3000:3000"
  user: "0"
  links:
    - influxdb
    - prometheus
  volumes:
    - ./data/grafana/data:/grafana

prometheus:
  image: prom/prometheus
  container_name: prometheus
  ports:
    - "9090:9090"
  links:
    - influxdb
    - telegraf
  volumes:
    - ./prom_conf.yml:/etc/prometheus/prometheus.yml:ro
    - ./data/prometheus/data:/prometheus
`

var mainFile = `package main

import (
	"github.com/gojekfarm/ziggurat-go/zig"
)

func main() {
	app := zig.NewApp()
	router := zig.NewRouter()

	router.HandlerFunc("{{.TopicEntity}}", func(messageEvent zig.MessageEvent, app *zig.App) zig.ProcessStatus {
		return zig.ProcessingSuccess
	}, zig.MessageLogger)

	<-app.Run(router, zig.RunOptions{
		StartCallback: func(a *zig.App) {

		},
		StopCallback: func() {
			
		},
	})

}
`

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
		d := Data{
			AppName:       appName,
			TopicEntity:   "my-topic-entity",
			ConsumerGroup: appName + "_" + "consumer",
		}
		if err := os.MkdirAll(appName+"/"+"config", 07770); err != nil {
			return 1
		}
		if err := os.MkdirAll(appName+"/"+"sandbox", 07770); err != nil {
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
		kafkaComposeT, _ := template.New("kafkaComposeTpl").Parse(kafkaComposeFile)
		kafkaComposeF, _ := os.Create(appName + "/" + "sandbox" + "/kafka-compose.yaml")
		kafkaComposeT.Execute(kafkaComposeF, d)
		metricsComposeT, _ := template.New("metricsComposeTpl").Parse(metricsComposeFile)
		metricsComposeF, _ := os.Create(appName + "/" + "sandbox" + "/metrics-compose.yaml")
		metricsComposeT.Execute(metricsComposeF, d)

		fmt.Printf("successfully created app %s\n", appName)
		return 0
	})
	cli.Run(os.Args)
}
