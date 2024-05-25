package api

import (
	"log"
	"time"

	"go-yandex-metrics/internal/config"
	"go-yandex-metrics/internal/logger"
	"go-yandex-metrics/internal/storage"

	"go.uber.org/zap"
)

type AgentCfg struct {
	Host           string
	PollInterval   uint64
	ReportInterval uint64
}

type Agent struct {
	logger *zap.Logger
	store  *storage.MemStorage
	cfg    config.AgentCfg
}

func NewAgent(cfg config.AgentCfg, store *storage.MemStorage) *Agent {
	lg := logger.InitLogger("agent.log")

	agt := &Agent{
		cfg:    cfg,
		store:  store,
		logger: lg,
	}
	return agt
}

func (a *Agent) Start() error {
	lg := a.logger

	tickerSave := time.NewTicker(time.Duration(a.cfg.PollInterval) * time.Second)
	tickerSend := time.NewTicker(time.Duration(a.cfg.ReportInterval) * time.Second)

	for {
		select {
		case <-tickerSave.C:
			err := a.SaveMetrics(lg)
			if err != nil {
				lg.Error("Failed to save metrics")
				log.Printf("failed to save metrics: %v", err)
				return nil
			}
		case <-tickerSend.C:
			err := a.SendMetricsJSONgzip(lg)
			if err != nil {
				lg.Error("Failed to send metrics")
				log.Printf("failed to send metrics: %v", err)
				return nil
			}
		}
	}
}
