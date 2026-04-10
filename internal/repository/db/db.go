package db

import (
	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/config/db"
)

func NewPostgresDB(serverConfig config.ServerConfig) PostgresDatabase {
	database := PostgresDatabase{
		Config: db.DBConnectionConfig{
			DSN:         serverConfig.DBDataSourceName,
			RetryConfig: config.DefaultRetryConfig(),
		},
	}
	return database
}
