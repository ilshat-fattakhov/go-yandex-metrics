package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type ServerCfg struct {
	Host       string `json:"host"`
	HashKey    string
	StorageCfg StorageCfg
}

type StorageCfg struct {
	FileStoragePath string
	DatabaseDSN     string
	StoreInterval   uint64 `json:"store_interval"`
	Restore         bool   `json:"restore"`
}

type AgentCfg struct {
	Host           string `json:"host"`
	HashKey        string
	PollInterval   uint64 `json:"poll_interval"`
	ReportInterval uint64 `json:"report_interval"`
	RateLimit      uint64
}

func NewServerConfig() (ServerCfg, error) {
	var cfg ServerCfg
	var storageCfg StorageCfg

	const defaultRunAddr = "localhost:8080"
	const defaultStoreInterval uint64 = 300               // значение 0 делает запись синхронной
	const defaultFileStoragePath = "/tmp/metrics-db.json" // пустое значение отключает функцию записи на диск
	const defaultRestore = true
	const defaultDatabaseDSN = ""
	const defaultHashKey = ""

	var flagRunAddr string
	var flagStoreInterval uint64
	var flagFileStoragePath string
	var flagRestore bool
	var flagDatabaseDSN string
	var flagHashKey string

	flag.StringVar(&flagRunAddr, "a", defaultRunAddr, "address and port to run server")
	flag.BoolVar(&flagRestore, "r", defaultRestore, "restore data from file at server start")
	flag.Uint64Var(&flagStoreInterval, "i", defaultStoreInterval, "data storing interval")

	flag.StringVar(&flagFileStoragePath, "f", defaultFileStoragePath, "file storage path")
	flag.StringVar(&flagDatabaseDSN, "d", defaultDatabaseDSN, "DB connection string")
	flag.StringVar(&flagHashKey, "k", defaultHashKey, "hash key")

	flag.Parse()

	cfg.Host = flagRunAddr
	envRunAddr, ok := os.LookupEnv("ADDRESS")
	if ok {
		cfg.Host = envRunAddr
	}

	storageCfg.StoreInterval = flagStoreInterval
	envStoreInterval, ok := os.LookupEnv("STORE_INTERVAL")
	if ok {
		tmpStoreInterval, err := strconv.ParseUint(envStoreInterval, 10, 64)
		if err != nil {
			return cfg, fmt.Errorf("failed to parse %d as a report interval value: %w", tmpStoreInterval, err)
		}
		storageCfg.StoreInterval = tmpStoreInterval
	}

	storageCfg.Restore = flagRestore
	envRestore, ok := os.LookupEnv("RESTORE")
	if ok {
		boolValue, err := strconv.ParseBool(envRestore)
		if err != nil {
			return cfg, fmt.Errorf("an error occured parsing bool value: %w", err)
		}
		storageCfg.Restore = boolValue
	}

	storageCfg.DatabaseDSN = flagDatabaseDSN
	envflagDatabaseDSN, ok := os.LookupEnv("DATABASE_DSN")
	if ok {
		storageCfg.DatabaseDSN = envflagDatabaseDSN
	}

	cfg.HashKey = flagHashKey
	envHashKey, ok := os.LookupEnv("KEY")
	if ok {
		cfg.HashKey = envHashKey
	}

	storageCfg.FileStoragePath = flagFileStoragePath
	envFileStoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH")
	if ok {
		storageCfg.FileStoragePath = envFileStoragePath
	}

	cfg.StorageCfg = storageCfg
	return cfg, nil
}

func NewAgentConfig() (AgentCfg, error) {
	var cfg AgentCfg

	const defaultRunAddr = "localhost:8080"
	const defaultReportInterval uint64 = 10
	const defaultPollInterval uint64 = 2
	const defaultHashKey = ""
	const defaultRateLimit = 0

	var flagRunAddr string
	var flagReportInterval uint64
	var flagPollInterval uint64
	var flagHashKey string
	var flagRateLimit uint64

	flag.StringVar(&flagRunAddr, "a", defaultRunAddr, "address and port to run server")
	flag.Uint64Var(&flagPollInterval, "p", defaultPollInterval, "data poll interval")
	flag.Uint64Var(&flagReportInterval, "r", defaultReportInterval, "data report interval")
	flag.StringVar(&flagHashKey, "k", defaultHashKey, "hash key")
	flag.Uint64Var(&flagRateLimit, "l", defaultRateLimit, "data report interval")

	flag.Parse()

	cfg.Host = flagRunAddr
	envRunAddr, ok := os.LookupEnv("ADDRESS")
	if ok {
		cfg.Host = envRunAddr
	}

	cfg.ReportInterval = flagReportInterval
	envReportInterval, ok := os.LookupEnv("REPORT_INTERVAL")
	if ok {
		ReportInterval, err := strconv.ParseUint(envReportInterval, 10, 64)
		if err != nil {
			return cfg, fmt.Errorf("failed to parse %s as a report interval value: %w", envReportInterval, err)
		}
		cfg.ReportInterval = ReportInterval
	}

	cfg.PollInterval = flagPollInterval
	envPollInterval, ok := os.LookupEnv("POLL_INTERVAL")
	if ok {
		PollInterval, err := strconv.ParseUint(envPollInterval, 10, 64)
		if err != nil {
			return cfg, fmt.Errorf("failed to parse %s as a poll interval value: %w", envPollInterval, err)
		}
		cfg.PollInterval = PollInterval
	}

	cfg.HashKey = flagHashKey
	envHashKey, ok := os.LookupEnv("KEY")
	if ok {
		cfg.HashKey = envHashKey
	}
	cfg.RateLimit = flagRateLimit
	envRateLimit, ok := os.LookupEnv("RATE_LIMIT")
	if ok {
		RateLimit, err := strconv.ParseUint(envRateLimit, 10, 64)
		if err != nil {
			return cfg, fmt.Errorf("failed to parse %s as a rate limit value: %w", envRateLimit, err)
		}
		cfg.RateLimit = RateLimit
	}

	return cfg, nil
}
