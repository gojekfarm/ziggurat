package cmd

var templateConfig = []ZigTemplate{
	{
		TemplateName:    "yaml-config",
		TemplatePath:    "cmd/templates/config.yaml.tpl",
		TemplateOutPath: "$APP_NAME/config/config.yaml",
	},
	{
		TemplateName:    "go-mod",
		TemplatePath:    "cmd/templates/go.mod.tpl",
		TemplateOutPath: "$APP_NAME/go.mod",
	},
	{
		TemplateName:    "main",
		TemplatePath:    "cmd/templates/main.go.tpl",
		TemplateOutPath: "$APP_NAME/main.go",
	},
	{
		TemplateName:    "docker-compose-metrics.yaml",
		TemplatePath:    "cmd/templates/docker-compose-metrics.yaml.tpl",
		TemplateOutPath: "$APP_NAME/sandbox/docker-compose-metrics.yaml",
	},
	{
		TemplateName:    "prom-conf",
		TemplatePath:    "cmd/templates/prom_conf.yaml.tpl",
		TemplateOutPath: "$APP_NAME/sandbox/prom_conf.yaml",
	},
	{
		TemplateName:    "docker-compose-kafka",
		TemplatePath:    "cmd/templates/docker-compose-kafka.yaml.tpl",
		TemplateOutPath: "$APP_NAME/sandbox/docker-compose-kafka.yaml",
	},
	{
		TemplateName:    "makefile",
		TemplatePath:    "cmd/templates/Makefile.tpl",
		TemplateOutPath: "$APP_NAME/Makefile",
	},
	{
		TemplateName:    "grafana-config",
		TemplatePath:    "cmd/templates/grafana.ini.tpl",
		TemplateOutPath: "$APP_NAME/sandbox/grafana.ini",
	},
	{
		TemplateName:    "telegraf-config",
		TemplatePath:    "cmd/templates/telegraf.conf.tpl",
		TemplateOutPath: "$APP_NAME/sandbox/telegraf.conf",
	},
	{
		TemplateName:    "datasource-grafana",
		TemplatePath:    "cmd/templates/datasource.yaml.tpl",
		TemplateOutPath: "$APP_NAME/sandbox/datasource.yaml",
	},
	{
		TemplateName:    "kafka-produce",
		TemplatePath:    "cmd/templates/produce_messages.sh.tpl",
		TemplateOutPath: "$APP_NAME/sandbox/kafka_produce.sh",
		IsExec:          true,
	},
}

func GetTemplateConfig() []ZigTemplate {
	return templateConfig
}
