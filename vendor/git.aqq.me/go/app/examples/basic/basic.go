package main

import (
	"git.aqq.me/go/app/appconf"
	"git.aqq.me/go/app/applog"
	"git.aqq.me/go/app/examples/basic/cat"
	"git.aqq.me/go/app/examples/basic/cow"
	"git.aqq.me/go/app/examples/basic/dog"
	"git.aqq.me/go/app/launcher"
	"github.com/iph0/conf/fileconf"
	"go.uber.org/zap"
)

func init() {
	fileLdr := fileconf.NewLoader("etc")
	appconf.RegisterLoader("file", fileLdr)
	appconf.Require("file:example.yml")
}

func main() {
	launcher.Run(
		func() error {
			logger := applog.GetLogger()
			cat := cat.GetCat()
			dog := dog.GetDog()
			cow := cow.GetCow()

			logger.Debug("Cat:", zap.Any("dump", cat))
			logger.Debug("Cow:", zap.Any("dump", cow))
			logger.Debug("Dog:", zap.Any("dump", dog))

			logger.Warn("Press Ctrl-C or send TERM signal to stop")

			return nil
		},
	)
}
