package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	logOld := logger.Sugar()

	log := zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()

	tg, err := newTg(log)
	if err != nil {
		logOld.Panic(err)
	}

	go func() {
		if err := tg.start(); err != nil {
			logOld.Panic(err)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	logOld.Info("Stop")

	if err := tg.stop(); err != nil {
		logOld.Panic(err)
	}

	_ = logOld.Sync()
}
