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
	flag.StringVar(&config.StorePath, "f", "./metrics_data.json", "metrics store file path")
	flag.BoolVar(&config.Restore, "r", false, "should restore metrics data from storage file or not")
	flag.StringVar(&config.DbDataSourceName, "d", "", "data source name for database connection")
	flag.Parse()
}

func parseEnvVariables(config *config.ServerConfig) {
	err := env.Parse(config)
	if err != nil {
		log.Fatal(err)
	}
}
