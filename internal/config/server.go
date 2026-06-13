package config

// ServerConfig holds all configuration settings for the metrics server.
type ServerConfig struct {
	// RunAddr network address to run service on.
	RunAddr string `env:"ADDRESS"`

	// LogLevel logging level. Example: error, warn, info, debug.
	LogLevel string `env:"LOG_LEVEL"`

	// Environment runnng environment. Example - dev or prod.
	Environment string `env:"SERVICE_ENVIRONMENT"`

	// StoreInterval time inteval to store metrics in file.
	StoreInterval int `env:"STORE_INTERVAL"`

	// StorePath local metrics storage file path.
	StorePath string `env:"FILE_STORAGE_PATH"`

	// Restore flag to restore metrics from local storage.
	Restore bool `env:"RESTORE"`

	// DBDataSourceName DB data source address.
	DBDataSourceName string `env:"DATABASE_DSN"`

	// HMACKey is a key used to check request body hash.
	HMACKey string `env:"KEY"`

	// AuditFile path to file where audit events should be written.
	AuditFile string `env:"AUDIT_FILE"`

	// AuditURL URL of service where audit events should be sent to.
	AuditURL string `env:"AUDIT_URL"`

	// ProfileEnabled flag to start profiing server on port 6060.
	ProfileEnabled bool `env:"PROFILE_ENABLED"`
}

// DefaultServerConfig returns a ServerConfig instance with default values.
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		RunAddr:          "localhost:8080",
		LogLevel:         "info",
		Environment:      "dev",
		StoreInterval:    -1,
		StorePath:        "",
		Restore:          false,
		DBDataSourceName: "postgres://postgres:password@localhost:5432/mydb?sslmode=disable",
		HMACKey:          "",
		AuditFile:        "",
		AuditURL:         "",
		ProfileEnabled:   false,
	}
}
