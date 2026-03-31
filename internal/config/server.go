package config

type ServerConfig struct {
	RunAddr          string `env:"ADDRESS"`
	LogLevel         string `env:"LOG_LEVEL"`
	Environment      string `env:"SERVICE_ENVIRONMENT"`
	StoreInterval    int    `env:"STORE_INTERVAL"`
	StorePath        string `env:"FILE_STORAGE_PATH"`
	Restore          bool   `env:"RESTORE"`
	DbDataSourceName string `env:"DATABASE_DSN"`
}

func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		RunAddr:          "localhost:8080",
		LogLevel:         "info",
		Environment:      "dev",
		StoreInterval:    -1,
		StorePath:        "",
		Restore:          false,
		DbDataSourceName: "postgres://postgres:password@localhost:5432/mydb?sslmode=disable",
	}
}
