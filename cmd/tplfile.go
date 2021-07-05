package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

type Data struct {
	AppName          string
	TopicEntity      string
	ConsumerGroup    string
	OriginTopics     string
	BootstrapServers string
	ZigguratVersion  string
}

type ZigTemplate struct {
	TemplateName    string
	TemplateText    string
	TemplateOutPath string
	IsExec          bool
	outFile         *os.File
	t               *template.Template
}

type ZigTemplateSet struct {
	tplConfig []ZigTemplate
	basedir   string
}

func NewZigTemplateSet(basedir string, tplcfg []ZigTemplate) *ZigTemplateSet {

	return &ZigTemplateSet{
		tplConfig: tplcfg,
		basedir:   basedir,
	}
}

func (zts *ZigTemplateSet) createBaseDirectories() error {
	err := os.MkdirAll(zts.basedir+"/"+"sandbox", 0777)
	if err != nil {
		return fmt.Errorf("error creating sandbox dir: %s", err.Error())
	}
	return nil
}

func (zts *ZigTemplateSet) CreateOutFiles() error {
	if err := zts.createBaseDirectories(); err != nil {
		return err
	}

	cfg := zts.tplConfig
	for i, _ := range cfg {
		outpath := strings.ReplaceAll(cfg[i].TemplateOutPath, "$APP_NAME", zts.basedir)
		f, err := os.Create(outpath)
		if err != nil {
			return fmt.Errorf("error creating file for template %s: %s", cfg[i].TemplateName, err.Error())
		}
		cfg[i].outFile = f
	}
	return nil
}

func (zts *ZigTemplateSet) Parse() error {
	cfg := zts.tplConfig
	for i, _ := range cfg {
		t, err := template.New(cfg[i].TemplateName).Parse(cfg[i].TemplateText)
		if err != nil {
			return fmt.Errorf("error parsing template %s: %s", cfg[i].TemplateName, err.Error())
		}
		cfg[i].t = t
	}
	return nil
}

func (zts *ZigTemplateSet) Render(data Data) error {
	cfg := zts.tplConfig
	for _, z := range cfg {
		err := z.t.Execute(z.outFile, data)
		if err != nil {
			return fmt.Errorf("error rendering template %s: %s", z.TemplateName, err.Error())
		}
		if z.IsExec {
			if permErr := z.outFile.Chmod(0755); permErr != nil {
				return permErr
			}
		}
	}
	return nil
}
