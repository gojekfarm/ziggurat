.PHONY: all

TOPIC_JSON="json-log"
TOPIC_PLAIN_TEXT="plain-text-log"
BROKER_ADDRESS="localhost:9092"

docker.start-kafka:
	docker-compose -f ./sandbox/docker-compose-kafka.yaml down
	docker-compose -f ./sandbox/docker-compose-kafka.yaml up -d
	sleep 10
	docker exec -it {{.AppName}}_kafka /opt/bitnami/kafka/bin/kafka-topics.sh --create --topic $(TOPIC_PLAIN_TEXT) --partitions 3 --replication-factor 1 --zookeeper {{.AppName}}_zookeeper
	sleep 5

docker.start-metrics:
	docker-compose -f ./sandbox/docker-compose-metrics.yaml down
	docker-compose -f ./sandbox/docker-compose-metrics.yaml up -d
	sleep 10

app.start:
	go build
	./{{.AppName}}

docker.cleanup:
	docker-compose -f ./sandbox/docker-compose-kafka.yaml down
	docker-compose -f ./sandbox/docker-compose-kafka.yaml rm
	docker-compose -f ./sandbox/docker-compose-metrics.yaml down
	docker-compose -f ./sandbox/docker-compose-metrics.yaml rm