service-name: "{{.AppName}}"
stream-router:
  {{.TopicEntity}}:
    bootstrap-servers: "localhost:9092"
    instance-count: 2
    origin-topics: "plain-text-log"
    group-id: "{{.ConsumerGroup}}"
log-level: "debug"
retry:
  enabled: true
  count: 5
rabbitmq:
  host: "amqp://user:bitnami@localhost:5672/"
  delay-queue-expiration: "1000"
statsd:
  host: "localhost:8125"
http-server:
  port: 8080