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
	"github.com/gojekfarm/ziggurat/mw"
)

const RoutePlainTextLog = "plain-text-log"

func main() {
	app := &ziggurat.Ziggurat{}
	router := ziggurat.NewRouter()
	statusLogger := mw.NewProcessingStatusLogger()

	router.HandleFunc(RoutePlainTextLog, func(event ziggurat.Message, ctx context.Context) ziggurat.ProcessStatus {
		return ziggurat.ProcessingSuccess
	})

	handler := router.Compose(statusLogger.LogStatus)

	app.StartFunc(func(ctx context.Context) {

	})

	<-app.Run(context.Background(), handler,
		ziggurat.StreamRoutes{
			RoutePlainTextLog: {
				BootstrapServers: "localhost:9092",
				OriginTopics:     "plain-text-log",
				ConsumerGroupID:  "plain_text_consumer",
				ConsumerCount:    2,
			},
		})
}
```