# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v2.0.21] 2024-03-25

- Manually commit uncommitted offsets before closing the Kafka Consumer

## [v2.0.18] 2024-03-25

# Removed

- StatsD client
- kafka.Streams

# Changes

- Internal interface changes for reliablity
- Improved test cases
- Smaller API surface area

# Breaking changes

- Streams struct is replaced by kafka.MessageConsumer
- Streams interface is replaced by MessageConsumer interface
- StatsD package has been removed
- Go 1.22 is the min version required
- ziggurat.RunAll has been removed
- ziggurat.StartFunc and StopFunc have been removed
- ziggurat.Event struct has been cleaned up
- ziggurat.Handler does not return an error anymore

## [v1.7.4] 2023-11-01

# Changes

- Upgrades Confluent Kafka client to `v2.2.0`
- Fixes RabbitMQ connection URL bug

## [1.7.1] 2022-08-05

# Changes

- Update Update github.com/confluentinc/confluent-kafka-go
  to [v1.9.2](https://github.com/confluentinc/confluent-kafka-go/releases/tag/v1.9.2)

## [1.7.0] 2022-07-12

# Changes

- Update github.com/confluentinc/confluent-kafka-go
  to [v1.9.1](https://github.com/confluentinc/confluent-kafka-go/releases/tag/v1.9.1)
- Built-in support for Mac OSX M1 / arm64.

## [1.6.5] 2022-05-11

# Changes

- Fixes goroutine publisher loop in statsd

## [1.6.3] 2022-05-10

# Changes

- Fixes goroutine leak in statsd publisher

## [1.6.2] 2022-05-10

# Changes

- Uses RabbitMQ publisher pooling for publish operations. This significantly improves the publish throughput

## [1.6.1] 2022-02-17

# Added

- The function `rabbitmq.GetRetryCount` is now exposed and can be used to get the retry count without having to do
  manual assertions

## [1.6.0] 2022-02-13

# Changes

- changes internal retry structure for RabbitMQ

# Breaking Changes

- The RabbitMQ `QueueConfig` now uses `QueueKey` instead of `QueueName`

## [1.5.1] 2022-01-28

# Added

- consumer config now takes in a new option called `PartititionAssignment` which is used to configure how kafka
  partitions are assigned to a consumer

## [1.5.0] 2022-01-07

# Changes

- Kafka consumer default `PollTimeout` is `100 ms`
- `ConsumerGroupID` is renamed to `GroupID` and `OriginTopics` is renamed to `Topics` to keep it consistent with kafka
  terminology
- The run loop for kafka consumer polling has been restructured.

# Removed

- The server package only exposes a Run method which orchestrates an HTTP Server rather than providing a full web server
  implementation
- `httprouter` dependency has been removed

## [1.4.5] 2021-12-16

# Changes

- Exposes the ARetry struct from RabbitMQ

## [1.4.4] 2021-12-16

# Added

- RabbitMQ MW adds a header `retry-origin:ziggurat-go` to identify the message retry source
- Adds a `WithHandler` option function inside the server package to pass in custom handlers
- Adds a deprecation comment for `ConfigureHandler` and `ConfigureHTTPEndpoints` as this is implementation specific
- Internal code refactoring
- Exposes a close method on the statsd client to prevent early closes due to context cancellation
- StatsD Run method takes in an Option to configure the goroutine publish interval. a negative value will not publish
  goroutines

## [1.4.3] 2021-10-20

# Changes

- fixes go-routine leak in the Dead set replay handler

## [1.4.2] 2021-10-19

# Breaking changes

- kafka messages now use the route group to route messages, `<route_group>/<topic>/<partition>`. Usage
  of `<bootstrap_server>/<consumer_group>/<topic>/<partition>` is deprecated as this makes the path very long and routes
  become difficult to configure.

## [1.4.1] 2021-09-09

# Breaking changes

- `Compose` methods are removed from the router API(s)

# Changes

- uses `go v1.17`
- various bug fixes and panic fixes
- more descriptive errors
- RabbitMQ's consumers do not use go-routine based workers anymore
- uses go embed to embed templates to render
- RabbitMQ WorkerCount is removed from QueueConfig
- Uses sync.Once to initialize publishers in the Wrap method

# Added

- integration tests for kafka streams and RabbitMQ
- `ziggurat.Use` API to compose middleware
- all packages use the `logger.Noop` as the default implementation for logging
- `rabbitmq.WithConnectionTimeout` to specify a timeout for queue creation
- Parallelize RabbitMQ consumption using the `ConsumerCount`

## [1.3.5] 2021-08-27

- adds a rabbitmq dead set replay handler
- fixes RabbitMQ panic

## [1.3.2] 2021-08-25

# Added

- RabbitMQ middleware for auto retrying messages
- A human friendly logger

# Changes

- Event now includes a metadata map of type `map[string]interface{}`
- kafka key values are copied to a new slice before passing it on
- statsD middleware no longer sends the `route` label
- loggers use the `logger.DiscardLogger` as the default logger

## [1.2.2] 2021-07-06

# Changes

- Adds headers to maintain backwards compatibility

## [1.2.0] 2021-07-05

# Added

- Regex based kafka router

# Changes

- Moves PipeHandlers to util pkg

## [1.1.2] 2021-06-01

# Added

- Makes Prometheus server port configurable

## [1.1.1] 2021-05-31

# Changes

- Fixes Prometheus server port

# Added

- An error for streams clean shutdown

## [1.0.9] 2021-05-19

# Changes

- fixes parent context propagation in the `ziggurat.RunAll` method

## [1.0.9] 2021-05-14

# Changes

- retract v1.0.8 due to tagging in pipelines

## [1.0.8] 2021-05-03

# Changes

- fixes statsD middleware handler metrics publish unit

## [1.0.6] 2021-05-02

# Added

- statsD middleware now publishes a generic event delay

## [1.0.5] 2021-05-02

# Changes

- statsD middleware only publishes kafka lag and not all events lag
- fixes handler execution calculation in the statsD middleware

## [1.0.4] 2021-04-24

# Changes

- Handler signature now uses the `*Event` as the second arg
- Includes a `RunAll` method to run multiple streams
- The Run methods returns an error and blocks, leaving concurrency to the caller
- Kafka polling uses the `Poll` method instead of the  `readMessage` method

# Added

- statsD middleware
- Prometheus middleware
- logging middleware

## [1.0.0-alpha.16] 2021-02-23

# Added

- an error for processing failed

## [1.0.0-alpha.15] 2021-02-19

# Changes

- streams kafka logs using channels
- removes rabbitmq middleware
- removes statsd middleware
- handler interface is expected to return an error
- changes the handler signature
- moves http-server to a new package
- renames the command from `zig` to `ziggurat`

# Added

- interface for structured logging
- streamer interface
- event interface

## [0.9.0] 2020-12-01

# Added

- consumers process messages concurrently
- Smaller app API
- Middleware composition

## Changes

- Refines interface methods
- fixes tests

## [0.8.0] 2020-11-18

## Added

- support for rabbitmq clusters
- a safer amqp implementation which eliminates the race conditions while re-connecting
- improves test coverage

## Changes

- reorganised packages
- provides a message decode hook on MessageEvent

## [0.7.1] 2020-11-07

## Changes

- fixes templates

## [0.7.0] 2020-11-07

## Added

- updates the `kakfa-client` lib

## Changes

- uses `http.Handler` to configure http-routes
- updates main template
- defines an interface for App

## [0.6.1] 2020-11-06 UNRELEASED

## Changes

- fixes template not found bug
- uses raw strings instead of tempalte files

## [v0.6.0] 20202-11-06

## Added

- adds a sandbox to test out metric publishing,kafka production and retry flow

## Changes

- fixes CLI bugs

## [v0.5.2] 2020-11-04

## Changes

- updates ziggurat version in the CLI

## [v0.5.1] 2020-11-04

## Changes

- fixes bug in CLI where `<-` was being escaped

## [v0.5.0] 2020-11-04

## Added

- adds new make tasks
- CLI now generates new compose files for sandbox-ing

### Changes

- renames the `MessageRetrier` interface to `MessageRetry`
- exports `handlerFn` in the `topicEntity` struct
- removes `rabbitmq.go`
- uses a thread-safe rabbitmq implementation
- runs app in async mode

## [v0.4.4] 2020-10-28

### Changes

- fixes CLI related issues

## [v0.4.0] 2020-10-27

### Added

- Adds a zig CLI to scaffold apps

## [v0.3.2] 2020-10-25

### Changes

- fixes httpserver not starting up

## [v0.3.1] 2020-10-25

### Changes

- fixes httpserver bug which caused configured routes to be lost

## [v0.3.0] 2020-10-25

### Added

- Adds a new middleware to publish message metrics
- Adds thread safety to `MessageAttributes` in `MessageEvent`
- Adds a metric to measure the `handlerFunc` exec time
- Adds a make task to start the app
- Adds a make task to start-up the metrics containers
- Adds a make task to produce messages to kafka

### Changes

- halts the app if retries are disabled, and a message is retried
- changes the `--zigurat-config` to `-config`
- `app.Run` accepts a `zig.RunOptions` type
- fixes race conditions when starting-up the app
- uses constructors functions to initialize components

## [v0.2.0] 2020-10-20

### Added

- Sends `app_name` as a tag in StatsD

### Changes

- `RabbitRetrier` uses app context to exit the replay delivery loop
- App components are not mutable
- Disables Kafka broker logs

## [v0.1.6] - 2020-10-19

### Added

- Add tests for app
- Add tests for middleware

### Changes

- Passes all middleware args by value
- Fixes race condition in pipe handlers
- Makes stop function a part of the `app.Run` method

## [v0.1.5] - 2020-10-16

### Added

- Add tests for util

### Changes

- Fixes middleware execution order
- Fixes log formatting
- Disables colored output for logs

## [v0.1.4] - 2020-10-15

### Added

- Adds function to DefaultHTTPServer to attach routes
- Adds ping endpoint to DefaultHTTPServer

## [v0.1.3] - 2020-10-14

### Changes

- Retrier interface Start method returns a channel to wait on
- Rabbit Retrier's consumer polling is moved into Start method

## [v0.1.1] - 2020-10-12

### Changes

- Fixes go mod module path

## [v0.1.0] - 2020-10-12

Initial release of Ziggurat-golang

### Added

- Kafka consumer group support
- Middleware support
- Message retries using RabbitMQ
- At least once delivery semantics
- Override config using ENV variables
- Adds HTTP server for replaying dead set messages
- Default middleware to deserialize protobuf messages and JSON messages
- Publish StatsD style counters
