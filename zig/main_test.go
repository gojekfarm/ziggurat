package zig

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	turnOffLogger()
	os.Exit(m.Run())
}

func turnOffLogger() {
	configureLogger("disabled")
	logInfo = func(msg string, args map[string]interface{}) {

	}
	logError = func(err error, msg string, args map[string]interface{}) {

	}

	logFatal = func(err error, msg string, args map[string]interface{}) {

	}

	logWarn = func(msg string, args map[string]interface{}) {

	}
}
