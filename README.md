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
    * [Cooperative incremental rebalancing](#cooperative-incremental-rebalancing)
    * [SASL/SCRAM authentication](#saslscram-authentication)
    * [Rack-aware consumption](#rack-aware-consumption)
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

## Observability

Ziggurat ships Prometheus metrics (`mw/prometheus`), OpenTelemetry tracing (`mw/otelemetry`) and an event logger
(`mw/event`). All of it is opt-in.

```go
prometheus.Register()                      // safe to call more than once
prometheus.SetAppVersion("v1.2.3")         // ziggurat_go_app_info{version}
go prometheus.StartMonitoringServer(ctx, prometheus.WithAddr(":9090")) // default is localhost:9090

shutdown, _ := otelemetry.Setup(ctx, "my-service", "otel-collector.example.com:4317", true,
	otelemetry.WithSampleRatio(0.05),
	otelemetry.WithServiceVersion("v1.2.3"),
	otelemetry.WithEnvironment("production"),
)
defer shutdown(ctx)

handler := ziggurat.Use(router,
	otelemetry.TraceHandler("my-service"),                           // outermost: starts the span
	event.Logger(log),                                               // logs get trace_id/span_id
	prometheus.HandlerMetrics(prometheus.WithoutPartitionInRoute()), // exemplars link to the span
)
```

Metrics worth alerting on: `ziggurat_go_handler_events_total{outcome="panic"}` (a panic ends the process, so also watch
pod restarts), `ziggurat_go_kafka_consumer_lag`, `ziggurat_go_kafka_rebalance_pause_seconds`,
`ziggurat_go_kafka_revoke_commit_errors_total`, `ziggurat_go_kafka_errors_total`, `ziggurat_go_retry_dlq_published_total`
and `ziggurat_go_retry_publish_errors_total`. Trace exemplars need an OpenMetrics scrape and Prometheus exemplar storage.

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
	router := ziggurat.NewStreamRouter()
	router.HandlerFunc("booking", func(messageEvent ziggurat.MessageEvent) ziggurat.ProcessStatus {
		fmt.Println("Message -> ", messageEvent)
		return ziggurat.ProcessingSuccess
	})

	ziggurat.Start(router, ziggurat.StartupOptions{
		StartFunction: func(config ziggurat.Config) {
			fmt.Println("Start function called...")
		},
		StopFunction: func() {
			fmt.Println("Stopping app...")
		},
		Retrier: nil,
	})

}
```
 


#### TODO
- [x] Balanced Consumer groups
- [x] RabbitMQ retries
- [x] At least once delivery semantics
- [x] Retry interface
- [x] Default middleware to deserialize messages
- [x] Env vars Config override
- [x] HTTP server
- [x] Replay RabbitMQ deadset messages
- [x] Log formatting
- [ ] Metrics
- [ ] Configurable RabbitMQ consumer count
- [ ] Unit tests
