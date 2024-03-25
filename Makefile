.PHONY: all

TOPIC_JSON="json-log"
TOPIC_PLAIN_TEXT="plain-text-log"
TEST_PACKAGES=$(shell go list ./... | grep -v -E 'cmd|logger|example|mock|mw|kafka')
TEST_PACKAGES_INTEGRATION=$(shell go list ./... | grep -E 'mw/rabbitmq|kafka')
EXAMPLE_BUILD_PKG="./example/sampleapp/main.go"

docker.start:
	docker-compose down
	docker-compose up -d
	sleep 10
	docker exec ziggurat_go_kafka /opt/bitnami/kafka/bin/kafka-topics.sh --create --topic $(TOPIC_JSON) --partitions 3 --replication-factor 1 --zookeeper ziggurat_go_zookeeper
	docker exec ziggurat_go_kafka /opt/bitnami/kafka/bin/kafka-topics.sh --create --topic $(TOPIC_PLAIN_TEXT) --partitions 3 --replication-factor 1 --zookeeper ziggurat_go_zookeeper
	@echo 'Please run `go run main.go` in a new tab or terminal'
	sleep 5


tidy:
	go mod tidy -v

format:
	@goimports -l -w ./

docker.start-metrics:
	docker-compose -f docker-compose-metrics.yml down
	docker-compose -f docker-compose-metrics.yml up -d
	sleep 10

lib.build:
	go build ./...

app.start:
	go build -o sample_app $(EXAMPLE_BUILD_PKG)
	./sample_app

lib.test:
	go test -count 1 -v $(TEST_PACKAGES)

app.start-race:
	
	go build -race -o sample_app $(EXAMPLE_BUILD_PKG)
	./sample_app

docker.cleanup:
	docker-compose down
	docker-compose rm
	docker-compose -f docker-compose-metrics.yml down
	docker-compose -f docker-compose-metrics.yml rm

docker.kafka-produce:
	./scripts/produce_messages

pkg.release:
	./scripts/release.sh ${VERSION}

lib.test-coverage-html:
	go test -count 1 -v $(TEST_PACKAGES) -coverprofile cp.out
	go tool cover -html=cp.out

lib.test-coverage:
	go test -count 1 -v $(TEST_PACKAGES) -coverprofile cp.out
	go tool cover -func=cp.out

lib.test-integration:
	go test -count 1 -v $(TEST_PACKAGES_INTEGRATION)

