# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v2.1.0] 2026-10-17

# Added

- Observability & Telemetry foundations, Rabbitmq publisher confirms:
  - `mw/prometheus`: handler metrics now carry an `outcome` label (`success`/`panic` -
    `ziggurat.Handler.Handle` has no error return, so "error" isn't an outcome this
    middleware can observe). A panic is recorded (outcome=panic, duration up to the
    panic) and then re-panicked - this middleware observes, it never swallows, per the
    decision to revert blanket panic recovery (see the `v2.0.24` entry below).
  - `mw/prometheus`: new `ziggurat_go_retry_published_total{queue}`,
    `ziggurat_go_retry_dlq_published_total{queue}`, and
    `ziggurat_go_retry_replay_total{queue,result}` counters, wired into
    `mw/rabbitmq`'s retry/DLQ publish and DLQ-replay paths (previously uninstrumented).
    These increment unconditionally whenever `AutoRetry` publishes/replays, regardless
    of whether the application calls `prometheus.Register()` - harmless no-op memory
    cost if it doesn't, since an unregistered counter is never scraped.
  - `kafka` and `mw/rabbitmq`: native transport message headers (Kafka record headers,
    AMQP headers) are now captured into `Event.Metadata[ziggurat.MetadataKeyHeaders]`
    (`map[string]string`) - previously read by nothing. This is unconditional, not
    behind a flag: for RabbitMQ consumers specifically, a message that previously
    arrived with a `nil` `Metadata` can now arrive with a non-nil map (one `"headers"`
    key), if the producer set native AMQP headers - a real, visible change for any
    caller doing `Metadata == nil`/`len(Metadata) == 0` checks.
  - New `mw/otelemetry` package: OpenTelemetry-based distributed tracing middleware
    (`TraceHandler`), following the same `ziggurat.Use`-compatible shape as
    `mw/prometheus`/`mw/event`. Extracts/starts a span as a child of any trace context
    found in a message's headers; a panic is recorded on the span as an error and then
    re-panicked, same never-swallow contract as the metrics middleware. `Setup` wires a
    TracerProvider exporting via OTLP/gRPC (default: `AlwaysSample`, every message
    gets a span - revisit if Kafka throughput makes this too expensive). Purely
    opt-in: nothing in this package runs unless the application explicitly imports it
    and adds `TraceHandler` to its own `ziggurat.Use(...)` chain.
  - `mw/otelemetry.Setup` takes an explicit `enabled bool` as its last argument.
    Tracing is opt-in: when `enabled` is `false`, `Setup` does no OTLP dial, registers
    no TracerProvider/propagator, and returns a no-op shutdown - safe to call
    unconditionally from `main` regardless of the flag. Callers sourcing `enabled` from
    an env var (e.g. `TRACING_ENABLED`) get an off-by-default posture for free, since an
    unset var evaluates to `false`. This is a breaking change to `Setup`'s signature
    (see "Changed" below).
  - `mw/rabbitmq`: new `rabbitmq.WithPublisherConfirms(buffer)` option on `AutoRetry`
    opts in to RabbitMQ publisher confirms (previously confirmed absent - see the
    `[!WARNING]` this replaces in the README) for every publisher an `ARetry` creates:
    `Publish`, `Retry` (and the delay/DLQ republish it triggers), and DLQ replay. With
    this enabled, a broker-level nack surfaces as a returned error from `Publish`/`Retry`
    instead of being indistinguishable from a successful hand-off to the TCP connection.
    `buffer` caps in-flight confirms across the publisher pool; `0` defaults to 100.
    Off by default (adds publish latency in exchange for the stronger guarantee, so it's
    an explicit opt-in rather than a behavior change for existing callers).

# Changed

