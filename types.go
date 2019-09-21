package main

import (
	"net/http"

	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

type tgConf struct {
	Path  string
	Proxy string
	Token string
	URL   string
}

type instanceConf struct {
	Listen     string
	RedisAddrs string
	Telegram   tgConf
}

type instanceObj struct {
	bot *tgbotapi.BotAPI
	cnf *instanceConf
	enc jsoniter.API
	log *zap.SugaredLogger
	rdb *redis.ClusterClient
	srv *http.Server
}
