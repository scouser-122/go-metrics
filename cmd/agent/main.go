package main

import (
	"flag"
	"fmt"

	"github.com/scouser-122/go-metrics/internal/agent"
)

func main() {
	config := agent.GetDefaultAgentConfig()
	initFlags(&config)
	flag.Parse()
	agent := agent.RuntimeMetricsAgent{
		Config: config,
	}
	agent.CollectAndSendMetricsInLooop()
}

func initFlags(agentConfig *agent.AgentConfig) {
	serverAddress := ""
	flag.StringVar(&serverAddress, "a", "localhost:8080", "server address in format host:port")
	agentConfig.ServerAddress = fmt.Sprintf("http://%s", serverAddress)
	flag.IntVar(&agentConfig.ReportInterval, "r", 10, "metrics report interval in seconds")
	flag.IntVar(&agentConfig.PollInterval, "p", 2, "metrics poll interval in seconds")
}
