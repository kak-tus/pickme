package launcher_test

import (
	"os"
	"syscall"
	"testing"
	"time"

	"git.aqq.me/go/app/appconf"
	"git.aqq.me/go/app/launcher"
)

const sleepInterval = 100 * time.Millisecond

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

func TestRun(t *testing.T) {
	var called bool

	go func() {
		launcher.Run(
			func() error {
				called = true
				return nil
			},
		)
	}()

	time.Sleep(sleepInterval)

	if !called {
		t.Error("start failed")
		return
	}

	proc, err := os.FindProcess(os.Getpid())

	if err != nil {
		t.Error(err)
	}

	proc.Signal(syscall.SIGHUP)
	time.Sleep(sleepInterval)

	proc.Signal(syscall.SIGINT)
	time.Sleep(sleepInterval)
}
