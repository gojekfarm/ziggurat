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
}

func GetTemplateConfig() []ZigTemplate {
	return templateConfig
}
