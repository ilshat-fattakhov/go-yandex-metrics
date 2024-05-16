package api

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

const (
	GaugeType   string = "gauge"
	CounterType string = "counter"
)

type ServerCfg struct {
	Host string
}

type AgentCfg struct {
	Host           string
	PollInterval   uint64
	ReportInterval uint64
}

func NewServerConfig() (ServerCfg, error) {
	var cfg ServerCfg

	const defaultRunAddr = "localhost:8080"
	var flagRunAddr string

	flag.StringVar(&flagRunAddr, "a", defaultRunAddr, "address and port to run server")
	flag.Parse()

	cfg.Host = flagRunAddr
	envRunAddr, ok := os.LookupEnv("ADDRESS")
	if ok {
		cfg.Host = envRunAddr
	}
	return cfg, nil
}

func NewAgentConfig() (AgentCfg, error) {
	var cfg AgentCfg

	const defaultRunAddr = "localhost:8080"
	const defaultReportInterval uint64 = 10
	const defaultPollInterval uint64 = 2

	var flagRunAddr string
	var flagReportInterval uint64
	var flagPollInterval uint64

	var runAddr string
	var ReportInterval uint64
	var PollInterval uint64

	flag.StringVar(&flagRunAddr, "a", defaultRunAddr, "address and port to run server")
	flag.Uint64Var(&flagPollInterval, "p", defaultPollInterval, "data poll interval")
	flag.Uint64Var(&flagReportInterval, "r", defaultReportInterval, "data report interval")
	flag.Parse()

	runAddr = flagRunAddr
	envRunAddr, ok := os.LookupEnv("ADDRESS")
	if ok {
		runAddr = envRunAddr
	}
	cfg.Host = runAddr

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
			return cfg, fmt.Errorf("failed to parse %d as a poll interval value: %w", PollInterval, err)
		}
	}
	cfg.PollInterval = PollInterval

	return cfg, nil
}
