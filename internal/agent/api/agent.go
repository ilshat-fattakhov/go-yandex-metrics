package api

import (
	"fmt"
	"time"

	logger "go-yandex-metrics/cmd/server/middleware"
	"go-yandex-metrics/internal/config"
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
	lg := a.logger

	tickerSave := time.NewTicker(time.Duration(a.cfg.PollInterval) * time.Second)
	tickerSend := time.NewTicker(time.Duration(a.cfg.ReportInterval) * time.Second)

	for {
		select {
		case <-tickerSave.C:
			err := a.SaveMetrics(lg)
			if err != nil {
				lg.Info(fmt.Sprintf("failed to save metrics: %v", err))
				return nil
			}
		case <-tickerSend.C:
			err := a.SendMetrics(lg)
			if err != nil {
				lg.Info(fmt.Sprintf("failed to send metric: %v", err))
				return nil
			}
		}
	}
}
