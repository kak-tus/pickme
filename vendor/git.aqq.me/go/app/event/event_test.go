package event_test

import (
	"testing"

	"git.aqq.me/go/app/event"
)

func TestAddHandler(t *testing.T) {
	t.Run("add_init_handler",
		func(t *testing.T) {
			event.Init.AddHandler(
				func() error {
					return nil
				},
			)

			if len(event.Init.Handlers) == 0 {
				t.Error("failed to add handler")
			}
		},
	)

	t.Run("add_reload_handler",
		func(t *testing.T) {
			event.Reload.AddHandler(
				func() error {
					return nil
				},
			)

			if len(event.Reload.Handlers) == 0 {
				t.Error("failed to add handler")
			}
		},
	)

	t.Run("add_stop_handler",
		func(t *testing.T) {
			event.Stop.AddHandler(
				func() error {
					return nil
				},
			)

			if len(event.Stop.Handlers) == 0 {
				t.Error("failed to add handler")
			}
		},
	)
}
