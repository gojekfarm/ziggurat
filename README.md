### Ziggurat Golang

Stream Processing made easy

### Install the zig CLI

```shell script
go get -v -u github.com/gojekfarm/ziggurat/cmd/zig
go install github.com/gojekfarm/ziggurat/cmd/zig                                                                                                                                                       
```

#### How to use

- create a new app using the `new` command

```shell
zig new <app_name>
```

- sample `main.go`

### Pass an optional config file path

If you wish to read config from a location other than the default location run the app
with `--ziggurat-config="your_path/config_file.yaml"` option

```go
package main

import (
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/mw"
	"github.com/gojekfarm/ziggurat/mw/statsd"
	"github.com/gojekfarm/ziggurat/server"
)

const RoutePlainTextLog = "plain-text-log"

func main() {
	app := ziggurat.NewApp()
	router := ziggurat.NewRouter()
	statsdClient := statsd.NewStatsD(statsd.WithPrefix("super-app"))
	httpServer := server.New()

	router.HandleFunc(RoutePlainTextLog, func(event ziggurat.MessageEvent, app ziggurat.App) ziggurat.ProcessStatus {
		return ziggurat.ProcessingSuccess
	})

	rmw := router.Compose(mw.MessageLogger, statsdClient.PublishKafkaLag, statsdClient.PublishHandlerMetrics)

	app.OnStart(func(a ziggurat.App) {
		statsdClient.Start(app)
		httpServer.Start(app)
	})

	<-app.Run(rmw, ziggurat.Routes{
		RoutePlainTextLog: {
			InstanceCount:    1,
			BootstrapServers: "localhost:9092",
			OriginTopics:     "plain-text-log",
			GroupID:          "plain_text_consumer",
		},
	})
}

```