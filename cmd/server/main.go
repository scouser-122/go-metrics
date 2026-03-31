package main

import (
	"net/http"

	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/config/db"
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

	database := db.Database{
		Config: db.DBConnectionConfig{
			DSN: config.DbDataSourceName,
		},
	}
	if err := database.Open(); err != nil {
		logger.Sugar.Fatalf("cannot connect to database: %w", err)
	}
	defer database.Close()

	memStorage := repository.MemStorage{}

	metricsService := service.MetricsService{
		Storage: &memStorage,
	}
	metricsService.Initialize(&config)

	handlers := handler.InitializeHandlers(&metricsService, &database)

	r := handler.CreateChiRouter(&handlers)

	logger.Sugar.Infof("starting server on http://%s", config.RunAddr)
	logger.Sugar.Fatal(http.ListenAndServe(config.RunAddr, r))
}
