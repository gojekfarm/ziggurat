.PHONY: all

docker.start:
	docker compose down
	docker compose up -d
	sleep 15
	docker exec {{.AppName}}_kafka /opt/kafka/bin/kafka-topics.sh --create --topic $(TOPIC_JSON) --partitions 3 --replication-factor 1 --bootstrap-server localhost:9092
	docker exec {{.AppName}}_kafka /opt/kafka/bin/kafka-topics.sh --create --topic $(TOPIC_PLAIN_TEXT) --partitions 3 --replication-factor 1 --bootstrap-server localhost:9092
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



