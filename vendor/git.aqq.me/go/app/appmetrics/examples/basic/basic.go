package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"git.aqq.me/go/app"
	"git.aqq.me/go/app/appconf"
	"git.aqq.me/go/app/appmetrics"
	"github.com/iph0/conf/fileconf"
)

func init() {
	fileLdr := fileconf.NewLoader("etc")
	appconf.RegisterLoader("file", fileLdr)
	appconf.Require("file:example.yml")
}

func main() {
	err := app.Init()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	client := appmetrics.GetClient()

	stop := make(chan struct{}, 1)
	listenSignal(stop)

LOOP:
	for {
		value := 10 + rand.Int63n(20)

		client.Timing("requestTime", value,
			"host", "test.com",
			"dc", "dtln",
		)

		fmt.Println("Value sent:", value)

		interval := rand.Int63n(300)
		t := time.NewTimer(time.Millisecond * time.Duration(interval))

		select {
		case <-stop:
			t.Stop()
			break LOOP
		case <-t.C:
		}
	}

	err = app.Stop()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func listenSignal(stop chan struct{}) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	var sigSent bool

	go func() {
		for s := range signals {
			fmt.Println("Got signal:", s)

			if !sigSent {
				close(stop)
				sigSent = true
			}
		}
	}()
}
