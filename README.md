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
	"context"
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/kafka"
	"github.com/gojekfarm/ziggurat/logger"
	"github.com/gojekfarm/ziggurat/mw"
	"github.com/gojekfarm/ziggurat/router"
)

func main() {
	jsonLogger := logger.NewJSONLogger("info")

	kafkaStreams := &kafka.Streams{
		RouteGroup: kafka.RouteGroup{
			"plain-text-log": {
				BootstrapServers: "localhost:9092",
				OriginTopics:     "plain-text-log",
				ConsumerGroupID:  "plain_text_consumer",
				ConsumerCount:    1,
			},
		},
		Logger: jsonLogger,
	}
	r := router.New()

	r.HandleFunc("plain-text-log", func(event ziggurat.Event) ziggurat.ProcessStatus {
		return ziggurat.ProcessingSuccess
	})

	processingLogger := &mw.ProcessingStatusLogger{Logger: jsonLogger}

	handler := r.Compose(processingLogger.LogStatus)

	zig := &ziggurat.Ziggurat{Logger: jsonLogger}
	<-zig.Run(context.Background(), kafkaStreams, handler)
}
```