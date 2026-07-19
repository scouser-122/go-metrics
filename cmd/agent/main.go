package main

import (
	"github.com/scouser-122/go-metrics/internal/agent"
	"github.com/scouser-122/go-metrics/internal/logger"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	config := agent.AgentConfig{}
	config.Load()
	if err := logger.Initialize(config.LogLevel, config.Environment); err != nil {
		panic(err)
	}
	printBuildVersion()
	agent := agent.NewAgent(&config)
	agent.CollectAndSendMetricsInLoop()
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
