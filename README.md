### Ziggurat GO [WIP]

### Installation
- Configure git for using private modules
```shell script
git config --global url."git@source.golabs.io:".insteadOf "https://source.golabs.io/"
```
- Run go get
```shell script
go get -v -u source.golabs.io/lambda/zigg-go/ziggurat                                                                                                                                                          
```

#### How to use
- create a `config/config.yaml` in your project root
- sample `config.yaml`
```yaml
service-name: "test-app"
stream-router:
  - topic-entity: "booking"
    bootstrap-servers: "g-gojek-id-mainstream.golabs.io:6668"
    instance-count: 2
    origin-topics: "^.*-booking-log"
    group-id: "booking-group-go"
  - topic-entity: "test-entity"
    bootstrap-servers: "localhost:9092"
    instance-count: 2
    origin-topics: "test-topic2"
    group-id: "test-group2"
  - topic-entity: "json-entity"
    bootstrap-servers: "localhost:9092"
    instance-count: 1
    origin-topics: "json-test"
    group-id: "json-group"
log-level: "debug"
retry:
  enabled: true
  count: 5
rabbitmq:
  host: "amqp://user:bitnami@localhost:5672/"
  delay-queue-expiration: "1000"
```

- sample `main.go`

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
- [ ] Async offset commits
- [x] RabbitMQ retries
- [x] Retry interface
- [x] Default middleware to deserialize messages
- [x] Env vars Config override
- [ ] Replay RabbitMQ deadset messages
- [ ] Log formatting
- [ ] Configurable RabbitMQ consumer count
- [ ] Unit tests
