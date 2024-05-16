package api

import (
	"log"
	"time"

	"go-yandex-metrics/internal/storage"
)

type Agent struct {
	store storage.MemStorage
	cfg   AgentCfg
}

func NewAgent(cfg AgentCfg, store *storage.MemStorage) *Agent {
	agt := &Agent{
		cfg:   cfg,
		store: *store,
	}

	return agt
}

func (a *Agent) Start() error {
	tickerSave := time.NewTicker(time.Duration(a.cfg.PollInterval) * time.Second)
	tickerSend := time.NewTicker(time.Duration(a.cfg.ReportInterval) * time.Second)

	for {
		select {
		case <-tickerSave.C:
			err := a.SaveMetrics()
			if err != nil {
				log.Println("failed to save metrics: %w", err)
				return nil
			}
		case <-tickerSend.C:
			err := a.SendMetrics()
			if err != nil {
				log.Println("failed to send metrics: %w", err)
				return nil
			}
		}
	}
}
