package main

import (
	"fmt"
	"log"

	"go.uber.org/zap"

	"go-yandex-metrics/internal/config"
	"go-yandex-metrics/internal/server/api"
	logger "go-yandex-metrics/internal/server/middleware"
	"go-yandex-metrics/internal/storage"
)

func main() {
	lg := logger.InitLogger()
	defer func() {
		if err := lg.Sync(); err != nil {
			lg.Info(fmt.Sprintf("failed to sync logger: %v", err))
		}
	}()

	cfg, err := config.NewServerConfig()
	lg.Info("server configuration settings" + fmt.Sprint(cfg))
	if err != nil {
		lg.Info("failed to create config", zap.Error(err))
		log.Panicf("failed to create config: %v", err)
	}

	store, err := storage.NewStore(&cfg)
	if err != nil {
		lg.Info("failed to create storage", zap.Error(err))
		log.Panicf("failed to create storage %v", err)
	}
	server, err := api.NewServer(cfg, store)
	if err != nil {
		lg.Info("failed to create server", zap.Error(err))
		log.Panicf("failed to create server: %v", err)
	}

	err = server.Start(cfg)
	if err != nil {
		lg.Info("failed to start server", zap.Error(err))
		log.Panicf("failed to start server %v", err)
	}
}
