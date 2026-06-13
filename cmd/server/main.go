package main

import (
	"context"
	"log"
	"net/http"

	"github.com/botchris/go-pubsub/provider/memory"
	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/handler"
	"github.com/scouser-122/go-metrics/internal/logger"
	"github.com/scouser-122/go-metrics/internal/repository/db"
	"github.com/scouser-122/go-metrics/internal/service"

	_ "net/http/pprof" // подключаем пакет pprof
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

	eventBroker := memory.NewBroker()
	brokerContext, cancel := context.WithCancel(context.Background())
	defer cancel()
	auditService := service.NewAuditService(&serverConfig)
	auditService.SubscribeToMetricEvents(eventBroker, brokerContext)

	metricsService := service.NewMetricsService(&serverConfig, &database, eventBroker)

	cryptoService := service.CryptoService{
		ServerConfig: &serverConfig,
	}

	handlers := handler.InitializeHandlers(metricsService, &cryptoService, &database)

	r := handler.CreateChiRouter(&handlers)

	if serverConfig.ProfileEnabled == true {
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	logger.Sugar.Infof("starting server on http://%s", serverConfig.RunAddr)
	logger.Sugar.Fatal(http.ListenAndServe(serverConfig.RunAddr, r))
}
