package api

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"go.uber.org/zap"

	"go-yandex-metrics/internal/config"
	logger "go-yandex-metrics/internal/server/middleware"
)

type AgentCfg struct {
	Host           string
	HashKey        string
	PollInterval   uint64
	ReportInterval uint64
}

type Agent struct {
	logger *zap.Logger
	store  *MemStorage
	client *http.Client
	cfg    config.AgentCfg
}

type MemStorage struct {
	Gauge   map[string]float64 `json:"gauge"`
	Counter map[string]int64   `json:"counter"`
	memLock *sync.Mutex
}

func NewAgent(cfg config.AgentCfg, store *MemStorage) (*Agent, error) {
	lg, err := logger.InitLogger()
	if err != nil {
		return nil, fmt.Errorf("failed to init logger: %w", err)
	}

	agt := &Agent{
		logger: lg,
		store:  store,
		client: &http.Client{
			Timeout:   time.Duration(1) * time.Second,
			Transport: &http.Transport{},
		},
		cfg: cfg,
	}
	return agt, nil
}

func (a *Agent) Start() error {
	tickerSave := time.NewTicker(time.Duration(a.cfg.PollInterval) * time.Second)
	tickerSend := time.NewTicker(time.Duration(a.cfg.ReportInterval) * time.Second)
	batchSend := time.NewTicker(time.Duration(a.cfg.ReportInterval) * time.Second)

	for {
		select {
		case <-tickerSave.C:
			a.saveMetrics()
		case <-tickerSend.C:
			err := a.sendMetrics()
			if err != nil {
				a.logger.Error("failed to send metrics, trying again: %w", zap.Error(err))
				if a.cfg.ReportInterval > 1 {
					for i := 1; i <= 5; i += 2 {
						if i > int(a.cfg.ReportInterval) {
							break
						}
						time.Sleep(time.Duration(i) * time.Second)
						err := a.sendMetrics()
						if err != nil {
							a.logger.Error("failed to send metrics after "+strconv.Itoa(i)+" second(s)", zap.Error(err))
						}
					}
				}
			}
		case <-batchSend.C:
			err := a.sendMetricsBatch()
			i := 1
			for err != nil {
				time.Sleep(time.Duration(i) * time.Second)
				err = a.sendMetricsBatch()
				i += 2
				if i >= 5 || i > int(a.cfg.ReportInterval) {
					break
				}
			}
			if err != nil {
				a.logger.Error("failed to send a batch of metrics after "+strconv.Itoa(i)+" second(s)", zap.Error(err))
			}
		}
	}
}

func NewAgentMemStorage(cfg config.AgentCfg) (*MemStorage, error) {
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
		memLock: &sync.Mutex{},
	}, nil
}
