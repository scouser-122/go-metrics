package main

import (
	"log"
	"net/http"

	"github.com/scouser-122/go-metrics/internal/handler"
	"github.com/scouser-122/go-metrics/internal/logger"
	"github.com/scouser-122/go-metrics/internal/repository"
	"github.com/scouser-122/go-metrics/internal/service"
)

func main() {
	parseFlags()
	parseEnvVariables()
	if err := logger.Initialize(flagLogLevel); err != nil {
		panic(err)
	}

	memStorage := repository.MemStorage{}

	metricsService := service.MetricsService{
		Storage: &memStorage,
	}

	handlers := handler.InitializeHandlers(&metricsService)

	r := handler.CreateChiRouter(handlers)

	logger.Log.Sugar().Infof("Starting server on http://%s\n", flagRunAddr)
	log.Fatal(http.ListenAndServe(flagRunAddr, r))
}
