package main

import (
	"net/http"

	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/handler"
	"github.com/scouser-122/go-metrics/internal/logger"
	"github.com/scouser-122/go-metrics/internal/repository/db"
	"github.com/scouser-122/go-metrics/internal/service"
)

func main() {
	serverConfig := config.DefaultServerConfig()
	parseFlags(&serverConfig)
	parseEnvVariables(&serverConfig)
	if err := logger.Initialize(serverConfig.LogLevel, serverConfig.Environment); err != nil {
		panic(err)
	}

	database := db.NewPostgresDB(serverConfig)
	if err := database.Open(); err != nil {
		logger.Sugar.Errorf("cannot connect to database: %w", err)
	}
	defer database.Close()

	metricsService := service.MetricsService{}
	metricsService.Initialize(&serverConfig, &database)

	cryptoService := service.CryptoService{
		ServerConfig: &serverConfig,
	}

	handlers := handler.InitializeHandlers(&metricsService, &cryptoService, &database)

	r := handler.CreateChiRouter(&handlers)

	logger.Sugar.Infof("starting server on http://%s", serverConfig.RunAddr)
	logger.Sugar.Fatal(http.ListenAndServe(serverConfig.RunAddr, r))
}
