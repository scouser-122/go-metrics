package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

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

	// ShutdownTimeout timeout which server will wait to finist processing requests before shutdown
	ShutdownTimeout time.Duration `env:"SHUTDOW_TIMEOUT"`
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
		ShutdownTimeout:  30 * time.Second,
	}
}

// Load sequentually loads config from different sources
func (s *ServerConfig) Load() {
	s.parseFlags()
	s.parseEnvVariables()
	s.overrideFromLocalFileIfExists()
	s.overrideFromDefault()
}

func (s *ServerConfig) parseFlags() {
	flag.StringVar(&s.RunAddr, "a", "", "address and port to run server")
	flag.StringVar(&s.LogLevel, "l", "", "logging level")
	flag.StringVar(&s.Environment, "e", "", "environment")
	flag.IntVar(&s.StoreInterval, "i", 0, "time interval in seconds to store metrics in file system")
	flag.StringVar(&s.StorePath, "f", "", "metrics store file path")
	flag.BoolVar(&s.Restore, "r", false, "should restore metrics data from storage file or not")
	flag.StringVar(&s.DBDataSourceName, "d", "", "data source name for database connection")
	flag.StringVar(&s.HMACKey, "k", "", "HMAC key to calculate hash of request")
	flag.StringVar(&s.AuditFile, "audit-file", "", "path to file where audit events should be written")
	flag.StringVar(&s.AuditURL, "audit-url", "", "URL of service where audit events should be sent to")
	flag.BoolVar(&s.ProfileEnabled, "profile-enabled", false, "flag to start profiing server on port 6060")
	flag.StringVar(&s.CryptoKey, "crypto-key", "", "private key path to decode requests")
	flag.StringVar(&s.ConfigFile, "c", "", "path to config file")
	flag.StringVar(&s.ConfigFile, "config", "", "path to config file")
	flag.Parse()
}

func (s *ServerConfig) parseEnvVariables() {
	err := env.Parse(s)
	if err != nil {
		log.Fatal(err)
	}
}

// overrideFromLocalFileIfExists overrides parameters which were not set by flags or env variables
func (s *ServerConfig) overrideFromLocalFileIfExists() {
	if s.ConfigFile == "" {
		fmt.Printf("config file not specified\n")
		return
	}
	var dirPath string
	lastSlash := strings.LastIndex(s.ConfigFile, "/")
	if lastSlash >= 0 {
		dirPath = s.ConfigFile[:lastSlash]
	}

	root, err := os.OpenRoot(dirPath)
	if err != nil {
		fmt.Printf("open config file directory error: %q\n", err)
		return
	}
	defer root.Close()

	fileName := s.ConfigFile[lastSlash+1:]
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

	utils.MergeStructs(s, &configFromFile)
	fmt.Printf("successfully loaded config file %q\n", s.ConfigFile)
}

// overrideFromDefault overrides parameters which were not set by flags or env variables
func (s *ServerConfig) overrideFromDefault() {
	defaultConfig := DefaultServerConfig()
	utils.MergeStructs(s, &defaultConfig)
}
