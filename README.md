### Ziggurat Golang

Stream Processing made easy

### Install the ziggurat CLI

```shell script
go get -v -u github.com/gojekfarm/ziggurat/cmd/...
go install github.com/gojekfarm/ziggurat/cmd/...                                                                                                                                                     
```

### create a new app using the `new` command

```shell
ziggurat new <app_name>
go mod tidy -v #cleans up dependencies
```

#### How to use

```go
package main

import (
	"context"

	"github.com/gojekfarm/ziggurat"
	"github.com/gojekfarm/ziggurat/kafka"
	"github.com/gojekfarm/ziggurat/logger"
	"github.com/gojekfarm/ziggurat/mw/rabbitmq"
	"github.com/gojekfarm/ziggurat/mw/statsd"
)

func main() {
	var zig ziggurat.Ziggurat
	var r kafka.Router

	statsdPub := statsd.NewPublisher(statsd.WithDefaultTags(map[string]string{
		"app_name": "sample_app",
	}))
	ctx := context.Background()
	l := logger.NewLogger(logger.LevelInfo)

	ar := rabbitmq.AutoRetry(rabbitmq.Queues{{
		QueueName:             "pt_retries",
		DelayExpirationInMS:   "1000",
		RetryCount:            3,
		ConsumerPrefetchCount: 10,
		ConsumerCount:         10,
	}}, rabbitmq.WithUsername("user"),
		rabbitmq.WithLogger(l),
		rabbitmq.WithPassword("bitnami"))

	ks := kafka.Streams{
		StreamConfig: kafka.StreamConfig{
			{
				BootstrapServers: "localhost:9092",
				OriginTopics:     "plain-text-log",
				ConsumerGroupID:  "pt_consumer",
				ConsumerCount:    2,
			},
		},
		Logger: l,
	}

	r.HandleFunc("localhost:9092/pt_consumer/", ar.Wrap(func(ctx context.Context, event *ziggurat.Event) error {
		return ziggurat.Retry
	}, "pt_retries"))

	zig.StartFunc(func(ctx context.Context) {
		err := statsdPub.Run(ctx)
		l.Error("statsd publisher error", err)
	})

	if runErr := zig.RunAll(ctx, &r, &ks, ar); runErr != nil {
		l.Error("error running streams", runErr)
	}

}
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

### Using the RabbitMQ auto retry middleware

- Starting from ziggurat `v1.3.1` a new rabbitMQ retry middleware is included.
- Messages can be auto-retried in case of processing errors/failures.

#### How are messages retried?

Stream A ----> Handler --Retry--> RabbitMQ <br>
&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;|_____
Stream B _____|

- The rabbitmq auto retry implements the streamer interface.
- The rabbitmq auto retry exposes a Wrap method in which the handlerFunc can be wrapped and provide the queue name to
  retry with.

#### Config

```go
rabbitmq.WithLogger(loggerImpl)
rabbitmq.WithUsername("user")
rabbitmq.WithHosts("localhost:15672", "localhost-2:15672") // provide multiple hosts to dial a cluster
rabbitmq.WithPassword("bitnami")
rabbitmq.WithConnectionTimeout(10*time.Duration) // times out the connection and returns an error
```

Queue config

```go
type QueueConfig struct {
    QueueName             string   //queue to push the retried messages to 
    DelayExpirationInMS   string   //time to wait before being consumed again 
    RetryCount            int      //number of times to retry the message
    ConsumerCount         int      //number of concurrent RabbitMQ consumers
    ConsumerPrefetchCount int      //max number of messages to be sent in parallel to consumers
}
```

Example Usage

```go
r.HandleFunc("localhost:9092/pt_consumer/", ar.Wrap(func(ctx context.Context, event *ziggurat.Event) error {
    return ziggurat.Retry
}, "pt_retries"))
```

### How are messages retried?

- The handler function should be wrapped in the `Wrap` method provided by the Autoretry struct.
- Always return the `ziggurat.Retry` error for a message to be retried.
- Autoretry internally created 3 queues based on the `QueueName` in the `QueueConfig`.
  - queue_name_delay_queue : messages always get published here first, and stay here for a given time determined by the `DelayExpirationInMS` config.
  - queue_name_instant_queue : messages move to the instant queue once the delay has expired, waiting to be consumed. Once the messages are consumed they are processed by the handler.
  - queue_name_dlq_queue : messages move here when the retry count is exhausted, `RetryCount` config.
- You can have as many consumers as you wish, this value can be tweaked based on you throughput and your machine's capacity. This can be tweaked using the `ConsumerCount` config. 
