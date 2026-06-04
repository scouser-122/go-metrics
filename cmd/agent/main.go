package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/scouser-122/go-metrics/internal/agent"
	"github.com/scouser-122/go-metrics/internal/logger"
)

func main() {
	config := agent.GetDefaultAgentConfig()
	parseFlags(&config)
	parseEnvVariables(&config)
	if err := logger.Initialize(config.LogLevel, config.Environment); err != nil {
		panic(err)
	}
	agent := agent.RuntimeMetricsAgent{
		Config: config,
	}
	agent.CollectAndSendMetricsInLoop()
}

func parseFlags(agentConfig *agent.AgentConfig) {
	flag.StringVar(&agentConfig.ServerAddress, "a", "localhost:8080", "server address in format host:port")
	flag.IntVar(&agentConfig.ReportInterval, "r", 10, "metrics report interval in seconds")
	flag.IntVar(&agentConfig.PollInterval, "p", 2, "metrics poll interval in seconds")
	flag.StringVar(&agentConfig.LogLevel, "l", "info", "logging level")
	flag.StringVar(&agentConfig.Environment, "e", "dev", "agent environment")
	flag.StringVar(&agentConfig.HMACKey, "k", "", "HMAC key to calculate hash of request")
	flag.Parse()
	agentConfig.CheckAndCorrectServerAddress()
}

func parseEnvVariables(agentConfig *agent.AgentConfig) {
	err := env.Parse(agentConfig)
	if err != nil {
		log.Fatal(err)
	}
	agentConfig.CheckAndCorrectServerAddress()
}
