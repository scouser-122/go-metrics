package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
	"github.com/scouser-122/go-metrics/internal/config"
)

func parseFlags(config *config.ServerConfig) {
	flag.StringVar(&config.RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&config.LogLevel, "l", "info", "logging level")
	flag.StringVar(&config.Environment, "e", "dev", "environment")
	flag.IntVar(&config.StoreInterval, "i", 300, "time interval in seconds to store metrics in file system")
	flag.StringVar(&config.StorePath, "f", "", "metrics store file path")
	flag.BoolVar(&config.Restore, "r", false, "should restore metrics data from storage file or not")
	flag.StringVar(&config.DBDataSourceName, "d", "", "data source name for database connection")
	flag.StringVar(&config.HMACKey, "k", "", "HMAC key to calculate hash of request")
	flag.StringVar(&config.AuditFile, "audit-file", "", "path to file where audit events should be written")
	flag.StringVar(&config.AuditURL, "audit-url", "", "URL of service where audit events should be sent to")
	flag.BoolVar(&config.ProfileEnabled, "profile-enabled", false, "flag to start profiing server on port 6060")
	flag.StringVar(&config.CryptoKey, "crypto-key", "", "private key path to decode requests")
	flag.Parse()
}

func parseEnvVariables(config *config.ServerConfig) {
	err := env.Parse(config)
	if err != nil {
		log.Fatal(err)
	}
}
