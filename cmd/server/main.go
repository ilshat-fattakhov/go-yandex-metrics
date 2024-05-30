package main

import (
	"log"

	logger "go-yandex-metrics/cmd/server/middleware"
	"go-yandex-metrics/internal/config"
	"go-yandex-metrics/internal/server/api"
	"go-yandex-metrics/internal/storage"
)

func main() {
	lg := logger.InitLogger()

	lg.Info("creating configuration")
	cfg, storageCfg, err := config.NewServerConfig()
	if err != nil {
		lg.Info("got error creating configuration")
		log.Fatalf("failed to create config: %v", err)
	}

	store := storage.NewFileStorage()
	lg.Info("creating server")
	server, err := api.NewServer(cfg, storageCfg, store)
	if err != nil {
		lg.Info("got error creating server")
		log.Fatalf("failed to create server: %v", err)
	}

	lg.Info("starting server")
	err = server.Start()
	if err != nil {
		lg.Info("got error starting server")
		log.Fatalf("failed to start server %v", err)
	}
}
