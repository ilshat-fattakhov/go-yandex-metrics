package main

import (
	"fmt"
	"log"
	"runtime"
	"strconv"
	"time"

	"go-yandex-metrics/cmd/agent/handlers"
	"go-yandex-metrics/cmd/config"
)

func main() {
	if err := runAgent(); err != nil {
		log.Fatal(err)
	}
}

func runAgent() error {
	cfg, err := config.NewAgentConfig()
	if err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}
	m := new(runtime.MemStats)
	runtime.ReadMemStats(m)

	pollInterval, err := strconv.ParseUint(cfg.Agent.PollInterval, 10, 64)
	if err != nil {
		return fmt.Errorf("error parsing PollInterval: %w", err)
	}
	tickerSave := time.NewTicker(time.Duration(pollInterval) * time.Second)

	reportInterval, err := strconv.ParseUint(cfg.Agent.ReportInterval, 10, 64)
	if err != nil {
		return fmt.Errorf("error parsing ReportInterval: %w", err)
	}
	tickerSend := time.NewTicker(time.Duration(reportInterval) * time.Second)

	for {
		select {
		case <-tickerSave.C:
			handlers.SaveMetrics(m)
		case <-tickerSend.C:
			handlers.SendMetrics(cfg.Agent.Host)
		}
	}
}
