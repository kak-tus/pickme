package main

import (
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	log := logger.Sugar()

	tg, err := newTg()
	if err != nil {
		log.Panic(err)
	}

	go func() {
		if err := tg.start(); err != nil {
			log.Panic(err)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	log.Info("Stop")

	if err := tg.stop(); err != nil {
		log.Panic(err)
	}

	_ = log.Sync()
}
