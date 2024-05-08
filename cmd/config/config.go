package config

import (
	"flag"
	"os"
)

type ServerConfig struct {
	Server struct {
		Host string `yaml:"host"`
	} `yaml:"server"`
}

type AgentConfig struct {
	Agent struct {
		Host           string `yaml:"host"`
		PollInterval   string `yaml:"pollinterval"`
		ReportInterval string `yaml:"reportinterval"`
	} `yaml:"agent"`
}

func NewServerConfig() (*ServerConfig, error) {
	defaultRunAddr := "localhost:8080"
	var flagRunAddr string

	config := &ServerConfig{}

	flag.StringVar(&flagRunAddr, "a", defaultRunAddr, "address and port to run server")
	flag.Parse()
	envRunAddr, ok := os.LookupEnv("ADDRESS")
	if !ok {
		config.Server.Host = flagRunAddr
	} else {
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

	envRunAddr, ok := os.LookupEnv("ADDRESS")
	if !ok {
		config.Agent.Host = flagRunAddr
	} else {
		config.Agent.Host = envRunAddr
	}

	envReportInterval, ok := os.LookupEnv("REPORT_INTERVAL")
	if !ok {
		config.Agent.ReportInterval = flagReportInterval
	} else {
		config.Agent.ReportInterval = envReportInterval
	}

	envPollInterval, ok := os.LookupEnv("POLL_INTERVAL")
	if !ok {
		config.Agent.PollInterval = flagPollInterval
	} else {
		config.Agent.PollInterval = envPollInterval
	}
	return config, nil
}
