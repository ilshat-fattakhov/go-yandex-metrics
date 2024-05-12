package api

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"go-yandex-metrics/internal/storage"
)

type HTTPAgent struct {
	Host           string
	PollInterval   uint64
	ReportInterval uint64
}

type Configuration struct {
	HTTPAgent
}

type Agent struct {
	store storage.MemStorage
	cfg   HTTPAgent
}

func NewAgentConfig() (Configuration, error) {
	var cfg Configuration

	var defaultRunAddr = "localhost:8080"
	var defaultReportInterval uint64 = 10
	var defaultPollInterval uint64 = 2

	var flagRunAddr string
	var flagReportInterval uint64
	var flagPollInterval uint64

	var RunAddr string
	var ReportInterval uint64
	var PollInterval uint64

	flag.StringVar(&flagRunAddr, "a", defaultRunAddr, "address and port to run server")
	flag.Uint64Var(&flagPollInterval, "p", defaultPollInterval, "data poll interval")
	flag.Uint64Var(&flagReportInterval, "r", defaultReportInterval, "data report interval")
	flag.Parse()

	RunAddr = flagRunAddr
	envRunAddr, ok := os.LookupEnv("ADDRESS")
	if ok {
		RunAddr = envRunAddr
	}
	cfg.Host = RunAddr

	ReportInterval = flagReportInterval
	envReportInterval, ok := os.LookupEnv("REPORT_INTERVAL")
	if ok {
		ReportInterval, err := strconv.ParseUint(envReportInterval, 10, 64)
		if err != nil {
			return cfg, fmt.Errorf("failed to parse %d as a report interval value: %w", ReportInterval, err)
		}
	}
	cfg.ReportInterval = ReportInterval

	PollInterval = flagPollInterval
	envPollInterval, ok := os.LookupEnv("POLL_INTERVAL")
	if ok {
		PollInterval, err := strconv.ParseUint(envPollInterval, 10, 64)
		if err != nil {
			return cfg, fmt.Errorf("failed to parse %d as a report interval value: %w", PollInterval, err)
		}
	}
	cfg.PollInterval = PollInterval

	return cfg, nil
}
