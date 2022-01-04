package main

import (
	"net/http"

	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	jsoniter "github.com/json-iterator/go"
	regen "github.com/zach-klippenstein/goregen"
	"go.uber.org/zap"
)

type instanceObj struct {
	bot *tgbotapi.BotAPI
	cnf *instanceConf
	enc jsoniter.API
	gen regen.Generator
	log *zap.SugaredLogger
	rdb *redis.ClusterClient
	srv *http.Server
}
