package app_test

import (
	"errors"
	"testing"

	"git.aqq.me/go/app"
	"git.aqq.me/go/app/event"
)

func TestEvents(t *testing.T) {
	t.Run("init_event",
		func(t *testing.T) {
			var called bool

			event.Init.AddHandler(
				func() error {
					called = true
					return nil
				},
			)

			err := app.Init()

			if err != nil {
				t.Error(err)
				return
			}

			if !called {
				t.Error("handler not called")
			}
		},
	)

	t.Run("reload_event",
		func(t *testing.T) {
			var called bool

			event.Reload.AddHandler(
				func() error {
					called = true
					return nil
				},
			)

			err := app.Reload()

			if err != nil {
				t.Error(err)
				return
			}

			if !called {
				t.Error("handler not called")
			}
		},
	)

	t.Run("stop_event",
		func(t *testing.T) {
			var called bool

			event.Stop.AddHandler(
				func() error {
					called = true
					return nil
				},
			)

			err := app.Stop()

			if err != nil {
				t.Error(err)
				return
			}

			if !called {
				t.Error("handler not called")
			}
		},
	)
}

func TestError(t *testing.T) {
	t.Run("init_error",
		func(t *testing.T) {
			event.Init.AddHandler(
				func() error {
					return errors.New("some error")
				},
			)

			err := app.Init()

			if err == nil {
				t.Error("no error happened")
			}
		},
	)

	t.Run("reload_error",
		func(t *testing.T) {
			event.Reload.AddHandler(
				func() error {
					return errors.New("some error")
				},
			)

			err := app.Reload()

			if err == nil {
				t.Error("no error happened")
			}
		},
	)

	t.Run("stop_error",
		func(t *testing.T) {
			event.Stop.AddHandler(
				func() error {
					return errors.New("some error")
				},
			)

			err := app.Stop()

			if err == nil {
				t.Error("no error happened")
			}
		},
	)
}
