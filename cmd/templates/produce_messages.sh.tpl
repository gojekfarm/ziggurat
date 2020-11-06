#!/usr/bin/env bash

BROKER_ADDRESS="{{.BootstrapServers}}"
TOPIC="{{.OriginTopics}}"

function produce() {
  NUM="$1"
  echo "key_$NUM:value_$NUM" | docker exec -i {{.AppName}}_kafka /opt/bitnami/kafka/bin/kafka-console-producer.sh \
    --broker-list "$BROKER_ADDRESS" \
    --property key.separator=":" \
    --property parse.key=true \
    --topic "$TOPIC"
}

echo "PRESS CTRL+C to terminate message production"
i=0
while true; do
  echo "producing message $i"
  produce $i
  i=$((i + 1))
  sleep 0.2s
done
