package main

import (
	"flag"
	"os"
)

var (
	flagRunAddr     string
	flagLogLevel    string
	flagEnvironment string
)

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&flagLogLevel, "l", "info", "logging level")
	flag.StringVar(&flagEnvironment, "e", "dev", "environment")
	flag.Parse()
}

func parseEnvVariables() {
	runAddress := os.Getenv("ADDRESS")
	if runAddress != "" {
		flagRunAddr = runAddress
	}
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel != "" {
		flagLogLevel = logLevel
	}
	environment := os.Getenv("SERVICE_ENVIRONMENT")
	if environment != "" {
		flagEnvironment = environment
	}
}
