package cmdparser_test

import (
	"github.com/gojekfarm/ziggurat-go/pkg/cmdparser"
	"github.com/gojekfarm/ziggurat-go/pkg/zbasic"
	"github.com/gojekfarm/ziggurat-go/pkg/zlogger"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	zlogger.ConfigureLogger("disabled")
	os.Exit(m.Run())
}

func TestParseCommandLineArgumentsWithDefaultValues(t *testing.T) {
	expected := zbasic.CommandLineOptions{ConfigFilePath: "./config/config.yaml"}
	cmdOptions := cmdparser.ParseCommandLineArguments()
	if expected != cmdOptions {
		t.Errorf("EXPECTED %+v GOT %+v", expected, cmdOptions)
	}

}

func TestParseCommandLineArguments(t *testing.T) {
	os.Args = append(os.Args, "--config=overriddenPath")
	cmdOptions := cmdparser.ParseCommandLineArguments()
	newOptions := zbasic.CommandLineOptions{ConfigFilePath: "overriddenPath"}
	if newOptions != cmdOptions {
		t.Errorf("FAILED got %+v EXPECTED %+v", cmdOptions, newOptions)
	}
}
