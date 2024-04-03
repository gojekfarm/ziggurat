package main

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"os"
	"strings"
)

func die(err error) {
	if err != nil {
		fmt.Println("command failed with error(s):")
		fmt.Println(err.Error())
		os.Exit(1)
	}

}

//go:embed templates/*
var res embed.FS
var Version = "v2.0.20"

var definePaths = func(basePath string) map[string]string {
	return map[string]string{
		"main.go":            fmt.Sprintf("%s/%s", basePath, "cmd"),
		"go.mod":             basePath,
		"docker-compose.yml": basePath,
		"Makefile":           basePath,
	}
}

type Data struct {
	AppName string
	Version string
}

func main() {
	args := os.Args[1:]
	fmt.Println("CLI version:", Version)
	usage := `USAGE:
  ziggurat new <app_name>`
	if len(args) < 2 {
		die(errors.New(usage))
	}

	appName := args[1]
	d := Data{AppName: appName, Version: Version}
	wd, err := os.Getwd()
	die(err)

	paths := definePaths(wd + "/" + appName)

	t, err := template.ParseFS(res, "templates/*")
	die(err)

	fileInfos, err := res.ReadDir("templates")
	die(err)

	for _, fi := range fileInfos {
		fileName := strings.TrimSuffix(fi.Name(), ".tpl")
		dirPath, ok := paths[fileName]
		if !ok {
			die(fmt.Errorf("path not found for [%s]", fileName))
		}
		err = os.MkdirAll(dirPath, 0755)
		die(err)
		f, err := os.Create(fmt.Sprintf("%s/%s", dirPath, fileName))
		die(err)
		err = t.ExecuteTemplate(f, fi.Name(), d)
		die(err)
	}
	fmt.Printf("[%s] created without errors!\n", appName)
}
