# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.2] - 20202-10-04

### Changed

- Adds a condition for invalid mapper return code
- Publishes message failure metrics
- Makes StatsD configurable

## [v0.1.1] - 2020-10-03

### Changed

- Fixes module path while using `go get`

## [v0.1.0] - 2020-10-03
Initial release of Ziggurat-golang

### Added
- Kafka consumer group support
- Message retries using RabbitMQ
- At least once delivery semantics
- Override config using ENV variables
- Adds HTTP server for replaying dead set messages
- Default middleware to deserialize protobuf messages and JSON messages
- Publish StatsD style counters