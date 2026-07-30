package agent

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/scouser-122/go-metrics/internal/config"
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
	GrpcServerAddress  string             `env:"GRPC_SERVER_ADDRESS" json:"grpc_server_address"`
}

func (a *AgentConfig) merge(other *AgentConfig) {
	if len(other.RuntimeMetricNames) > 0 {
		namesEqual := true
		if len(other.RuntimeMetricNames) != len(a.RuntimeMetricNames) {
			namesEqual = false
		} else {
			for _, n := range other.RuntimeMetricNames {
				if slices.Index(a.RuntimeMetricNames, n) == -1 {
					namesEqual = false
					break
				}
			}
		}
		if !namesEqual {
			a.RuntimeMetricNames = other.RuntimeMetricNames
		}
	}
	if other.ServerAddress != "" && other.ServerAddress != a.ServerAddress {
		a.ServerAddress = other.ServerAddress
	}
	if other.ReportInterval != a.ReportInterval {
		a.ReportInterval = other.ReportInterval
	}
	if other.PollInterval != a.PollInterval {
		a.PollInterval = other.PollInterval
	}
	if other.LogLevel != "" && other.LogLevel != a.LogLevel {
		a.LogLevel = other.LogLevel
	}
	if other.Environment != "" && other.Environment != a.Environment {
		a.Environment = other.Environment
	}
	if other.HMACKey != "" && other.HMACKey != a.HMACKey {
		a.HMACKey = other.HMACKey
	}
	if other.RequestRateLimit != a.RequestRateLimit {
		a.RequestRateLimit = other.RequestRateLimit
	}
	if other.CryptoKey != "" && other.CryptoKey != a.CryptoKey {
		a.CryptoKey = other.CryptoKey
	}
	if other.RetryConfig.MaxAttempts != a.RetryConfig.MaxAttempts {
		a.RetryConfig.MaxAttempts = other.RetryConfig.MaxAttempts
	}
	if other.RetryConfig.InitialBackoff != a.RetryConfig.InitialBackoff {
		a.RetryConfig.InitialBackoff = other.RetryConfig.InitialBackoff
	}
	if other.RetryConfig.BackoffMultiplier != a.RetryConfig.BackoffMultiplier {
		a.RetryConfig.BackoffMultiplier = other.RetryConfig.BackoffMultiplier
	}
	if other.CollectChannelSize != a.CollectChannelSize {
		a.CollectChannelSize = other.CollectChannelSize
	}
	if other.GrpcServerAddress != "" && other.GrpcServerAddress != a.GrpcServerAddress {
		a.GrpcServerAddress = other.GrpcServerAddress
	}
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
	agentConfig.ReportInterval = 7
	agentConfig.LogLevel = "info"
	agentConfig.Environment = "dev"
	agentConfig.RetryConfig = config.DefaultRetryConfig()
	agentConfig.RequestRateLimit = 3
	agentConfig.CollectChannelSize = 50
	return agentConfig
}

type agentConfigFlags struct {
	serverAddress     *string
	logLevel          *string
	environment       *string
	reportInterval    *int
	pollInterval      *int
	configFile        *string
	configFileLong    *string
	fHMACKey          *string
	requestRateLimit  *int
	cryptoKey         *string
	grpcServerAddress *string
}

func (af *agentConfigFlags) define() {
	af.serverAddress = flag.String("a", "", "server address in format host:port")
	af.logLevel = flag.String("log", "debug", "logging level")
	af.environment = flag.String("e", "prod", "environment")
	af.reportInterval = flag.Int("r", 7, "metrics report interval in seconds")
	af.pollInterval = flag.Int("p", 2, "metrics poll interval in seconds")
	af.configFile = flag.String("c", "", "path to config file")
	af.configFileLong = flag.String("config", "prod", "environment")
	af.fHMACKey = flag.String("k", "", "HMAC key to calculate hash of request")
	af.requestRateLimit = flag.Int("l", 3, "send metrics request rate limit")
	af.cryptoKey = flag.String("crypto-key", "", "private key path to decode requests")
	af.grpcServerAddress = flag.String("grpc-server-address", "", "gRPC server address")
}

// Load sequentually loads config from different sources
func (a *AgentConfig) Load() {
	defaultConfig := GetDefaultAgentConfig()
	a.merge(&defaultConfig)

	af := agentConfigFlags{}
	af.define()

	a.getConfigFileParam(&af)

	a.overrideFromLocalFileIfExists()

	a.parseFlags(&af)
	a.parseEnvVariables()

	a.checkAndCorrectServerAddress()
}

func (a *AgentConfig) getConfigFileParam(af *agentConfigFlags) {
	flag.Parse()
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "c":
			a.ConfigFile = *af.configFile
		case "config":
			a.ConfigFile = *af.configFileLong
		}
	})
	envConfigFile := os.Getenv("CONFIG")
	if envConfigFile != "" {
		a.ConfigFile = envConfigFile
	}
}

func (a *AgentConfig) parseFlags(af *agentConfigFlags) {
	flag.Parse()
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "a":
			a.ServerAddress = *af.serverAddress
		case "log":
			a.LogLevel = *af.logLevel
		case "e":
			a.Environment = *af.environment
		case "r":
			a.ReportInterval = *af.reportInterval
		case "p":
			a.PollInterval = *af.pollInterval
		case "k":
			a.HMACKey = *af.fHMACKey
		case "l":
			a.RequestRateLimit = *af.requestRateLimit
		case "crypto-key":
			a.CryptoKey = *af.cryptoKey
		case "grpc-server-address":
			a.GrpcServerAddress = *af.grpcServerAddress
		}
	})
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

	configFromFile := GetDefaultAgentConfig()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configFromFile)
	if err != nil {
		fmt.Printf("read config file error: %q\n", err)
		return
	}

	a.merge(&configFromFile)
	fmt.Printf("successfully loaded config file %q\n", a.ConfigFile)
}
