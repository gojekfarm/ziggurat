version: '3.3'

services:
  zookeeper:
    image: 'bitnami/zookeeper:latest'
    ports:
      - '2181:2181'
    container_name: '{{.AppName}}_zookeeper'
    environment:
      - ALLOW_ANONYMOUS_LOGIN=yes
  kafka:
    image: 'bitnami/kafka:2'
    ports:
      - '9092:9092'
    container_name: '{{.AppName}}_kafka'
    environment:
      - KAFKA_ZOOKEEPER_CONNECT=zookeeper:2181
      - KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092
      - ALLOW_PLAINTEXT_LISTENER=yes
  rabbitmq:
    image: 'bitnami/rabbitmq:latest'
    ports:
      - '15672:15672'
      - '5672:5672'