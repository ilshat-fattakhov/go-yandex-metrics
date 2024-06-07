package api

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"go-yandex-metrics/internal/config"
	logger "go-yandex-metrics/internal/server/middleware"
)

type AgentCfg struct {
	Host           string
	PollInterval   uint64
	ReportInterval uint64
}

type Agent struct {
	logger *zap.Logger
	store  *MemStorage
	cfg    config.AgentCfg
}

type MemStorage struct {
	Gauge   map[string]float64 `json:"gauge"`
	Counter map[string]int64   `json:"counter"`
	memLock sync.Mutex
}

func NewAgent(cfg config.AgentCfg, store *MemStorage) *Agent {
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

func NewAgentMemStorage(cfg config.AgentCfg) (*MemStorage, error) {
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}, nil
}
