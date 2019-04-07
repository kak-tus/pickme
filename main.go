package main

import (
	"git.aqq.me/go/app/appconf"
	"git.aqq.me/go/app/launcher"
	"github.com/iph0/conf/envconf"
	"github.com/iph0/conf/fileconf"
)

func init() {
	fileLdr := fileconf.NewLoader("etc", "/etc")
	envLdr := envconf.NewLoader()

	appconf.RegisterLoader("file", fileLdr)
	appconf.RegisterLoader("env", envLdr)

	appconf.Require("file:pickme.yml")
	appconf.Require("env:^PICKME_")
}

func main() {
	launcher.Run(func() error {
		err := inst.Start()
		if err != nil {
			return err
		}

		return nil
	})
}
