.PHONY: all

all: run-spike
topic="test-topic"
topic2="test-topic2"

setup:
	docker-compose down
	docker-compose up -d
	sleep 10
	docker exec -it go_ziggurat_kafka /opt/bitnami/kafka/bin/kafka-topics.sh --create --topic $(topic) --partitions 3 --replication-factor 1 --zookeeper kafka_zookeeper
	docker exec -it go_ziggurat_kafka /opt/bitnami/kafka/bin/kafka-topics.sh --create --topic $(topic2) --partitions 3 --replication-factor 1 --zookeeper kafka_zookeeper

run-spike: setup
	./produce_messages
