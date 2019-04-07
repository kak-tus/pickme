package cat

import (
	"git.aqq.me/go/app/appconf"
	"git.aqq.me/go/app/applog"
	"git.aqq.me/go/app/event"
	"go.uber.org/zap"
)

// Cat is cat
type Cat struct {
	Name string
	Age  uint
	Size uint
}

var cat *Cat
var logger *zap.Logger

func init() {
	appconf.Require("file:cat.json")

	event.Init.AddHandler(
		func() error {
			cat = new(Cat)
			configRaw := appconf.GetConfig()
			appconf.Decode(configRaw["cat"], cat)

			logger = applog.GetLogger()
			logger = logger.With(zap.String("name", cat.Name))

			logger.Info("Cat initialized")

			return nil
		},
	)

	event.Reload.AddHandler(
		func() error {
			logger.Info("Cat reloaded")
			return nil
		},
	)

	event.Stop.AddHandler(
		func() error {
			logger.Info("Cat stopped")
			return nil
		},
	)
}

// GetCat returns Cat
func GetCat() *Cat {
	if cat == nil {
		panic("cat must be initialized first")
	}

	return cat
}
