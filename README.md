### Ziggurat Golang

Stream Processing made easy

### Install the ziggurat CLI

```shell script
go get -v -u github.com/gojekfarm/ziggurat/cmd/...
go install github.com/gojekfarm/ziggurat/cmd/...                                                                                                                                                     
```

#### How to use

- create a new app using the `new` command

```shell
ziggurat new <app_name>
```

- sample `main.go`

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
		StreamConfig: kafka.StreamConfig{
			{
				BootstrapServers: "localhost:9092",
				OriginTopics:     "plain-text-log",
				ConsumerGroupID:  "plain_text_consumer",
				ConsumerCount:    1,
				RouteGroup:       "plain-text-log",
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