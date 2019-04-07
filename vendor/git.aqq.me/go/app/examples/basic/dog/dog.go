package dog

import (
	"git.aqq.me/go/app/appconf"
	"git.aqq.me/go/app/applog"
	"git.aqq.me/go/app/event"
	"go.uber.org/zap"
)

// Dog is Dog
type Dog struct {
	Name string
	Age  uint
	Size uint
}

var dog *Dog
var logger *zap.Logger

func init() {
	appconf.Require("file:dog.toml")

	event.Init.AddHandler(
		func() error {
			dog = new(Dog)
			configRaw := appconf.GetConfig()
			appconf.Decode(configRaw["dog"], dog)

			logger = applog.GetLogger()
			logger = logger.With(zap.String("name", dog.Name))

			logger.Info("Dog initialized")

			return nil
		},
	)

	event.Reload.AddHandler(
		func() error {
			logger.Info("Dog reloaded")
			return nil
		},
	)

	event.Stop.AddHandler(
		func() error {
			logger.Info("Dog stopped")
			return nil
		},
	)
}

// GetDog returns Dog
func GetDog() *Dog {
	if dog == nil {
		panic("dog must be initialized first")
	}

	return dog
}
