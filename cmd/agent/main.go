package main

import (
	"context"
	"log"

	"go-yandex-metrics/cmd/agent/api"
	"go-yandex-metrics/internal/storage"
)

func main() {
	ctx := context.Background()
	cfg, err := api.NewAgentConfig()
	if err != nil {
		log.Fatal("failed to create config: %w", err)
	}
	store := storage.NewMemStorage()
	agent := api.NewAgent(cfg.HTTPAgent, store)

	err = agent.Start(ctx)
	if err != nil {
		log.Fatal("failed to start agent %w", err)
	}
}
