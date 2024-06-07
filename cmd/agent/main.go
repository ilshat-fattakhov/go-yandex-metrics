package main

import (
	"fmt"
	"log"

	"go-yandex-metrics/internal/agent/api"
	"go-yandex-metrics/internal/config"
	logger "go-yandex-metrics/internal/server/middleware"
)

func main() {
	lg := logger.InitLogger()
	defer func() {
		if err := lg.Sync(); err != nil {
			lg.Info(fmt.Sprintf("failed to sync logger: %v", err))
		}
	}()

	cfg, err := config.NewAgentConfig()
	lg.Info("agent configuration settings" + fmt.Sprint(cfg))
	if err != nil {
		log.Panicf("failed to create config: %v", err)
	}

	store, err := api.NewAgentMemStorage(cfg)
	if err != nil {
		log.Panicf("got error creating storage %v", err)
	}
	agent := api.NewAgent(cfg, store)

	err = agent.Start()
	if err != nil {
		log.Panicf("failed to start agent %v", err)
	}
}
