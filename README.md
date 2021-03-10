### Ziggurat Golang

Stream Processing made easy

### Install the ziggurat CLI

```shell script
go get -v -u github.com/gojekfarm/ziggurat/cmd/...
go install github.com/gojekfarm/ziggurat/cmd/...                                                                                                                                                     
```

#### How to use

### create a new app using the `new` command

```shell
ziggurat new <app_name>
```

### Main file

```go
//+build ignore

package main

import (
	"context"
	"fmt"
	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/kafka"
	"github.com/gojekfarm/ziggurat/logger"
	"github.com/gojekfarm/ziggurat/mw"
	"github.com/gojekfarm/ziggurat/router"
)

func main() {
	jsonLogger := logger.NewJSONLogger("info")
	ctx := context.Background()

	kafkaStreams := &kafka.Streams{
		StreamConfig: kafka.StreamConfig{
			{
				BootstrapServers: "localhost:9092",
				OriginTopics:     "plain-text-log",
				ConsumerGroupID:  "plain_text_consumer",
				ConsumerCount:    1,
				RouteGroup:       "plain-text-log",
			},
			{
				BootstrapServers: "localhost:9092",
				OriginTopics:     "json-log",
				ConsumerGroupID:  "json_consumer",
				ConsumerCount:    1,
				RouteGroup:       "json-log",
			},
		},
		Logger: jsonLogger,
	}
	r := router.New()

	r.HandleFunc("plain-text-log", func(event ziggurat.Event, ctx context.Context) error {
		return nil
	})

	r.HandleFunc("json-log", func(event ziggurat.Event, ctx context.Context) error {
		return ziggurat.ErrProcessingFailed{"retry"}
	})

	handler := &mw.ProcessingStatusLogger{Logger: jsonLogger, Handler: r}

	zig := &ziggurat.Ziggurat{Logger: jsonLogger}

	if runErr := zig.Run(ctx, kafkaStreams, handler); runErr != nil {
		fmt.Println("error running streams ", runErr)
	}

}
```

### Concepts

- There are 3 main interfaces that your structs can implement to plug into the Ziggurat library

### Streamer interface

```go
package ziggurat

import "context"

type Streamer interface {
	Stream(ctx context.Context, handler Handler) chan error
}

// Any type can implement the Streamer interface to stream data from any source
```

### Handler interface

```go
package ziggurat

import "context"

type Handler interface {
	HandleEvent(event Event, ctx context.Context) error
}

// The Handler interface is very similar to the http.Handler interface
// The default router shipped with Ziggurat also implements the Handler interface
```

### Event interface

```go
package ziggurat

type Event interface {
	Value() []byte
	Headers() map[string]string
}

// Every stream must produce a series of events to be handled by the handler
// An event typically returns a byte value, headers which can contain additional metadata
```
