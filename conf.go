package main

import (
	"github.com/iph0/conf"
	"github.com/iph0/conf/envconf"
	"github.com/iph0/conf/fileconf"
)

func newConf() (*instanceConf, error) {
	fileLdr := fileconf.NewLoader("etc", "/etc")
	envLdr := envconf.NewLoader()

	configProc := conf.NewProcessor(
		conf.ProcessorConfig{
			Loaders: map[string]conf.Loader{
				"file": fileLdr,
				"env":  envLdr,
			},
		},
	)

	configRaw, err := configProc.Load(
		"file:pickme.yml",
		"env:^PICKME_",
	)

	if err != nil {
		return nil, err
	}

	var cnf instanceConf
	if err := conf.Decode(configRaw["pickme"], &cnf); err != nil {
		return nil, err
	}

	return &cnf, nil
}
