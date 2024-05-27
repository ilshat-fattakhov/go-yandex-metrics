package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

const (
	GaugeType   string = "gauge"
	CounterType string = "counter"
)

type ServerCfg struct {
	Host            string `json:"host"`
	FileStoragePath string `json:"file_storage_path"`
	StoreInterval   uint64 `json:"store_interval"`
	Restore         bool   `json:"restore"`
}

type AgentCfg struct {
	Host           string `json:"host"`
	PollInterval   uint64 `json:"poll_interval"`
	ReportInterval uint64 `json:"report_interval"`
}

func NewServerConfig() (ServerCfg, error) {
	var cfg ServerCfg

	const defaultRunAddr = "localhost:8080"
	const defaultStoreInterval uint64 = 300               // значение 0 делает запись синхронной
	const defaultFileStoragePath = "/tmp/metrics-db.json" // пустое значение отключает функцию записи на диск
	const defaultRestore = true

	var flagRunAddr string
	var flagStoreInterval uint64
	var flagFileStoragePath string
	var flagRestore bool

	flag.StringVar(&flagRunAddr, "a", defaultRunAddr, "address and port to run server")

	flag.BoolVar(&flagRestore, "r", defaultRestore, "restore data from file at server start")
	flag.Uint64Var(&flagStoreInterval, "i", defaultStoreInterval, "data storing interval")
	flag.StringVar(&flagFileStoragePath, "f", defaultFileStoragePath, "file storage path")
	flag.Parse()

	cfg.Host = flagRunAddr
	envRunAddr, ok := os.LookupEnv("ADDRESS")
	if ok {
		cfg.Host = envRunAddr
	}

	cfg.StoreInterval = flagStoreInterval
	envStoreInterval, ok := os.LookupEnv("STORE_INTERVAL")
	if ok {
		tmpStoreInterval, err := strconv.ParseUint(envStoreInterval, 10, 64)
		if err != nil {
			return cfg, fmt.Errorf("failed to parse %d as a report interval value: %w", tmpStoreInterval, err)
		}
		cfg.StoreInterval = tmpStoreInterval
	}

	cfg.FileStoragePath = flagFileStoragePath
	envFileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH")
	if ok {
		cfg.FileStoragePath = envFileStoragePath
	}

	cfg.Restore = flagRestore
	envRestore, ok := os.LookupEnv("RESTORE")
	if ok {
		boolValue, err := strconv.ParseBool(envRestore)
		if err != nil {
			log.Fatal(err)
		}
		cfg.Restore = boolValue
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

	var ReportInterval uint64
	var PollInterval uint64

	flag.StringVar(&flagRunAddr, "a", defaultRunAddr, "address and port to run server")
	flag.Uint64Var(&flagPollInterval, "p", defaultPollInterval, "data poll interval")
	flag.Uint64Var(&flagReportInterval, "r", defaultReportInterval, "data report interval")
	flag.Parse()

	cfg.Host = flagRunAddr
	envRunAddr, ok := os.LookupEnv("ADDRESS")
	if ok {
		cfg.Host = envRunAddr
	}

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
