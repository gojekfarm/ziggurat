package cmd

import "github.com/gojekfarm/ziggurat/cmd/templates"

var templateConfig = []ZigTemplate{
	{
		TemplateName:    "go-mod",
		TemplateText:    templates.GoMod,
		TemplateOutPath: "$APP_NAME/go.mod",
	},
	{
		TemplateName:    "main",
		TemplateText:    templates.Main,
		TemplateOutPath: "$APP_NAME/main.go",
	},
	{
		TemplateName:    "docker-compose-metrics.yaml",
		TemplateText:    templates.DockerComposeMetrics,
		TemplateOutPath: "$APP_NAME/sandbox/docker-compose-metrics.yaml",
	},
	{
		TemplateName:    "prom-conf",
		TemplateText:    templates.PromConf,
		TemplateOutPath: "$APP_NAME/sandbox/prom_conf.yaml",
	},
	{
		TemplateName:    "docker-compose-kafka",
		TemplateText:    templates.DockerComposeKafka,
		TemplateOutPath: "$APP_NAME/sandbox/docker-compose-kafka.yaml",
	},
	{
		TemplateName:    "makefile",
		TemplateText:    templates.Makefile,
		TemplateOutPath: "$APP_NAME/Makefile",
	},
	{
		TemplateName:    "grafana-config",
		TemplateText:    templates.GrafanaINI,
		TemplateOutPath: "$APP_NAME/sandbox/grafana.ini",
	},
	{
		TemplateName:    "telegraf-config",
		TemplateText:    templates.TelegrafConf,
		TemplateOutPath: "$APP_NAME/sandbox/telegraf.conf",
	},
	{
		TemplateName:    "datasource-grafana",
		TemplateText:    templates.GrafanaDS,
		TemplateOutPath: "$APP_NAME/sandbox/datasource.yaml",
	},
	{
		TemplateName:    "kafka-produce",
		TemplateText:    templates.ProduceMessages,
		TemplateOutPath: "$APP_NAME/sandbox/kafka_produce.sh",
		IsExec:          true,
	},
}

func GetTemplateConfig() []ZigTemplate {
	return templateConfig
}
