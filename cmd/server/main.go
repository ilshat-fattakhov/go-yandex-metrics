package main

import (
	"log"

	"go-yandex-metrics/internal/config"
	"go-yandex-metrics/internal/server/api"
	"go-yandex-metrics/internal/storage"
)

func main() {
	cfg, err := config.NewServerConfig()
	if err != nil {
		log.Fatalf("failed to create config: %v", err)
	}

	store := storage.NewMemStorage()
	server, err := api.NewServer(cfg, store)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	err = server.Start()
	if err != nil {
		log.Fatalf("failed to start server %v", err)
	}
}
