package main

import (
	"fmt"
	logger "go-yandex-metrics/cmd/server/middleware"
	"go-yandex-metrics/internal/config"
	"go-yandex-metrics/internal/server/api"
	"go-yandex-metrics/internal/storage"

	"go.uber.org/zap"
)

func main() {
	lg := logger.InitLogger()

	lg.Info("creating configuration")
	cfg, storageCfg, err := config.NewServerConfig()
	if err != nil {
		lg.Info("got error creating configuration", zap.Error(err))
		// log.Fatalf("failed to create config: %v", err)
	}
	fmt.Println(cfg, storageCfg)
	store := storage.NewFileStorage()
	lg.Info("creating server")
	server, err := api.NewServer(cfg, storageCfg, store)
	if err != nil {
		lg.Info("got error creating server", zap.Error(err))
		// log.Fatalf("failed to create server: %v", err)
	}

	lg.Info("starting server")
	err = server.Start()
	if err != nil {
		lg.Info("got error starting server", zap.Error(err))
		// log.Fatalf("failed to start server %v", err)
	}
}
