package main

import (
	"github.com/scouser-122/go-metrics/internal/agent"
)

func main() {
	agent := agent.RuntimeMetricsAgent{
		Config: agent.GetDefaultAgentConfig(),
	}
	agent.CollectAndSendMetricsInLooop()
}
