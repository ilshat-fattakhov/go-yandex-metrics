package main

import (
	"log"

	"go-yandex-metrics/internal/config"
	"go-yandex-metrics/internal/server/api"
	"go-yandex-metrics/internal/storage"
)

func main() {
	cfg, storageCfg, err := config.NewServerConfig()
	if err != nil {
		log.Fatalf("failed to create config: %v", err)
	}

	store := storage.NewFileStorage()
	server, err := api.NewServer(cfg, storageCfg, store)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	err = server.Start()
	if err != nil {
		log.Fatalf("failed to start server %v", err)
	}
}
