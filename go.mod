module github.com/gojekfarm/ziggurat

go 1.17

require (
	github.com/cactus/go-statsd-client/v5 v5.0.0
	github.com/confluentinc/confluent-kafka-go v1.7.0
	github.com/julienschmidt/httprouter v1.3.0
	github.com/makasim/amqpextra v0.16.4
	github.com/prometheus/client_golang v1.10.0
	github.com/rs/zerolog v1.19.0
	github.com/streadway/amqp v1.0.1-0.20200716223359-e6b33f460591
)

// go sum hash mismatch errors caused on some CIs
retract v1.0.6

// go sum hash mismatch errors caused on some CIs need to fix the GH pipeline
retract v1.0.8
