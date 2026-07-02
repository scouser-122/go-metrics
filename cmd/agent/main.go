package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/scouser-122/go-metrics/internal/agent"
	"github.com/scouser-122/go-metrics/internal/logger"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	config := agent.GetDefaultAgentConfig()
	parseFlags(&config)
	parseEnvVariables(&config)
	if err := logger.Initialize(config.LogLevel, config.Environment); err != nil {
		panic(err)
	}
	printBuildVersion()
	agent := agent.NewAgent(&config)
	agent.CollectAndSendMetricsInLoop()
}

func parseFlags(agentConfig *agent.AgentConfig) {
	flag.StringVar(&agentConfig.ServerAddress, "a", "localhost:8080", "server address in format host:port")
	flag.IntVar(&agentConfig.ReportInterval, "r", 10, "metrics report interval in seconds")
	flag.IntVar(&agentConfig.PollInterval, "p", 2, "metrics poll interval in seconds")
	flag.StringVar(&agentConfig.LogLevel, "log", "info", "logging level")
	flag.StringVar(&agentConfig.Environment, "e", "dev", "agent environment")
	flag.StringVar(&agentConfig.HMACKey, "k", "", "HMAC key to calculate hash of request")
	flag.IntVar(&agentConfig.RequestRateLimit, "l", 2, "send metrics request rate limit")
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

func printBuildVersion() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}
	logger.Sugar.Infof("\nBuild version: %s\nBuild date: %s\nBuild commit: %s", buildVersion, buildDate, buildCommit)
}
