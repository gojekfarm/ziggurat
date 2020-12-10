package zretry

import (
	"strings"
)

type RabbitMQConfig struct {
	hosts                string
	delayQueueExpiration string
	dialTimeoutInS       int
	queuePrefix          string
	count                int
}

func splitHosts(hosts string) []string {
	return strings.Split(hosts, ",")
}
