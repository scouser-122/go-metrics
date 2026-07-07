package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/scouser-122/go-metrics/internal/utils"
)

//go:generate go run github.com/scouser-122/go-metrics/cmd/reset

// ServerConfig holds all configuration settings for the metrics server.
// generate:reset
type ServerConfig struct {
	// RunAddr network address to run service on.
	RunAddr string `env:"ADDRESS" json:"address"`

	// LogLevel logging level. Example: error, warn, info, debug.
	LogLevel string `env:"LOG_LEVEL" json:"log_level"`

	// Environment runnng environment. Example - dev or prod.
	Environment string `env:"SERVICE_ENVIRONMENT" json:"service_environment"`

	// StoreInterval time inteval to store metrics in file.
	StoreInterval int `env:"STORE_INTERVAL" json:"store_interval"`

	// StorePath local metrics storage file path.
	StorePath string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`

	// Restore flag to restore metrics from local storage.
	Restore bool `env:"RESTORE" json:"restore"`

	// DBDataSourceName DB data source address.
	DBDataSourceName string `env:"DATABASE_DSN" json:"database_dsn"`

	// HMACKey is a key used to check request body hash.
	HMACKey string `env:"KEY" json:"key"`

	// AuditFile path to file where audit events should be written.
	AuditFile string `env:"AUDIT_FILE" json:"audit_file"`

	// AuditURL URL of service where audit events should be sent to.
	AuditURL string `env:"AUDIT_URL" json:"audit_url"`

	// ProfileEnabled flag to start profiing server on port 6060.
	ProfileEnabled bool `env:"PROFILE_ENABLED" json:"profile_enabled"`

	// CryptoKey path to private key for request body decryption
	CryptoKey string `env:"CRYPTO_KEY" json:"crypto_key"`

	// ConfigFile path to config file
	ConfigFile string `env:"CONFIG"`
}

// DefaultServerConfig returns a ServerConfig instance with default values.
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		RunAddr:          "localhost:8080",
		LogLevel:         "info",
		Environment:      "dev",
		StoreInterval:    0,
		StorePath:        "",
		Restore:          false,
		DBDataSourceName: "postgres://postgres:password@localhost:5432/mydb?sslmode=disable",
		HMACKey:          "",
		AuditFile:        "",
		AuditURL:         "",
		ProfileEnabled:   false,
		ConfigFile:       "",
	}
}

// Load sequentually loads config from different sources
func (c *ServerConfig) Load() {
	c.parseFlags()
	c.parseEnvVariables()
	c.overrideFromLocalFileIfExists()
	c.overrideFromDefault()
}

func (c *ServerConfig) parseFlags() {
	flag.StringVar(&c.RunAddr, "a", "", "address and port to run server")
	flag.StringVar(&c.LogLevel, "l", "", "logging level")
	flag.StringVar(&c.Environment, "e", "", "environment")
	flag.IntVar(&c.StoreInterval, "i", 0, "time interval in seconds to store metrics in file system")
	flag.StringVar(&c.StorePath, "f", "", "metrics store file path")
	flag.BoolVar(&c.Restore, "r", false, "should restore metrics data from storage file or not")
	flag.StringVar(&c.DBDataSourceName, "d", "", "data source name for database connection")
	flag.StringVar(&c.HMACKey, "k", "", "HMAC key to calculate hash of request")
	flag.StringVar(&c.AuditFile, "audit-file", "", "path to file where audit events should be written")
	flag.StringVar(&c.AuditURL, "audit-url", "", "URL of service where audit events should be sent to")
	flag.BoolVar(&c.ProfileEnabled, "profile-enabled", false, "flag to start profiing server on port 6060")
	flag.StringVar(&c.CryptoKey, "crypto-key", "", "private key path to decode requests")
	flag.StringVar(&c.ConfigFile, "c", "", "path to config file")
	flag.StringVar(&c.ConfigFile, "config", "", "path to config file")
	flag.Parse()
}

func (c *ServerConfig) parseEnvVariables() {
	err := env.Parse(c)
	if err != nil {
		log.Fatal(err)
	}
}

// overrideFromLocalFileIfExists overrides parameters which were not set by flags or env variables
func (c *ServerConfig) overrideFromLocalFileIfExists() {
	if c.ConfigFile == "" {
		fmt.Printf("config file not specified\n")
		return
	}
	var dirPath string
	lastSlash := strings.LastIndex(c.ConfigFile, "/")
	if lastSlash >= 0 {
		dirPath = c.ConfigFile[:lastSlash]
	}

	root, err := os.OpenRoot(dirPath)
	if err != nil {
		fmt.Printf("open config file directory error: %q\n", err)
		return
	}
	defer root.Close()

	fileName := c.ConfigFile[lastSlash+1:]
	file, err := root.Open(fileName)
	if err != nil {
		fmt.Printf("open config file error: %q\n", err)
		return
	}
	defer file.Close()

	var configFromFile ServerConfig
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configFromFile)
	if err != nil {
		fmt.Printf("read config file error: %q\n", err)
		return
	}

	utils.MergeStructs(c, &configFromFile)
	fmt.Printf("successfully loaded config file %q\n", c.ConfigFile)
}

// overrideFromDefault overrides parameters which were not set by flags or env variables
func (c *ServerConfig) overrideFromDefault() {
	defaultConfig := DefaultServerConfig()
	utils.MergeStructs(c, &defaultConfig)
}
