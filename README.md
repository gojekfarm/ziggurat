# Ziggurat Golang

Consumer Orchestration made easy. Consume events from Kafka without hassles. Ziggurat Golang is a library that aims to simplify the consumer orchestration and lets you focus on your business logic. Define simple functions to handle events from Kafka.

<!-- TOC -->
* [Ziggurat Golang](#ziggurat-golang)
  * [Install the ziggurat CLI](#install-the-ziggurat-cli)
    * [Using with an existing application](#using-with-an-existing-application)
    * [Creating a new app using the CLI](#creating-a-new-app-using-the-cli)
    * [Building from source](#building-from-source)
    * [Running unit tests](#running-unit-tests)
    * [Running integration tests](#running-integration-tests)
    * [Coverage Report in HTML](#coverage-report-in-html)
    * [Contribution guidelines](#contribution-guidelines)
  * [Features](#features)
  * [How to consume messages from Kafka](#how-to-consume-messages-from-kafka)
  * [Configuring the `Ziggurat` struct](#configuring-the-ziggurat-struct)
    * [Ziggurat Run method](#ziggurat-run-method)
  * [Ziggurat Handler interface](#ziggurat-handler-interface)
    * [Writing custom re-usable middlewares](#writing-custom-re-usable-middlewares)
      * [A practical example](#a-practical-example)
      * [Bundled middlewares with Ziggurat-go](#bundled-middlewares-with-ziggurat-go)
  * [Ziggurat Event struct](#ziggurat-event-struct)
    * [Description](#description)
  * [Ziggurat MessageConsumer interface](#ziggurat-messageconsumer-interface)
  * [Using Kafka Consumer](#using-kafka-consumer)
    * [ConsumerConfig](#consumerconfig)
      * [Practical example on setting the `ConsumerCount` value](#practical-example-on-setting-the-consumercount-value)
    * [Events emitted by the kafka.ConsumerGroup implementation](#events-emitted-by-the-kafkaconsumergroup-implementation)
  * [How to use the ziggurat Event Router](#how-to-use-the-ziggurat-event-router)
    * [A practical example](#a-practical-example-1)
  * [Retries using RabbitMQ](#retries-using-rabbitmq)
    * [RabbitMQ Queue config](#rabbitmq-queue-config)
    * [Code sample to retry a message](#code-sample-to-retry-a-message)
    * [Events emitted by the `rabbitmq.AutoRetry`](#events-emitted-by-the-rabbitmqautoretry)
    * [How do I know if my message has been retried ?](#how-do-i-know-if-my-message-has-been-retried-)
    * [How does a queue key work?](#how-does-a-queue-key-work)
      * [A practical example](#a-practical-example-2)
  * [I have a lot of messages in my dead letter queue, how do I replay them](#i-have-a-lot-of-messages-in-my-dead-letter-queue-how-do-i-replay-them)
<!-- TOC -->

## Install the ziggurat CLI

```shell script
go install github.com/gojekfarm/ziggurat/v2/cmd/ziggurat@latest                                                                                                                                                    
```

### Using with an existing application

```shell
go get github.com/gojekfarm/ziggurat/v2
```

### Creating a new app using the CLI

```shell
ziggurat new <app_name>
go mod tidy -v  # cleans up dependencies
```
### Building from source

```shell
make lib.build
```

### Running unit tests

```shell
make lib.test
```

### Running integration tests

```shell
docker-compose up -d # starts up the RabbitMQ and Kafka containers
make lib.test-integration
```

> [!NOTE]
> There are no integration tests for Kafka, only for RabbitMQ

### Coverage Report in HTML

```shell
lib.test-coverage-html
```

> [!NOTE]
> For other make tests refer to the Makefile

### Contribution guidelines

- Avoid exposing unwanted APIs publicly
- Make sure the APIs provide certain guarantees about what they return
- Leave concurrency to the caller, an API if using concurrency internally must leave the decision to invoke it concurrently to the caller. Your APIs should be as simple as a function / method which takes in args and returns a value. Make sure to document it if an API blocks forever, like the `ziggurat.Run` function.
- Keep configuration minimum and leave the decisions on the user
- Do not write an interface first and then implement, write a struct first and add your methods, **discover interfaces do not invent them**. A premature abstraction is mostly going to be a leaky one.
- Last but not the least, Keep it simple stupid.

## Features

- Ziggurat-Go enables you to orchestrate multiple message consumers by decoupling the consumer implementation from the
  orchestration
- A small and simple API footprint
- Ziggurat Go currently supports only Kafka as a message consumer implementation
- Ziggurat Go includes a regex based router to support complex routing patterns
- Ziggurat Go provides a RabbitMQ middleware for retrying messages
- Ziggurat Go provides a RabbitMQ message consumer implementation to consume "retried" messages from RabbitMQ
- Ziggurat Go also includes a Prometheus middleware and exposes a Prometheus exporter server for instrumentation

## How to consume messages from Kafka

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

## Configuring the `Ziggurat` struct

```go
ziggurat.Ziggurat{
    Logger            StructuredLogger  // a logger implementation of ziggurat.StructuredLogger
    ShutdownTimeout  time.Duration      // wait timeout when consumers are shutdown, default value: 6 seconds
    ErrorHandler     func(err error)    // a notifier for when one of the message consumers is shutdown abruptly
}
```
> [!NOTE]
> Some message consumer implementations might not honor the context timeout/cancelation and this can cause your application to hangup when it is killed, use the shutdown timeout value to exit the application even if the message consunmers misbehave. The default value is 6 seconds.

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
> [!NOTE]
> The `Run` method returns a `ziggurat.ErrCleanShutdown` incase of a clean shutdown

## Ziggurat Handler interface

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

### Writing custom re-usable middlewares
Middlewares are a good way to run specific code before every handler is run. They provide a neat way to abstract common code which can be composed with other middlewares

Any function/Method of the signature
```go
type Middelware func(ziggurat.Handler) ziggurat.Handler
```
Can be used as a middleware in the `ziggurat.Use` function to compose middlewares

#### A practical example
I want to authenticate a certain user before I run my handler, if the auth succeeds only then I want to execute my business logic 

Code snippet
```go
type Auther interface {
	Authenticate(user string) bool
}

type GoogleAuth struct{}

func (g GoogleAuth) Authenticate(user string) bool {
	return user == "foo"
}

type AuthMiddleware struct {
	Authenticator Auther
}

func (a *AuthMiddleware) Authenticate(next ziggurat.Handler) ziggurat.Handler {
	return ziggurat.HandlerFunc(func(ctx context.Context, event *ziggurat.Event) {
		type user struct{ Username string }
		var u user
		if err := json.Unmarshal(event.Value, &u); err != nil {
			// handle error
			return // do not execute the next handler
		}
		if a.Authenticator.Authenticate(u.Username) {
			// if auth succeeds call the next handler
			next.Handle(ctx, event)
			return
		}
		// handle the failed auth
	})
}

func main() {
	var zig Ziggurat
	kcg := kafka.ConsumerGroup{
		Logger: logger.NewLogger(logger.LevelInfo),
		GroupConfig: kafka.ConsumerConfig{
			BootstrapServers: "localhost:9092",
			GroupID:          "foo.id",
			ConsumerCount:    1,
			Topics:           []string{"^.*auth-log"},
		},
	}
	
	authMW := &AuthMiddleware{Authenticator: &GoogleAuth{}}
	router := ziggurat.NewRouter()
	router.HandlerFunc("foo.id/india-auth-log$",func(ctx context.Context,e *ziggurat.Event){...})
	router.HandlerFunc("foo.id/usa-auth-log$",func(ctx context.Context,e *ziggurat.Event){...})
	router.HandlerFunc("foo.id/uk-auth-log$",func(ctx context.Context,e *ziggurat.Event){...})
	router.HandlerFunc("foo.id/aus-auth-log$",func(ctx context.Context,e *ziggurat.Event){...})
	router.HandlerFunc("foo.id/spain-auth-log$",func(ctx context.Context,e *ziggurat.Event){...})
	
	// Authenticate using the AuthMiddleware
	ziggurat.Use(router,authMW.Authenticate)
	handler := ziggurat.Use(router, authMW.Authenticate)
	_ = zig.Run(context.Background(),handler,&kcg)
	
}
```

#### Bundled middlewares with Ziggurat-go

Ziggurat Go includes two middlewares out of the box. 
- Event Logger middleware
  - The event logger middleware logs to the STDOUT whenever an event is received.
  - Usage
```go
hf := ziggurat.HandlerFunc(func(context.Context,*ziggurat.Event){...})
l := someLogger() // must implement the ziggurat.StructuredLogger interface
eventLoggerMW := event.Logger(l)
handler := ziggurat.Use(hf,eventLoggerMW)
ziggurat.Run(context.Background(),handler)
```  
- Prometheus middleware
  - The Prometheus middleware emits handler metrics using the Prometheus exporter server
  - Usage
```go
hf := ziggurat.HandlerFunc(func(context.Context,*ziggurat.Event){...})
prometheus.Register() // registers promethues registry handlers
go func() {
	prometheus.StartMonitoringServer(context.Background(),....)
}
handler := ziggurat.Use(hf,promtheues.PublishHandlerMetrics)
ziggurat.Run(context.Background(),handler)
```
To get the metrics 
```shell
curl -vv "http://localhost:9090/metrics"

ziggurat_go_handler_duration_seconds_bucket{route="<some_string_value>",le="0.05"} 228
ziggurat_go_handler_duration_seconds_bucket{route="<some_string_value>",le="0.1"} 228
ziggurat_go_handler_events_total{route="<some_string_value>"} 460
```

## Ziggurat Event struct

The `ziggurat.Event` struct is a generic event struct that is passed to the handler. This is a pointer value and should
not be modified by handlers as it is not thread safe. The struct can be cloned and modified.

### Description

```go
ziggurat.Event{
    Metadata            map[string]any     `json:"meta"` // metadata is a generic map for storing event related info
    Value               []byte             `json:"value"` // a byte slice value which contains the actual message 
    Key                 []byte             `json:"key"`   // a byte slice value which contains the actual key
    RoutingPath         string             `json:"routing_path"`       // an arbitrary string set by the message consumer implementation
    ProducerTimestamp   time.Time          `json:"producer_timestamp"` // the producer timestamp set by the message consumer implementation
    ReceivedTimestamp   time.Time          `json:"received_timestamp"` // the timestamp at which the message was ingested by the system, this is also set by the message consumer implementation
    EventType           string             `json:"event_type"`         // the type of event, ex:= kafka,rabbitmq, this is also set by the message consumer implementation
}
```

> [!NOTE]
> A note for message consumer implementations, the Metadata field is not a dumping ground for all sort of key values, it should be sparingly used and should contain only the most required fields

## Ziggurat MessageConsumer interface

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

type NumberConsumer struct {}

func (nc *NumberConsumer) Consume(ctx context.Context, h Handler) error {
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
				EventType:   "numbergen"}
			h.Handle(ctx, e)
		}
	}
}
```


## Using Kafka Consumer

### ConsumerConfig
```go
type ConsumerConfig struct {
    BootstrapServers      string // A required comma separated list of broker addresses
    DebugLevel            string // generic, broker, topic, metadata, feature, queue, msg, protocol, cgrp, security, fetch, interceptor, plugin, consumer, admin, eos, mock, assignor, conf, all
    GroupID               string // A required string 
    Topics                []string // A required non-empty list of topics to consume from
    AutoCommitInterval    int      // A commit Interval time in milliseconds
    ConsumerCount         int      // Number of concurrent consumer instances to consume from Kafka
    PollTimeout           int    // Kafka Poll timeout in milliseconds
    AutoOffsetReset       string // earliest or latest
    PartitionAssignment   string // refer partition.assignment.strategy https://github.com/confluentinc/librdkafka/blob/master/CONFIGURATION.md
    MaxPollIntervalMS     int    // Kafka Failure detection interval in milliseconds
}
```
> For more info on what the config keys and values mean please check the below link ( not all config keys are included in Ziggurat, they might be added in the future )
https://github.com/confluentinc/librdkafka/blob/master/CONFIGURATION.md

#### Practical example on setting the `ConsumerCount` value
The `ConsumerCount` value is used to control the concurrency of your handler execution, a higher value does not mean better performance, for an optimum performance please set it to the number of partitions you are consuming from.
> If you are consuming from 12 partitions using 4 individual VMs / Pods then each VM / Pod should have a ConsumerCount of 3. This adds upto 4(VM/Pods) * 3(Consumers) = 12(Consumers)
> Please follow the above rule for optimising concurrency.
> Golang Goroutines are multiplexed across multiple OS threads, ConsumerCount doesn't imply they will run in parallel.

> [!NOTE]
> We don't support manual commits at the moment, as it can lead to unwanted bugs, if need be in the future we can expose it as a feature.
> We also use the `CONSUMER` protocol and not the `STREAMS` protocol as it is not supported by the client and also since we just deal with stateless events consumption, `CONSUMER` protocol is better suited for such workloads.

### Events emitted by the kafka.ConsumerGroup implementation
```go
ziggurat.Event{
    Metadata map[string]any  // map[string]any{"kafka-partition":1,"kafka-topic":"foo-log"}
    Value    []byte         `json:"value"` // byte slice 
    Key      []byte         `json:"key"`   // byte slice
    RoutingPath       string    `json:"routing_path"`  // <consumer_group_id>/<topic_name>/<parition_num> can be used in routing
    ProducerTimestamp time.Time `json:"producer_timestamp"`  // A normal time.Time struct
    ReceivedTimestamp time.Time `json:"received_timestamp"` // A normal time.Time struct
    EventType         string    `json:"event_type"`         // kafka
}
```

## How to use the ziggurat Event Router
First of all understand if you need a router, a router is required only if you have different handlers for different type of events, if your application
just consumes from one topic, and you just want to handle all events in the same way then a router is not required, you can just pass a `ziggurat.HandlerFunc` OR a type that implements the `ziggurat.Handler` interface directly. A router lets you handle different events in a different ways by defining regex rules.
```go
ctx := context.Background()
h := ziggurat.HandlerFunc(func (context.Context, *ziggurat.Event)  {
	// handle all events 
})
groupOne := kafka.ConsumerGroup{...}
if runErr := zig.Run(ctx, h, &groupOne); runErr != nil {
    logger.Error("error running consumers", runErr)
}
```

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
router.HandlerFunc("mobile_app_log_consumer/mobile-application-logs/(1|3|5|7|9|11)$", func (ctx, *ziggurat.Event) {....})
// to execute logic for Android logs I would use this
router.HandlerFunc("mobile_app_log_consumer/mobile-application-logs/(2|4|6|8|10|12)$", func (ctx, *ziggurat.Event) {....})
```

Based on how the routing path is set by the message consumer implementation, you can define your regex patterns.

## Retries using RabbitMQ
Ziggurat-Go includes rabbitmq as the backend for message retries. Message retries are useful when message processing from one message consumer fails and needs to be retried.

The `rabbitMQ.AutoRetry(qc QueueConfig,opts ...Opts)` function creates an instance of the `rabbitmq.ARetry struct`

> [!WARNING]
> RabbitMQ publish does not use publisher confirms https://www.rabbitmq.com/tutorials/tutorial-seven-java
> Once a message is retried to RabbitMQ there is no acknowledgement received from RabbitMQ whether the message has reached the broker or not. This implies that
> the message could be lost and will not be retried if there is an error in the network layer. For most use cases it should not 
> matter. Using publishing confirms reduces publish throughput but guarantees reliability over performance.

### RabbitMQ Queue config
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

### Code sample to retry a message
```go
ar := rabbitMQ.AutoRetry(qc QueueConfig,opts ...Opts)
ar := rabbitmq.AutoRetry(rabbitmq.Queues{
		{
			QueueKey:              "foo",
			DelayExpirationInMS:   "100",
			RetryCount:            4,
			ConsumerPrefetchCount: 300,
			ConsumerCount:         10,
		},
	},
		rabbitmq.WithUsername("guest"),
		rabbitmq.WithPassword("guest"))
hf := ziggurat.HandlerFunc(func(ctx context.Context, event *ziggurat.Event) {
		err := ar.Retry(ctx, event, "foo") 
		// Retry always pushes to the delay queue
		// OR FOR A MORE GRANULAR CONTROL USE
		err := ar.Publish(ctx,event,"foo",rabbitmq.QueueTypeDLQ,"200")
		// handle error
	})
// important pass the auto retry struct as a message consumer to ziggurat.Run
zig.Run(ctx, hf, ar)
```
The `rabbitmq.AutoRetry` struct implements the `ziggurat.MessageConsumer` interface which makes it a viable candidate for consumer orchestration! Passing this to `ziggurat.Run` will consume the messages from the retry queue and feed it to your handler for re-consumption.
### Events emitted by the `rabbitmq.AutoRetry`

```go
ziggurat.Event{
    Metadata map[string]any  // map[string]any{... source key values + "rabbitmqAutoRetryCount":3}
    Value    []byte         `json:"value"` // byte slice 
    Key      []byte         `json:"key"`   // byte slice
    RoutingPath       string    `json:"routing_path"`  // <consumer_group_id>/<topic_name>/<parition_num> same as source path
    ProducerTimestamp time.Time `json:"producer_timestamp"`  // A normal time.Time struct
    ReceivedTimestamp time.Time `json:"received_timestamp"` // A normal time.Time struct
    EventType         string    `json:"event_type"`         // source path
}
```
> [!NOTE]
> The `rabbitmq.MessageConsumer` implementation does not modify the `*ziggurat.Event` struct in any way apart from storing the rabbitmq metadata, the reason being that RabbitMQ MessageConsumer is not the origin / source of the event, it is just a re-consumption of the original message.
> Message Consumer implementations should keep this in mind before modifying the event struct.

### How do I know if my message has been retried ?
```go
router.HandlerFunc("foo.id/*", func(ctx context.Context, event *ziggurat.Event) {
		if rabbitmq.RetryCountFor(event) > 0 {
			fmt.Println("message has been retried")
		} else {
			fmt.Println("new message")
		}
	})
```
The `rabbitmq.RetryCountFor` function infers the RabbitMQ metadata and provides an integer value which gives the retry count

### How does a queue key work?

#### A practical example

Suppose your queue key is called `foo_retries`. The RabbitMQ retry module will automatically create 3 queues namely 
- `foo_retries_instant_queue`
- `foo_retries_delay_queue`
- `foo_retries_dlq`
- It will also create an exchange by the name `foo_retries_exchange`. 
- This exchange is internally used to send messages to the right queue.

> [!NOTE]
> Consumption only happens from the instant queue. The delay queue is where the retried message is sent and once the retries are exhausted they are sent to the dlq.

> [!CAUTION]
> Using a Prefetch of 1 is not beneficial for consumption and can fill up the RabbitMQ queues, use a higher value from 10 to 300.

## I have a lot of messages in my dead letter queue, how do I replay them
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
| Param   | Example   | Description                                                  |
| ------- | --------- | -----------------------------------------------------------  |
| count\* | 100       | integer value                                                |
| queue\* | foo_retry | the `queue_key` **only** as specified in you rabbitmq config   |

> [!NOTE]
> \* indicates a required param

