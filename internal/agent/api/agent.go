package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/hashicorp/go-retryablehttp"
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

	retClient := retryablehttp.NewClient()
	retClient.Backoff = Backoff

	retClient.RetryMax = 3
	retClient.RetryWaitMin = 1 * time.Second
	retClient.RetryWaitMax = 5 * time.Second

	stdClient := retClient.StandardClient()

	agt := &Agent{
		logger: lg,
		store:  store,
		client: stdClient,
		cfg:    cfg,
	}
	return agt, nil
}

func (a *Agent) Start() error {
	tickerSave := time.NewTicker(time.Duration(a.cfg.PollInterval) * time.Second)
	tickerSend := time.NewTicker(time.Duration(a.cfg.ReportInterval) * time.Second)

	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	wg := &sync.WaitGroup{}
	maxJobs := int(a.cfg.RateLimit)
	jobs := make(chan []MetricsToSend, maxJobs)
	for range maxJobs {
		go func() {
			defer wg.Done()
			for m := range jobs {
				err := a.sendMetricsBatch(m)
				if err != nil {
					a.logger.Info("failed to send metrics", zap.Error(err))
				}
			}
		}()
		wg.Add(1)
	}

	for {
		select {
		case <-tickerSend.C:
			metrics := a.getMetricsToSend()
			jobs <- metrics
		case <-tickerSave.C:
			go a.saveMetrics()
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return nil
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

func Backoff(minValue, maxValue time.Duration, attemptNum int, resp *http.Response) time.Duration {
	switch attemptNum {
	case 0:
		return 1 * time.Second
	case 1:
		return 3 * time.Second
	case 2:
		return 5 * time.Second
	default:
		return 3 * time.Second
	}
}
