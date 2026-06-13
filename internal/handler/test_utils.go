package handler

import (
	"context"

	"github.com/botchris/go-pubsub/provider/memory"
	"github.com/go-chi/chi/v5"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/logger"
	"github.com/scouser-122/go-metrics/internal/repository/db"
	"github.com/scouser-122/go-metrics/internal/service"
)

func createTestRouterPostgresDB(mockDB *db.MockPostgresDBTestData) *chi.Mux {
	serverConfig := config.DefaultServerConfig()
	if err := logger.Initialize(serverConfig.LogLevel, serverConfig.Environment); err != nil {
		panic(err)
	}
	cryptoService := service.CryptoService{
		ServerConfig: &serverConfig,
	}
	mock, err := pgxmock.NewPool()
	if err != nil {
		panic(err)
	}
	defer mock.Close()
	mockDB.PgxPoolIface = mock
	mockDB.MockDBCalls(*mockDB)
	db := db.NewPgxMockDB(serverConfig, mock)

	eventBroker := memory.NewBroker()
	brokerContext, cancel := context.WithCancel(context.Background())
	defer cancel()
	auditService := service.NewAuditService(&serverConfig)
	auditService.SubscribeToMetricEvents(eventBroker, brokerContext)

	metricsService := service.NewMetricsService(&serverConfig, &db, eventBroker)
	handlers := InitializeHandlers(metricsService, &cryptoService, nil)

	return CreateChiRouter(&handlers)
}
