package main

import "github.com/kelseyhightower/envconfig"

type tgConf struct {
	Path  string
	Proxy string
	Token string
	URL   string
}

type instanceConf struct {
	Listen     string `default:"0.0.0.0:8080"`
	RedisAddrs string
	Telegram   tgConf
}

func newConf() (*instanceConf, error) {
	var cnf instanceConf

	err := envconfig.Process("PICKME", &cnf)
	if err != nil {
		return nil, err
	}

	return &cnf, nil
}
