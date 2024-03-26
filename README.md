### Ziggurat Golang

Consumer Orchestration made easy

### Install the ziggurat CLI

```shell script
go install github.com/gojekfarm/ziggurat/v2/cmd/ziggurat@latest                                                                                                                                                    
```

### I already have an application and I just want to use Ziggurat Go without scaffolding a new app
#### Run the following command in your go application directory

```shell
go get github.com/gojekfarm/ziggurat/v2
```

### create a new app using the `new` command

```shell
ziggurat new <app_name>
go mod tidy -v  # cleans up dependencies
```

### Features

- Ziggurat-Go enables you to orchestrate multiple message consumers by decoupling the consumer implementation from the
  orchestration
- A small and simple API footprint
- Ziggurat Go currently supports only Kafka as a message consumer implementation
- Ziggurat Go includes a regex based router to support complex routing patterns
- Ziggurat Go provides a RabbitMQ middleware for retrying messages
- Ziggurat Go provides a RabbitMQ message consumer implementation to consume "retried" messages from RabbitMQ
- Ziggurat Go also includes a Prometheus middleware and exposes a Prometheus exporter server for instrumentation

### How to consume messages from Kafka

```go
package main

import (
	"context"
	"github.com/gojekfarm/ziggurat/v2"
	"github.com/gojekfarm/ziggurat/v2/kafka"
	"github.com/gojekfarm/ziggurat/v2/logger"
)

func main() {
	var zig ziggurat.Ziggurat
	router := ziggurat.NewRouter()

	ctx := context.Background()
	l := logger.NewLogger(logger.LevelInfo)

	kcg := kafka.ConsumerGroup{
		Logger: nil,
		GroupConfig: kafka.ConsumerConfig{
			BootstrapServers: "localhost:9092",
			GroupID:          "foo.id",
			Topics:           []string{"foo"},
		},
	}

	router.HandlerFunc("foo.id/*", func(ctx context.Context, event *ziggurat.Event)  {
		
	})

	h := ziggurat.Use(router)

	if runErr := zig.Run(ctx, h, &kcg); runErr != nil {
		l.Error("error running consumers", runErr)
	}

}
```

### Configuring the `Ziggurat` struct

```go
ziggurat.Ziggurat{
    Logger       StructuredLogger // a logger implementation of ziggurat.StructuredLogger
    WaitTimeout  time.Duration    // wait timeout when consumers are shutdown
    ErrorHandler func(err error) // a notifier for when one of the message consumers is shutdown abruptly
}
```

> [!NOTE]  
> The zero value of `ziggurat.Ziggurat` is perfectly valid and can be used without any issues

### Ziggurat Run method

The `ziggurat.Run` method is used to start the consumer orchestration. It takes in a `context.Context` implementation,
a `ziggurat.Handler` and a variable number of message consumer implementations.

```go
ctx := context.Background()
h := ziggurat.HandlerFunc(func (context.Context, *ziggurat.Event)  {...})
groupOne := kafka.ConsumerGroup{...}
groupTwo := kafka.ConsumerGroup{...}
if runErr := zig.Run(ctx, h, &groupOne, &groupTwo); runErr != nil {
    logger.Error("error running consumers", runErr)
}
```

### Ziggurat Handler interface

The `ziggurat.Handler` is an interface for handling ziggurat events, an event is just something that happens in a finite
timeframe. This event can come from
any source (kafka,redis,rabbitmq). The handler's job is to handle the event, i.e... the handler contains your
application's business logic

```go
type Handler interface {
    Handle(ctx context.Context, event *Event) 
}
type HandlerFunc func (ctx context.Context, event *Event)  // serves as an adapter for normal functions to be used as handlers
```

> Any function / struct which implements the above handler interface can be used in the ziggurat.Run method. The
> ziggurat.Router also implements the above interface.

### Ziggurat Event struct

The `ziggurat.Event` struct is a generic event struct that is passed to the handler. This is a pointer value and should
not be modified by handlers as it is not thread safe. The struct can be cloned and modified.

#### Description

```go
ziggurat.Event{
    Metadata map[string]any `json:"meta"` // metadata is a generic map for storing event related info
    Value    []byte         `json:"value"` // a byte slice value which contains the actual message 
    Key      []byte         `json:"key"`   // a byte slice value which contains the actual key
    RoutingPath       string    `json:"routing_path"`       // an arbitrary string set by the message consumer implementation
    ProducerTimestamp time.Time `json:"producer_timestamp"` // the producer timestamp set by the message consumer implementation
    ReceivedTimestamp time.Time `json:"received_timestamp"` // the timestamp at which the message was ingested by the system, this is also set by the message consumer implementation
    EventType         string    `json:"event_type"`         // the type of event, ex:= kafka,rabbitmq, this is also set by the message consumer implementation
}

```

### ziggurat.MessageConsumer interface

The `ziggurat.MessageConsumer` interface is the interface used for implementing message consumers which can be passed to
the ziggurat.Run method for orchestration.

```go
type MessageConsumer interface {
    Consume(ctx context.Context, handler Handler) error
}
```

