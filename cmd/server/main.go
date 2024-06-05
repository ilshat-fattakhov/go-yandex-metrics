package main

import (
	"fmt"
	"log"

	"go.uber.org/zap"

	"go-yandex-metrics/internal/config"
	"go-yandex-metrics/internal/server/api"
	logger "go-yandex-metrics/internal/server/middleware"
)

func main() {
	lg := logger.InitLogger()
	defer func() {
		if err := lg.Sync(); err != nil {
			lg.Info(fmt.Sprintf("failed to sync logger: %v", err))
			return
		}
	}()

	cfg, err := config.NewServerConfig()
	lg.Info("server configuration settings" + fmt.Sprint(cfg))

	if err != nil {
		lg.Info("got error creating configuration", zap.Error(err))
		log.Panicf("failed to create config: %v", err)
	}

	server, err := api.NewServer(*cfg)
	if err != nil {
		lg.Info("got error creating server", zap.Error(err))
		log.Panicf("failed to create server: %v", err)
	}

	err = server.Start(*cfg)
	if err != nil {
		lg.Info("got error starting server", zap.Error(err))
		log.Panicf("failed to start server %v", err)
	}
}
