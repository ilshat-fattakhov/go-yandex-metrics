package config

import (
	"flag"
	"fmt"
	"os"
)

type ServerConfig struct {
	Server struct {
		Host string
	}
}

type AgentConfig struct {
	Agent struct {
		Host           string
		PollInterval   string
		ReportInterval string
	}
}

func NewServerConfig() (*ServerConfig, error) {
	defaultRunAddr := "localhost:8080"
	var flagRunAddr string

	config := &ServerConfig{}

	flag.StringVar(&flagRunAddr, "a", defaultRunAddr, "address and port to run server")
	flag.Parse()

	config.Server.Host = flagRunAddr
	envRunAddr, ok := os.LookupEnv("ADDRESS")
	fmt.Println(envRunAddr)
	if ok {
		config.Server.Host = envRunAddr
	}
	return config, nil
}

func NewAgentConfig() (*AgentConfig, error) {
	defaultRunAddr := "localhost:8080"
	defaultReportInterval := "10"
	defaultPollInterval := "2"

	var flagRunAddr string
	var flagReportInterval string
	var flagPollInterval string

	config := &AgentConfig{}

	flag.StringVar(&flagRunAddr, "a", defaultRunAddr, "address and port to run server")
	flag.StringVar(&flagPollInterval, "p", defaultPollInterval, "data poll interval")
	flag.StringVar(&flagReportInterval, "r", defaultReportInterval, "data report interval")
	flag.Parse()

	config.Agent.Host = flagRunAddr
	envRunAddr, ok := os.LookupEnv("ADDRESS")
	if ok {
		config.Agent.Host = envRunAddr
	}

	config.Agent.ReportInterval = flagReportInterval
	envReportInterval, ok := os.LookupEnv("REPORT_INTERVAL")
	if ok {
		config.Agent.ReportInterval = envReportInterval
	}

	config.Agent.PollInterval = flagPollInterval
	envPollInterval, ok := os.LookupEnv("POLL_INTERVAL")
	if ok {
		config.Agent.PollInterval = envPollInterval
	}
	return config, nil
}
