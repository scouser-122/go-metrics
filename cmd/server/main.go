package main

import (
	"net/http"

	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/handler"
	"github.com/scouser-122/go-metrics/internal/logger"
	"github.com/scouser-122/go-metrics/internal/repository"
	"github.com/scouser-122/go-metrics/internal/service"
)

func main() {
	config := config.DefaultServerConfig()
	parseFlags(&config)
	parseEnvVariables(&config)
	if err := logger.Initialize(config.LogLevel, config.Environment); err != nil {
		panic(err)
	}

	memStorage := repository.MemStorage{}

	metricsService := service.MetricsService{
		Storage: &memStorage,
	}
	metricsService.Initialize(&config)

	handlers := handler.InitializeHandlers(&metricsService)

	r := handler.CreateChiRouter(&handlers)

	logger.Sugar.Infof("starting server on http://%s", config.RunAddr)
	logger.Sugar.Fatal(http.ListenAndServe(config.RunAddr, r))
}
