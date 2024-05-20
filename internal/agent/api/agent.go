package api

import (
	"log"
	"time"

	"go-yandex-metrics/internal/config"
	"go-yandex-metrics/internal/logger"
	"go-yandex-metrics/internal/storage"
)

type AgentCfg struct {
	Host           string
	PollInterval   uint64
	ReportInterval uint64
}

type Agent struct {
	store *storage.MemStorage
	cfg   config.AgentCfg
}

func NewAgent(cfg config.AgentCfg, store *storage.MemStorage) *Agent {
	agt := &Agent{
		cfg:   cfg,
		store: store,
	}
	return agt
}

func (a *Agent) Start() error {
	logger := logger.InitLogger()

	tickerSave := time.NewTicker(time.Duration(a.cfg.PollInterval) * time.Second)
	tickerSend := time.NewTicker(time.Duration(a.cfg.ReportInterval) * time.Second)

	for {
		select {
		case <-tickerSave.C:
			err := a.SaveMetrics(logger)
			if err != nil {
				logger.Error("Failed to save metrics")
				log.Printf("failed to save metrics: %v", err)
				return nil
			}
		case <-tickerSend.C:
			err := a.SendMetricsJSON(logger)
			if err != nil {
				logger.Error("Failed to send metrics")
				log.Printf("failed to send metrics: %v", err)
				return nil
			}
		}
	}
}
