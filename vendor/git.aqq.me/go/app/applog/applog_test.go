package applog_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"git.aqq.me/go/app"
	"git.aqq.me/go/app/appconf"
	"git.aqq.me/go/app/applog"
)

func init() {
	appconf.Require(
		map[string]interface{}{
			"log": map[string]interface{}{
				"tag":    "test",
				"level":  "debug",
				"output": "stdsep",
				"format": "console",
				"formatConfig": map[string]interface{}{
					"colors": true,
				},
			},
		},
	)
}

func TestStdLogger(t *testing.T) {
	err := app.Init()

	if err != nil {
		t.Error("Init failed:", err)
		return
	}

	logger := applog.GetLogger()

	logger.Debug("Message 1")
	logger.Info("Message 2")
	logger.Warn("Message 3")
	logger.Error("Message 4")

	err = app.Stop()

	if err != nil {
		t.Error("Stop failed:", err)
	}
}

func TestFileLogger(t *testing.T) {
	logFile := "test.log"

	appconf.Require(
		map[string]interface{}{
			"log": map[string]interface{}{
				"output": "file",
				"outputConfig": map[string]interface{}{
					"filePath": logFile,
				},
				"format": "json",
			},
		},
	)

	err := app.Init()

	if err != nil {
		t.Error("Init failed:", err)
		return
	}

	logger := applog.GetLogger()

	logger.Debug("Message 1")
	logger.Info("Message 2")
	logger.Warn("Message 3")
	logger.Error("Message 4")

	err = app.Stop()

	if err != nil {
		t.Error("Stop failed:", err)
	}

	err = os.Remove(logFile)

	if err != nil {
		t.Error("Can't delete log-file", err)
	}
}

func TestPanic(t *testing.T) {
	t.Run("not_initialized",
		func(t *testing.T) {
			defer func() {
				err := recover()
				errStr := fmt.Sprintf("%v", err)

				if err == nil {
					t.Error("no error happened")
				} else if strings.Index(errStr, "must be initialized first") == -1 {
					t.Error("other error happened")
				}
			}()

			applog.GetLogger()
		},
	)
}
