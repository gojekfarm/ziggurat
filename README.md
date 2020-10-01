### Ziggurat GO [WIP]

### Installation

- Run go get
```shell script
go get -v -u github.com/gojek/ziggurat-golang v0.1.0                                                                                                                                                         
```

#### How to use
- create a `config/config.yaml` in your project root
- sample `config.yaml`
```yaml
service-name: "test-service"
stream-router:
  test-entity2:
    bootstrap-servers: "localhost:9092"
    instance-count: 2
    origin-topics: "test-topic1"
    group-id: "test-group"
    topic-entity: "topic-entity2"
  test-entity:
    bootstrap-servers: "localhost:9092"
    instance-count: 2
    origin-topics: "test-topic2"
    group-id: "test-group2"
    topic-entity: "test-entity"
  json-entity:
    bootstrap-servers: "localhost:9092"
    instance-count: 1
    origin-topics: "json-test"
    group-id: "json-group"
    topic-entity: "json-entity"
log-level: "debug"
retry:
  enabled: true
  count: 5
rabbitmq:
  host: "amqp://user:bitnami@localhost:5672/"
  delay-queue-expiration: "1000"
```
#### Overriding the config using ENV variables
- To override the boostrap-servers under test-entity2
```shell script
export ZIGGURAT_STREAM_ROUTER_TEST_ENTITY2_BOOTSTRAP_SERVERS="localhost:9094"
```


- sample `main.go`
### Pass an optional config file path
If you wish to read config from a location other than the default location run the app with `--config="your_path/config_file.yaml"` option

```go
package main

import (
	"fmt"
	"source.golabs.io/lambda/zigg-go/ziggurat"
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
