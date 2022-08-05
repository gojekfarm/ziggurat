module github.com/gojekfarm/ziggurat

go 1.17

require (
	github.com/cactus/go-statsd-client/v5 v5.0.0
	github.com/confluentinc/confluent-kafka-go v1.9.2
	github.com/makasim/amqpextra v0.16.4
	github.com/prometheus/client_golang v1.10.0
	github.com/rs/zerolog v1.26.0
	github.com/streadway/amqp v1.0.1-0.20200716223359-e6b33f460591
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.18.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	golang.org/x/sys v0.0.0-20211007075335-d3039528d8ac // indirect
	google.golang.org/protobuf v1.28.0 // indirect
)

// go sum hash mismatch errors caused on some CIs
retract v1.0.6

// go sum hash mismatch errors caused on some CIs need to fix the GH pipeline
retract v1.0.8