The `kafka.ConsumerGroup` and `rabbitmq.AutoRetry` implement the above interface.

A sample implementation which consumes infinite numbers
```go
type NumberConsumer struct {
counter int
}

func (nc *NumberConsumer) Consume(ctx context.Context, h Handler) error{
    var i int64
    for {
        select {
		    case <-ctx.Done():
                return ctx.Err()
            default:
                time.Sleep(1000 * time.Millisecond)
                e := &Event{
						Value:       strconv.AppendInt(make([]byte, 8), i, 10),
						Key:         strconv.AppendInt(make([]byte, 8), i, 10),
						RoutingPath: "numpath",
						EventType:   "numbergen",}	
				h.Handle(ctx, e)
			}
		}
}
```


## Kafka consumer configuration

```go
type ConsumerConfig struct {
    BootstrapServers      string // A required comma separated list of broker addresses
    DebugLevel            string // generic, broker, topic, metadata, feature, queue, msg, protocol, cgrp, security, fetch, interceptor, plugin, consumer, admin, eos, mock, assignor, conf, all
    GroupID               string // A required string 
    Topics                []string // A required non-empty list of topics to consume from
    AutoCommitInterval    int      // A commit Interval time in milliseconds
    ConsumerCount         int      // Number of concurrent consumer instances to consume from Kafka
    PollTimeout           int    // Kafka Poll timeout in milliseconds
    RouteGroup            string // An optional route group to use for routing purposes
    AutoOffsetReset       string // earliest or latest
    PartitionAssignment   string // refer partition.assignment.strategy https://github.com/confluentinc/librdkafka/blob/master/CONFIGURATION.md
    MaxPollIntervalMS     int    // Kafka Failure detection interval in milliseconds
}
```

> For low level details please check For full details please check
https://github.com/confluentinc/librdkafka/blob/master/CONFIGURATION.md

## How to use the router

A router enables you to handle complex routing problems by defining handler functions for predefined regex paths.

### A practical example

I am consuming from the kafka topic `mobile-application-logs` which has 12 partitions. All the even partitions contain
logs for android devices and all the odd partitions contain logs for iOS devices. I want to execute different logic for
logs from different platforms.

The `ziggurat.Event` struct contains a field called `RoutingPath` this field is set by the `MessageConsumer`
implementation, the Kafka implementation uses the following format

```text
<consumer_group_id>/<topic_name>/<partition> 
ex: mobile_app_log_consumer/mobile-application-logs/1
```

```go
router := ziggurat.NewRouter()
// to execute logic for iOS logs I would use this
router.HandlerFunc("mobile_app_log_consumer/mobile-application-logs/(1|3|5|7|9|11)", func (ctx, *ziggurat.Event) error {....})
// to execute logic for Android logs I would use this
router.HandlerFunc("mobile_app_log_consumer/mobile-application-logs/(2|4|6|8|10|12)", func (ctx, *ziggurat.Event) error {....})
```

Based on how the routing path is set by the message consumer implementation, you can define your routing paths.

### Retries
Ziggurat-Go includes rabbitmq as the backend for message retries. Message retries are useful when message processing from one message consumer fails and needs to be retried.

The `rabbitMQ.AutoRetry(qc QueueConfig,opts ...Opts)` function creates an instance of the `rabbitmq.ARetry struct`

Required params
```go
type QueueConfig struct {
	QueueKey              string  // A queue key to be used for retries, any arbitrary string would do
	DelayExpirationInMS   string  // The time in milliseconds after which to reconsume the message for processing
	RetryCount            int     // The number of times to retry the message
	ConsumerPrefetchCount int     // The number of messages in batch to be consumed from RabbitMQ 
	ConsumerCount         int     // Number of concurrent consumer instances
}

type Queues []QueueConfig
```
#### How does queue key work
A practical example
Suppose your queue key is called `foo_retries`. The RabbitMQ retry module will automatically create 3 queues namely 
- `foo_retries_instant_queue`
- `foo_retries_delay_queue`
- `foo_retries_dlq`
It will also create an exchange by the name `foo_retries_exchange`. This exchange is internally used to send messages to the right queue.
Consumption only happens from the instant queue. The delay queue is where the retried message is sent and once the retries are exhausted they are sent to the dlq.

#### I have a lot of messages in my dead letter queue, how do I replay them
The RabbitMQ package provides HTTP handlers which clear the messages on the RabbitMQ queues. These handlers can be used with any HTTP server.

Example usage:
```go
router := http.NewServeMux()
ar := rabbitmq.AutoRetry()
router.Handle("POST /ds_replay", ar.DSReplayHandler(context.Background()))
router.Handle("POST /ds_view", ar.DSViewHandler(context.Background()))
http.ListenAndServe("localhost:8080", router)
```
Just invoke the API with the following query params
| Param   | Example   |
| ------- | --------- |
| count\* | 100       |
| queue\* | foo_retry |

> [!NOTE]
> \* indicates a required param

> [!CAUTION]
> Using a Prefetch of 1 is not beneficial for consumption and can fill up the RabbitMQ queues, use a higher value from 10 to 300.
