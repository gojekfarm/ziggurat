# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Ziggurat Go (`github.com/gojekfarm/ziggurat/v2`) is a Go library for orchestrating message consumers (Kafka, RabbitMQ) and routing events to handler functions. The CLI lives in `cmd/ziggurat`; a runnable sample is in `example/sampleapp`.

## Commands

- Build: `make lib.build` (`go build ./...`)
- Unit tests: `make lib.test`. It skips `cmd`, `logger`, `example`, `mock`, `mw` and `kafka` packages (see `TEST_PACKAGES` in the Makefile).
- Integration tests: `make docker.start` (RabbitMQ + Kafka via docker compose), then `make lib.test-integration` (`mw/rabbitmq` and `kafka` packages).
- Single test: `go test -count 1 -run TestName ./path/to/pkg`
- Race test of the sample app: `make app.start-race`
- Format: `make format` (needs `goimports`)
- Release: `make pkg.release VERSION=vX.Y.Z` (pushes master and tags). Update `CHANGELOG.md` per release.
- `.local-test*` and `.local-notes` are git-ignored scratch harnesses and notes, not part of the library.

## Architecture

Core flow: `MessageConsumer.Consume(ctx, Handler)` pulls messages and calls `Handler.Handle(ctx, *Event)`.

- Root package `ziggurat`:
  - `Ziggurat.Run` ([zigg.go](zigg.go)) runs every `MessageConsumer` in its own goroutine against one shared `Handler`. It blocks until all finish. On ctx cancel it waits up to `ShutdownTimeout` (default 6s). It returns `ErrCleanShutdown` when nothing failed, so treat that as success.
  - `Router` ([router.go](router.go)) is a `Handler` that matches `Event.RoutingPath` against registered regex patterns. An exact-string match wins. Otherwise the longest pattern is tried first. Duplicate registration panics.
  - `Middleware` is `func(Handler) Handler`, composed with `ziggurat.Use`.
  - Routing path format is `<group>/<topic>/<partition>` (see `constructPath` in [kafka/process.go](kafka/process.go)). Router patterns are regexes against it.
- `kafka/`: `ConsumerGroup` built on confluent-kafka-go (librdkafka, so CGO).
  - It creates `ConsumerCount` independent librdkafka clients, one per `worker`. A client handle is not thread-safe, so never share one across workers.
  - `ConsumerGroup.GroupConfig.Validate()` runs first and fails fast on bad SASL config.
  - Supports the cooperative-sticky assignor, rack awareness and SASL/SCRAM. The rebalance callback is exercised by `rebalance_test.go`.
  - `ConsumerHandle` ([kafka/chandle.go](kafka/chandle.go)) exposes only the first worker's client.
- `mw/`: optional middlewares and consumers.
  - `rabbitmq`: `AutoRetry` publishes events to delay/instance/dead-letter queues (retry) and acts as a `MessageConsumer` that replays them. It also has DLQ view/replay HTTP handlers.
  - `prometheus`: metrics middleware plus exporter server.
  - `otelemetry`: tracing and setup.
  - `event`: logging middleware.
- `logger/`: `StructuredLogger` implementations (JSON, NOOP). `Ziggurat.Logger` defaults to NOOP.

## Conventions (from README)

- Keep the public API small. Leave concurrency decisions to the caller and document anything that blocks forever.
- Do not define an interface first. Write the struct and discover interfaces later.
- Keep configuration minimal.
