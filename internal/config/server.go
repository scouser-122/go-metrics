package config

type ServerConfig struct {
	RunAddr          string `env:"ADDRESS"`
	LogLevel         string `env:"LOG_LEVEL"`
	Environment      string `env:"SERVICE_ENVIRONMENT"`
	StoreInterval    int    `env:"STORE_INTERVAL"`
	StorePath        string `env:"FILE_STORAGE_PATH"`
	Restore          bool   `env:"RESTORE"`
	DBDataSourceName string `env:"DATABASE_DSN"`
	HMACKey          string `env:"KEY"`

	// AuditFile path to file where audit events should be written
	AuditFile string `env:"AUDIT_FILE"`

	// AuditURL URL of service where audit events should be sent to
	AuditURL string `env:"AUDIT_URL"`

	// ProfileEnabled flag to start profiing server on port 6060
	ProfileEnabled bool `env:"PROFILE_ENABLED"`
}

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