- `mw/otelemetry.Setup(ctx, serviceName, otlpEndpoint string)` is now
  `Setup(ctx, serviceName, otlpEndpoint string, enabled bool)` - update call sites to
  pass the new `enabled` argument (see "Added" above for the behavior this unlocks).
  - `mw/rabbitmq`'s retry/DLQ republish paths (`publishInternal`, `ARetry.Publish`) now
    inject the active span's trace context into the outgoing message's headers
    alongside the existing `retry-origin` header, so a republished message continues
    the same trace as its original consume instead of starting a disconnected one.
    This call is unconditional (runs on every retry/DLQ publish regardless of whether
    the app uses `mw/otelemetry`), but is a true no-op against OTel's default no-op
    propagator when tracing was never set up - no extra header content is added, and
    `mw/rabbitmq` now has a direct import dependency on `mw/otelemetry`.
  - `go.opentelemetry.io/otel` (and its SDK/OTLP-exporter submodules) added as a direct
    dependency of the root module (not split into a separate optional module - deemed
    simpler, consistent with how `mw/prometheus`'s client library is already a direct
    dependency).
  - `docker-compose-tracing.yml` + `otel-collector-config.yml`: local OTel Collector +
    Jaeger stack for development (`docker compose -f docker-compose-tracing.yml up`,
    traces visible at `http://localhost:16686`), mirroring the existing
    `docker-compose-metrics.yml` for Prometheus/Grafana.
  - Retry-attempt tracking already existed via `Event.Metadata[KeyRetryCount]` /
    `RetryCountFor(event)` (set in `mw/rabbitmq/publish.go`) - no new field was needed
    for this.
  - New rabbitmq.WithPublisherConfirms(buffer uint) option on AutoRetry.
    buffer caps how many confirms can be in flight at once across the
    publisher pool; 0 defaults to 100.
  - Applies uniformly to every publisher an ARetry creates: Publish,
    Retry (and the delay/DLQ republish it triggers), and the DLQ replay
    publisher used by DSReplayHandler.
  - Off by default - existing callers see no behavior change unless they
    opt in.
  - README's old [!WARNING] about missing publisher confirms replaced with
    a [!NOTE] documenting the new option and its latency tradeoff.
- Kafka SASL/SCRAM authentication - developed on a separate feature branch stacked on
  top of the RabbitMQ publisher confirms branch above, listed as its own item for the
  same reason:
  - `kafka.ConsumerConfig` gained `SecurityProtocol`, `SASLMechanism`, `SASLUsername`,
    `SASLPassword`, and `SSLCALocation` fields, wired into `toConfigMap()`. Lets a
    `kafka.ConsumerGroup` authenticate against a Kafka cluster via SASL (tested against
    `SCRAM-SHA-256`/`SCRAM-SHA-512`) so that broker-side ACLs configured for that
    principal actually apply - this package has no ACL logic of its own; ACLs are
    authorized by the broker, this just lets the client authenticate in the first
    place. New `kafka.SecurityProtocolSASLPlaintext`/`SecurityProtocolSASLSSL` and
    `kafka.SASLMechanismScramSHA256`/`SASLMechanismScramSHA512` constants.
  - Defaults, both overridable by setting the field explicitly: `SASLMechanism`
    defaults to `SASLMechanismScramSHA512` when `SASLUsername`/`SASLPassword` are set
    but `SASLMechanism` is left empty; `SecurityProtocol` defaults to
    `SecurityProtocolSASLPlaintext` whenever SASL is configured and `SecurityProtocol`
    is left empty (set `SecurityProtocolSASLSSL` explicitly for a cluster that
    terminates TLS).
  - No behavior change for existing callers that don't set any of these fields -
    `security.protocol`/`sasl.*` are simply absent from the generated `kafka.ConfigMap`,
    same as before.


## [v2.0.24] 2026-10-01

# Fixed

- Fixed a race in `Ziggurat.Run` where two goroutines (context-cancellation/shutdown-timeout
  handling and consumer completion) could both close or send on the same `errChan`/timeout
  channels, causing a `panic: send on closed channel` or `panic: close of closed channel`
  during shutdown
- Fixed an error-swallowing bug in `kafka.ConsumerGroup.Consume`'s worker-error aggregation,
  where a buggy `switch` matched almost any real Kafka error via unintended OR-logic and
  silently treated it as a clean shutdown instead of surfacing it
- Fixed `kafka.ConsumerGroup` creating a single Kafka consumer client shared across all
  `ConsumerCount` worker goroutines; each worker now gets its own independent consumer
  instance. The shared client meant concurrent, unsynchronized `Poll`/`Commit`/
  `StoreOffsets`/`Close` calls on a single non-thread-safe librdkafka handle — observed
  under test to cause a native `SIGSEGV` inside cgo, not just a Go-level data race

# Added

- Unit tests: `TestConsumerGroup_EachWorkerGetsOwnConsumerInstance`,
  `TestRun_DoubleCloseRace`, and a tightened `TestConsumerGroup_FatalErrorIsNotSwallowed`
  regression test

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
