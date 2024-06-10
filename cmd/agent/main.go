package main

import (
	"fmt"
	"log"

	"go-yandex-metrics/internal/agent/api"
	"go-yandex-metrics/internal/config"
	logger "go-yandex-metrics/internal/server/middleware"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	lg := logger.InitLogger()
	defer func() {
		if err := lg.Sync(); err != nil {
			lg.Debug(fmt.Sprintf("failed to sync logger: %v", err))
		}
	}()

	cfg, err := config.NewAgentConfig()
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	store, err := api.NewAgentMemStorage(cfg)
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	lg.Info(fmt.Sprintf("agent config data: %v", cfg))
	lg.Info(fmt.Sprintf("agent storage data: %v", store))

	agent := api.NewAgent(cfg, store)

	err = agent.Start()
	if err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}
	return nil
}
