package api

import (
	"context"
	"fmt"
	"time"

	"go-yandex-metrics/internal/storage"
)

func NewAgent(cfg HTTPAgent, store *storage.MemStorage) *Agent {
	agt := &Agent{
		cfg:   cfg,
		store: *store,
	}

	return agt
}
func (a *Agent) Start(ctx context.Context) error {
	tickerSave := time.NewTicker(time.Duration(a.cfg.PollInterval) * time.Second)
	tickerSend := time.NewTicker(time.Duration(a.cfg.ReportInterval) * time.Second)

	for {
		select {
		case <-tickerSave.C:
			err := a.SaveMetrics()
			if err != nil {
				return fmt.Errorf("failed to save metrics: %w", err)
			}
		case <-tickerSend.C:
			err := a.SendMetrics()
			if err != nil {
				return fmt.Errorf("failed to send metrics: %w", err)
			}
		}
	}
}
