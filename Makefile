.PHONY: all

all: setup-spike
topic1="test-topic1"
topic2="test-topic2"

setup-spike:
	docker-compose down
	docker-compose up -d
	sleep 10
	docker exec -it ziggurat_kafka /opt/bitnami/kafka/bin/kafka-topics.sh --create --topic $(topic1) --partitions 3 --replication-factor 1 --zookeeper ziggurat_zookeeper
	docker exec -it ziggurat_kafka /opt/bitnami/kafka/bin/kafka-topics.sh --create --topic $(topic2) --partitions 3 --replication-factor 1 --zookeeper ziggurat_zookeeper
	@echo 'Please run `go run main.go` in a new tab or terminal'
	sleep 5
	./produce_messages
#	docker-compose down
#	docker-compose rm -fv

