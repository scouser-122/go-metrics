package main

import (
	"flag"
	"os"
)

var flagRunAddr = "localhost:8080"

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.Parse()
}

func parseEnvVariables() {
	runAddress := os.Getenv("ADDRESS")
	if runAddress != "" {
		flagRunAddr = runAddress
	}
}
