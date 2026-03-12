package agent

import (
	"fmt"
	"strings"
)

type AgentConfig struct {
	runtimeMetricNames []string
	ServerAddress      string `env:"ADDRESS"`
	ReportInterval     int    `env:"REPORT_INTERVAL"`
	PollInterval       int    `env:"POLL_INTERVAL"`
}

func GetDefaultAgentConfig() AgentConfig {
	config := AgentConfig{}
	config.runtimeMetricNames = []string{
		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
	}
	config.ServerAddress = "http://localhost:8080"
	config.PollInterval = 2
	config.ReportInterval = 10
	return config
}

func (agentConfig *AgentConfig) CheckAndCorrectServerAddress() {
	if !strings.Contains(agentConfig.ServerAddress, "http") {
		agentConfig.ServerAddress = fmt.Sprintf("http://%s", agentConfig.ServerAddress)
	}
}
