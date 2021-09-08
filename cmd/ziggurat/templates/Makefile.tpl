.PHONY: all

docker.start:
	docker-compose down -d -p {{.AppName}}
	docker-compose up -d -p {{.AppName}}
	sleep 10
	docker exec ziggurat_go_kafka /opt/bitnami/kafka/bin/kafka-topics.sh --create --topic $(TOPIC_JSON) --partitions 3 --replication-factor 1 --zookeeper {{.AppName}}_zookeeper
	docker exec ziggurat_go_kafka /opt/bitnami/kafka/bin/kafka-topics.sh --create --topic $(TOPIC_PLAIN_TEXT) --partitions 3 --replication-factor 1 --zookeeper {{.AppName}}_zookeeper
	@echo 'Please run `go run main.go` in a new tab or terminal'
	sleep 5

tidy:
	go mod tidy -v

app.build:
	go build .

app.run:
	go run ./cmd/main.go

docker.cleanup:
	docker-compose down
	docker-compose rm



