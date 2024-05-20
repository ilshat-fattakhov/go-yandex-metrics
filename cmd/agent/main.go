package main

import (
	"log"

	"go-yandex-metrics/internal/agent/api"
	"go-yandex-metrics/internal/config"
	"go-yandex-metrics/internal/logger"
	"go-yandex-metrics/internal/storage"
)

func main() {
	logger := logger.InitLogger()

	cfg, err := config.NewAgentConfig()
	if err != nil {
		logger.Error("Failed to create agent config")
		log.Fatalf("failed to create config: %v", err)
	}

	store := storage.NewMemStorage()
	agent := api.NewAgent(cfg, store)

	err = agent.Start()
	if err != nil {
		logger.Error("Failed to start agent")
		log.Fatalf("failed to start agent %v", err)
	}
}
