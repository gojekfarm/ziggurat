.PHONY: all

all: setup-spike
topic="test-topic"
topic2="test-topic2"

setup-spike:
	docker-compose down
	docker-compose up -d
	sleep 10
	docker exec -it ziggurat_kafka /opt/bitnami/kafka/bin/kafka-topics.sh --create --topic $(topic) --partitions 3 --replication-factor 1 --zookeeper ziggurat_zookeeper
	docker exec -it ziggurat_kafka /opt/bitnami/kafka/bin/kafka-topics.sh --create --topic $(topic2) --partitions 3 --replication-factor 1 --zookeeper ziggurat_zookeeper
	sleep 1
	./produce_messages
	docker-compose down
	docker-compose rm -fv

