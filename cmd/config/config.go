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
	defaultRunAddrServer := "localhost:8080"
	flagRunAddrServer := ""

	config := &ServerConfig{}
	flag.StringVar(&flagRunAddrServer, "a", defaultRunAddrServer, "address and port to run server")
	flag.Parse()

	if envRunAddrServer := os.Getenv("ADDRESS"); envRunAddrServer != "" {
		config.Server.Host = envRunAddrServer
	} else if IsFlagPassed("a") {
		config.Server.Host = flagRunAddrServer
	} else {
		config.Server.Host = defaultRunAddrServer
	}
	return config, nil
}

func NewAgentConfig() (*AgentConfig, error) {
	defaultRunAddrAgent := "localhost:8080"
	defaultReportIntervalAgent := "10"
	defaultPollIntervalAgent := "2"

	flagRunAddrAgent := ""
	flagReportIntervalAgent := ""
	flagPollIntervalAgent := ""

	config := &AgentConfig{}

	flag.StringVar(&flagRunAddrAgent, "a", defaultRunAddrAgent, "address and port to run agent")
	flag.StringVar(&flagReportIntervalAgent, "r", defaultReportIntervalAgent, "data report interval")
	flag.StringVar(&flagPollIntervalAgent, "p", defaultPollIntervalAgent, "data poll interval")
	flag.Parse()

	config.Agent.Host = setParam("ADDRESS", "a", flagRunAddrAgent, defaultRunAddrAgent)
	config.Agent.ReportInterval = setParam("REPORT_INTERVAL", "r", flagReportIntervalAgent, defaultReportIntervalAgent)
	config.Agent.PollInterval = setParam("POLL_INTERVAL", "p", flagPollIntervalAgent, defaultPollIntervalAgent)

	return config, nil
}

func setParam(envParamName, flagName, flagValue, defaultValue string) string {
	retValue := ""

	if paramValue := os.Getenv(envParamName); paramValue != "" {
		retValue = paramValue
	} else if IsFlagPassed(flagName) {
		switch flagName {
		case "a":
			retValue = flagValue
		case "r":
			retValue = flagValue
		case "p":
			retValue = flagValue
		}
	} else {
		retValue = defaultValue
	}
	return retValue
}

func IsFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
