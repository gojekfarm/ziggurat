.PHONY: all

all: setup-spike
topic1="test-topic1"
topic2="test-topic2"
topic3="json-topic"

setup-spike:
	docker-compose down
	docker-compose up -d
	sleep 10
	docker exec -it ziggurat_go_kafka /opt/bitnami/kafka/bin/kafka-topics.sh --create --topic $(topic1) --partitions 3 --replication-factor 1 --zookeeper ziggurat_go_zookeeper
	docker exec -it ziggurat_go_kafka /opt/bitnami/kafka/bin/kafka-topics.sh --create --topic $(topic2) --partitions 3 --replication-factor 1 --zookeeper ziggurat_go_zookeeper
	docker exec -it ziggurat_go_kafka /opt/bitnami/kafka/bin/kafka-topics.sh --create --topic $(topic3) --partitions 1 --replication-factor 1 --zookeeper ziggurat_go_zookeeper
	@echo 'Please run `go run main.go` in a new tab or terminal'
	sleep 5


