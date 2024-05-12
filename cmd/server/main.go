package main

import (
	"context"
	"log"

	"go-yandex-metrics/cmd/server/api"
	"go-yandex-metrics/internal/storage"
)

func main() {
	ctx := context.Background()
	cfg, err := api.NewServerConfig()
	if err != nil {
		log.Fatal("failed to create config: %w", err)
	}
	store := storage.NewMemStorage()
	server := api.NewServer(cfg.HTTPServer, store)

	err = server.Start(ctx)
	if err != nil {
		log.Fatal("failed to start server %w", err)
	}
}
