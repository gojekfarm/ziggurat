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
package main

import (
	"context"
	"fmt"
	"github.com/gojekfarm/ziggurat/router"

	"github.com/gojekfarm/ziggurat/mw/statsd"

	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/kafka"
	"github.com/gojekfarm/ziggurat/logger"
)

func main() {
	var zig ziggurat.Ziggurat
	var r = router.New()

	jsonLogger := logger.NewJSONLogger(logger.LevelInfo)
	ctx := context.Background()
	statsdPub := statsd.NewPublisher(
		statsd.WithLogger(jsonLogger),
		statsd.WithDefaultTags(statsd.StatsDTag{"app_name": "example_app"}),
	)

	kafkaStreams := kafka.Streams{
		StreamConfig: kafka.StreamConfig{
			{
				BootstrapServers: "localhost:9092",
				OriginTopics:     "plain-text-log",
				ConsumerGroupID:  "plain_text_consumer",
				ConsumerCount:    1,
				RouteGroup:       "pl-txt-log",
			},
		},
		Logger: jsonLogger,
	}

	r.HandleFunc("pl-txt-log", func(ctx context.Context, event *ziggurat.Event) error {
		fmt.Println("received message ", string(event.Value), " on partition 0")
		return nil
	})

	r.HandleFunc("pl-txt-log", func(ctx context.Context, event *ziggurat.Event) error {
		fmt.Println("received message ", string(event.Value), " on partition 1")
		return nil
	})

	zig.StartFunc(func(ctx context.Context) {
		jsonLogger.Error("error running statsd publisher", statsdPub.Run(ctx))
	})

	if runErr := zig.Run(ctx, &kafkaStreams, &r); runErr != nil {
		jsonLogger.Error("could not start streams", runErr)
	}
}
```

### Using the kafka router to set up granular routing

Note: This might not work with other stream consumers
```go
package main

import (
	"context"
	"github.com/gojekfarm/ziggurat/kafka"
)

// Declare a new router
var router kafka.Router

// HandleFunc accepts a path in the following format 
// bootstrap_server/consumer_group/topic/partition
// This pattern matches all topics that end with log but only runs for partition 0
r.HandleFunc("localhost:9092/plain_text_consumer/.*-text-log/0$", func (ctx context.Context, event *ziggurat.Event) error {
	fmt.Println("received message ", string(event.Value), " on partition 0")
	return nil
})

// This pattern matches all messages for foo_consumer
r.HandleFunc("localhost:9092/foo_consumer/", func (ctx context.Context, event *ziggurat.Event) {
	fmt.Println("received message ", string(event.Value), " for foo_consumer ")
	return nil
})
// It does a longest prefix match in-order to pick the closest matching route
```

### Concepts

- There are 2 main interfaces that your structs can implement to plug into the Ziggurat library

### Streamer interface

```go
package ziggurat

import "context"

type Streamer interface {
	Stream(ctx context.Context, handler Handler) error
}

// Any type can implement the Streamer interface to stream data from any source
```

### Handler interface

```go
package ziggurat

import "context"

type Handler interface {
	Handle(ctx context.Context, event *Event) error
}

// The Handler interface is very similar to the http.Handler interface
// The default router shipped with Ziggurat also implements the Handler interface
```



