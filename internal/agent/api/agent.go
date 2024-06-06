package api

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"go-yandex-metrics/internal/config"
	logger "go-yandex-metrics/internal/server/middleware"
	"go-yandex-metrics/internal/storage"
)

type AgentCfg struct {
	Host           string
	PollInterval   uint64
	ReportInterval uint64
}

type Agent struct {
	logger *zap.Logger
	store  *storage.FileStorage
	cfg    config.AgentCfg
}

func NewAgent(cfg config.AgentCfg, store *storage.FileStorage) *Agent {
	lg := logger.InitLogger()

	agt := &Agent{
		cfg:    cfg,
		store:  store,
		logger: lg,
	}
	return agt
}

func (a *Agent) Start() error {
	tickerSave := time.NewTicker(time.Duration(a.cfg.PollInterval) * time.Second)
	tickerSend := time.NewTicker(time.Duration(a.cfg.ReportInterval) * time.Second)

	for {
		select {
		case <-tickerSave.C:
			a.saveMetrics()
		case <-tickerSend.C:
			err := a.sendMetrics()
			if err != nil {
				a.logger.Info(fmt.Sprintf("failed to send metrics: %v", err))
				return fmt.Errorf("failed to send metrics: %w", err)
			}
		}
	}
}
