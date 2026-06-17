package agent

import (
	"fmt"
	"strings"

	"github.com/scouser-122/go-metrics/internal/config"
)

// AgentConfig holds all configuration settings for the metrics agent.
type AgentConfig struct {
	runtimeMetricNames []string
	ServerAddress      string `env:"ADDRESS"`
	ReportInterval     int    `env:"REPORT_INTERVAL"`
	PollInterval       int    `env:"POLL_INTERVAL"`
	LogLevel           string `env:"LOG_LEVEL"`
	Environment        string `env:"AGENT_ENVIRONMENT"`
	HMACKey            string `env:"KEY"`
	RequestRateLimit   int    `env:"RATE_LIMIT"`
	RetryConfig        config.RetryConfig
	CollectChannelSize int
}

// GetDefaultAgentConfig returns an AgentConfig instance with default values.
func GetDefaultAgentConfig() AgentConfig {
	agentConfig := AgentConfig{}
	agentConfig.runtimeMetricNames = []string{
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
	agentConfig.ServerAddress = "http://localhost:8080"
	agentConfig.PollInterval = 2
	agentConfig.ReportInterval = 10
	agentConfig.LogLevel = "info"
	agentConfig.Environment = "dev"
	agentConfig.RetryConfig = config.DefaultRetryConfig()
	agentConfig.CollectChannelSize = 50
	return agentConfig
}

// CheckAndCorrectServerAddress ensures the server address has an HTTP scheme.
func (agentConfig *AgentConfig) CheckAndCorrectServerAddress() {
	if !strings.Contains(agentConfig.ServerAddress, "http") {
		agentConfig.ServerAddress = fmt.Sprintf("http://%s", agentConfig.ServerAddress)
	}
}
