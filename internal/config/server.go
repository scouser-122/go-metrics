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

	// TrustedSubnet specifies range of trusted IP-s to make requests to this server
	TrustedSubnet string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`

	GrpcPort string `env:"GRPC_PORT" json:"grpc_port"`
}

func (s *ServerConfig) merge(other *ServerConfig) {
	if other.RunAddr != "" && other.RunAddr != s.RunAddr {
		s.RunAddr = other.RunAddr
	}
	if other.LogLevel != "" && other.LogLevel != s.LogLevel {
		s.LogLevel = other.LogLevel
	}
	if other.Environment != "" && other.Environment != s.Environment {
		s.Environment = other.Environment
	}
	if other.StoreInterval != s.StoreInterval {
		s.StoreInterval = other.StoreInterval
	}
	if other.StorePath != "" && other.StorePath != s.StorePath {
		s.StorePath = other.StorePath
	}
	if other.Restore != s.Restore {
		s.Restore = other.Restore
	}
	if other.DBDataSourceName != "" && other.DBDataSourceName != s.DBDataSourceName {
		s.DBDataSourceName = other.DBDataSourceName
	}
	if other.HMACKey != "" && other.HMACKey != s.HMACKey {
		s.HMACKey = other.HMACKey
	}
	if other.AuditFile != "" && other.AuditFile != s.AuditFile {
		s.AuditFile = other.AuditFile
	}
	if other.AuditURL != "" && other.AuditURL != s.AuditURL {
		s.AuditURL = other.AuditURL
	}
	if other.ProfileEnabled != s.ProfileEnabled {
		s.ProfileEnabled = other.ProfileEnabled
	}
	if other.CryptoKey != "" && other.CryptoKey != s.CryptoKey {
		s.CryptoKey = other.CryptoKey
	}
	if other.ShutdownTimeout != s.ShutdownTimeout {
		s.ShutdownTimeout = other.ShutdownTimeout
	}
	if other.TrustedSubnet != "" && other.TrustedSubnet != s.TrustedSubnet {
		s.TrustedSubnet = other.TrustedSubnet
	}
	if other.GrpcPort != "" && other.GrpcPort != s.GrpcPort {
		s.GrpcPort = other.GrpcPort
	}
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
		TrustedSubnet:    "",
		GrpcPort:         "",
	}
}

type serverConfigFlags struct {
	runAddr           *string
	logLevel          *string
	environment       *string
	storeInterval     *int
	storePath         *string
	configFile        *string
	configFileLong    *string
	restore           *bool
	fDBDataSourceName *string
	fHMACKey          *string
	auditFile         *string
	auditURL          *string
	profileEnabled    *bool
	cryptoKey         *string
	trustedSubnet     *string
	grpcPort          *string
}

func (sf *serverConfigFlags) define() {
	sf.runAddr = flag.String("a", "", "address and port to run server")
	sf.logLevel = flag.String("l", "debug", "logging level")
	sf.environment = flag.String("e", "prod", "environment")
	sf.storeInterval = flag.Int("i", 0, "time interval in seconds to store metrics in file system")
	sf.storePath = flag.String("f", "", "metrics store file path")
	sf.configFile = flag.String("c", "", "path to config file")
	sf.configFileLong = flag.String("config", "prod", "environment")
	sf.restore = flag.Bool("r", false, "should restore metrics data from storage file or not")
	sf.fDBDataSourceName = flag.String("d", "", "data source name for database connection")
	sf.fHMACKey = flag.String("k", "", "HMAC key to calculate hash of request")
	sf.auditFile = flag.String("audit-file", "", "path to file where audit events should be written")
	sf.auditURL = flag.String("audit-url", "", "URL of service where audit events should be sent to")
	sf.profileEnabled = flag.Bool("profile-enabled", false, "flag to start profiing server on port 6060")
	sf.cryptoKey = flag.String("crypto-key", "", "private key path to decode requests")
	sf.trustedSubnet = flag.String("t", "", "range of trusted IP-s to make requests to this server")
	sf.grpcPort = flag.String("grpc-port", "", "gRPC server port")
}

// Load sequentually loads config from different sources
func (s *ServerConfig) Load() {
	defaultConfig := DefaultServerConfig()
	s.merge(&defaultConfig)

	sf := serverConfigFlags{}
	sf.define()

	s.getConfigFileParam(&sf)

	s.overrideFromLocalFileIfExists()

	s.parseFlags(&sf)
	s.parseEnvVariables()
}

func (s *ServerConfig) getConfigFileParam(sf *serverConfigFlags) {
	flag.Parse()
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "c":
			s.ConfigFile = *sf.configFile
		case "config":
			s.ConfigFile = *sf.configFileLong
		}
	})
	envConfigFile := os.Getenv("CONFIG")
	if envConfigFile != "" {
		s.ConfigFile = envConfigFile
	}
}

func (s *ServerConfig) parseFlags(sf *serverConfigFlags) {
	flag.Parse()
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "a":
			s.RunAddr = *sf.runAddr
		case "l":
			s.LogLevel = *sf.logLevel
		case "e":
			s.Environment = *sf.environment
		case "i":
			s.StoreInterval = *sf.storeInterval
		case "f":
			s.StorePath = *sf.storePath
		case "r":
			s.Restore = *sf.restore
		case "d":
			s.DBDataSourceName = *sf.fDBDataSourceName
		case "k":
			s.HMACKey = *sf.fHMACKey
		case "audit-file":
			s.AuditFile = *sf.auditFile
		case "audit-url":
			s.AuditURL = *sf.auditURL
		case "profile-enabled":
			s.ProfileEnabled = *sf.profileEnabled
		case "crypto-key":
			s.CryptoKey = *sf.cryptoKey
		case "t":
			s.TrustedSubnet = *sf.trustedSubnet
		case "grpc-port":
			s.GrpcPort = *sf.grpcPort
		}
	})
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

	configFromFile := DefaultServerConfig()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configFromFile)
	if err != nil {
		fmt.Printf("read config file error: %q\n", err)
		return
	}

	s.merge(&configFromFile)
	fmt.Printf("successfully loaded config file %q\n", s.ConfigFile)
}
