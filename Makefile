.PHONY: all

TOPIC_JSON="json-log"
TOPIC_PLAIN_TEXT="plain-text-log"
TEST_PACKAGES=$(shell go list ./... | grep -v -E 'cmd|logger|example|mock|mw|kafka')
TEST_PACKAGES_INTEGRATION=$(shell go list ./... | grep -E 'mw/rabbitmq|kafka')
EXAMPLE_BUILD_PKG="./example/sampleapp/main.go"
CONF_GEN = go generate -x

docker.start:
	docker-compose down
	docker-compose up -d
	sleep 10
	docker exec ziggurat_go_kafka /opt/bitnami/kafka/bin/kafka-topics.sh --create --topic $(TOPIC_JSON) --partitions 3 --replication-factor 1 --zookeeper ziggurat_go_zookeeper
	docker exec ziggurat_go_kafka /opt/bitnami/kafka/bin/kafka-topics.sh --create --topic $(TOPIC_PLAIN_TEXT) --partitions 3 --replication-factor 1 --zookeeper ziggurat_go_zookeeper
	@echo 'Please run `go run main.go` in a new tab or terminal'
	sleep 5


tidy:
	$(CONF_GEN)
	go mod tidy -v

format:
	$(CONF_GEN)
	@goimports -l -w ./

docker.start-metrics:
	docker-compose -f docker-compose-metrics.yml down
	docker-compose -f docker-compose-metrics.yml up -d
	sleep 10

lib.build:
	$(CONF_GEN)
	go build .

app.start:
	$(CONF_GEN)
	go build -o sample_app $(EXAMPLE_BUILD_PKG)
	./sample_app

lib.test:
	$(CONF_GEN)
	go test -count 1 -v $(TEST_PACKAGES)

app.start-race:
	$(CONF_GEN)
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
	$(CONF_GEN)
	./scripts/release.sh ${VERSION}

lib.test-coverage-html:
	$(CONF_GEN)
	go test -count 1 -v $(TEST_PACKAGES) -coverprofile cp.out
	go tool cover -html=cp.out

lib.test-coverage:
	$(CONF_GEN)
	go test -count 1 -v $(TEST_PACKAGES) -coverprofile cp.out
	go tool cover -func=cp.out

lib.test-integration:
	$(CONF_GEN)
	go test -count 1 -v $(TEST_PACKAGES_INTEGRATION)

