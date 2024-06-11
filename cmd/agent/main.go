package main

import (
	"fmt"
	"log"

	"go-yandex-metrics/internal/agent/api"
	"go-yandex-metrics/internal/config"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.NewAgentConfig()
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	store, err := api.NewAgentMemStorage(cfg)
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	agent, err := api.NewAgent(cfg, store)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	err = agent.Start()
	if err != nil {
		return fmt.Errorf("failed to start agent: %w", err)
	}

	return nil
}
