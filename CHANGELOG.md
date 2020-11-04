# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
