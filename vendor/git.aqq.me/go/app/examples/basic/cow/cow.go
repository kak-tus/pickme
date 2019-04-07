package cow

import (
	"git.aqq.me/go/app/appconf"
	"git.aqq.me/go/app/applog"
	"git.aqq.me/go/app/event"
	"go.uber.org/zap"
)

// Cow is Cow
type Cow struct {
	Name string
	Age  uint
	Size uint
}

var cow *Cow
var logger *zap.Logger

func init() {
	appconf.Require("file:cow.yml")

	event.Init.AddHandler(
		func() error {
			cow = new(Cow)
			configRaw := appconf.GetConfig()
			appconf.Decode(configRaw["cow"], cow)

			logger = applog.GetLogger()
			logger = logger.With(zap.String("name", cow.Name))

			logger.Info("Cow initialized")

			return nil
		},
	)

	event.Reload.AddHandler(
		func() error {
			logger.Info("Cow reloaded")
			return nil
		},
	)

	event.Stop.AddHandler(
		func() error {
			logger.Info("Cow stopped")
			return nil
		},
	)
}

// GetCow returns Cow
func GetCow() *Cow {
	if cow == nil {
		panic("cow must be initialized first")
	}

	return cow
}
