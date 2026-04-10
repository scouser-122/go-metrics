package db

import "github.com/scouser-122/go-metrics/internal/config"

type DBConnectionConfig struct {
	DSN         string
	RetryConfig config.RetryConfig
}
