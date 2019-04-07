package appmetrics

import (
	"fmt"

	"git.aqq.me/go/app/event"
)

const errPref = "appmetrics"

var client *Client

func init() {
	initHandler := func() error {
		var err error
		client, err = NewClient()

		if err != nil {
			return err
		}

		return nil
	}

	event.Init.AddHandler(initHandler)
	event.Reload.AddHandler(initHandler)

	event.Stop.AddHandler(
		func() error {
			if client != nil {
				client.Close()
				client = nil
			}

			return nil
		},
	)
}

// GetClient returns StatsD client instance.
func GetClient() *Client {
	if client == nil {
		panic(fmt.Errorf("%s must be initialized first", errPref))
	}

	return client
}
