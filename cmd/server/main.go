package main

import (
	"fmt"
	"log"

	"go-yandex-metrics/internal/config"
	"go-yandex-metrics/internal/server/api"
	"go-yandex-metrics/internal/storage"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.NewServerConfig()
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	store, err := storage.NewStore(&cfg)
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	server, err := api.NewServer(&cfg, store)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	err = server.Start(&cfg)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}
