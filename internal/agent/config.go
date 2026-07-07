package agent

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/utils"
)

//go:generate go run github.com/scouser-122/go-metrics/cmd/reset

// AgentConfig holds all configuration settings for the metrics agent.
// generate:reset
type AgentConfig struct {
	RuntimeMetricNames []string
	ServerAddress      string             `env:"ADDRESS" json:"address"`
	ReportInterval     int                `env:"REPORT_INTERVAL" json:"report_interval"`
	PollInterval       int                `env:"POLL_INTERVAL" json:"poll_interval"`
	LogLevel           string             `env:"LOG_LEVEL" json:"log_level"`
	Environment        string             `env:"AGENT_ENVIRONMENT" json:"agent_environment"`
	HMACKey            string             `env:"KEY" json:"key"`
	RequestRateLimit   int                `env:"RATE_LIMIT" json:"rate_limit"`
	CryptoKey          string             `env:"CRYPTO_KEY" json:"crypto_key"`
	RetryConfig        config.RetryConfig `json:"retry_config"`
	CollectChannelSize int                `json:"collect_channel_size"`
	ConfigFile         string             `env:"CONFIG"`
}

// GetDefaultAgentConfig returns an AgentConfig instance with default values.
func GetDefaultAgentConfig() AgentConfig {
	agentConfig := AgentConfig{}
	agentConfig.RuntimeMetricNames = []string{
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

// Load sequentually loads config from different sources
func (a *AgentConfig) Load() {
	a.parseFlags()
	a.parseEnvVariables()
	a.overrideFromLocalFileIfExists()
	a.overrideFromDefault()
	a.checkAndCorrectServerAddress()
}

func (a *AgentConfig) parseFlags() {
	flag.StringVar(&a.ServerAddress, "a", "", "server address in format host:port")
	flag.IntVar(&a.ReportInterval, "r", 0, "metrics report interval in seconds")
	flag.IntVar(&a.PollInterval, "p", 0, "metrics poll interval in seconds")
	flag.StringVar(&a.LogLevel, "log", "", "logging level")
	flag.StringVar(&a.Environment, "e", "", "agent environment")
	flag.StringVar(&a.HMACKey, "k", "", "HMAC key to calculate hash of request")
	flag.IntVar(&a.RequestRateLimit, "l", 0, "send metrics request rate limit")
	flag.StringVar(&a.CryptoKey, "crypto-key", "", "public key path to encode requests")
	flag.StringVar(&a.ConfigFile, "c", "", "path to config file")
	flag.StringVar(&a.ConfigFile, "config", "", "path to config file")
	flag.Parse()
}

func (a *AgentConfig) parseEnvVariables() {
	err := env.Parse(a)
	if err != nil {
		log.Fatal(err)
	}
}

func (a *AgentConfig) checkAndCorrectServerAddress() {
	if !strings.Contains(a.ServerAddress, "http") {
		a.ServerAddress = fmt.Sprintf("http://%s", a.ServerAddress)
	}
}

// overrideFromLocalFileIfExists overrides parameters which were not set by flags or env variables
func (a *AgentConfig) overrideFromLocalFileIfExists() {
	if a.ConfigFile == "" {
		fmt.Printf("config file not specified\n")
		return
	}
	var dirPath string
	lastSlash := strings.LastIndex(a.ConfigFile, "/")
	if lastSlash >= 0 {
		dirPath = a.ConfigFile[:lastSlash]
	}

	root, err := os.OpenRoot(dirPath)
	if err != nil {
		fmt.Printf("open config file directory error: %q\n", err)
		return
	}
	defer root.Close()

	fileName := a.ConfigFile[lastSlash+1:]
	file, err := root.Open(fileName)
	if err != nil {
		fmt.Printf("open config file error: %q\n", err)
		return
	}
	defer file.Close()

	var configFromFile AgentConfig
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configFromFile)
	if err != nil {
		fmt.Printf("read config file error: %q\n", err)
		return
	}

	utils.MergeStructs(a, &configFromFile)
	fmt.Printf("successfully loaded config file %q\n", a.ConfigFile)
}

// overrideFromDefault overrides parameters which were not set by flags or env variables
func (a *AgentConfig) overrideFromDefault() {
	defaultConfig := GetDefaultAgentConfig()
	utils.MergeStructs(a, &defaultConfig)
}
