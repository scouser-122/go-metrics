package handler

import (
	"context"
	"fmt"

	"github.com/botchris/go-pubsub/provider/memory"
	"github.com/go-chi/chi/v5"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/logger"
	models "github.com/scouser-122/go-metrics/internal/model"
	"github.com/scouser-122/go-metrics/internal/repository/db"
	"github.com/scouser-122/go-metrics/internal/service"
)

func createTestRouterPostgresDB(mockDB *db.MockPostgresDBTestData) *chi.Mux {
	serverConfig := config.DefaultServerConfig()
	return createTestRouterPostgresDBServerConfig(mockDB, &serverConfig)
}

func createTestRouterPostgresDBServerConfig(
	mockDB *db.MockPostgresDBTestData,
	serverConfig *config.ServerConfig,
) *chi.Mux {
	if err := logger.Initialize(serverConfig.LogLevel, serverConfig.Environment); err != nil {
		panic(err)
	}
	cryptoService := service.CryptoService{
		ServerConfig: serverConfig,
	}
	cryptoService.LoadPrivateKeyIfExists()
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
	auditService := service.NewAuditService(serverConfig)
	auditService.SubscribeToMetricEvents(eventBroker, brokerContext)

	metricsService := service.NewMetricsService(serverConfig, &db, eventBroker)
	handlers := InitializeHandlers(metricsService, &cryptoService, nil)

	return CreateChiRouterWithHandlers(&handlers, serverConfig)
}

func Ptr[T any](v T) *T {
	return &v
}

func generateTestMetrics(size int) []models.Metrics {
	metrics := make([]models.Metrics, 0, size)
	for i := 0; i < size; i += 2 {
		metrics = append(metrics, models.Metrics{
			ID:    fmt.Sprintf("TestCounter_%d", i),
			MType: models.Counter,
			Delta: Ptr(int64(i * 10)),
		})
		metrics = append(metrics, models.Metrics{
			ID:    fmt.Sprintf("TestGauge_%d", i+1),
			MType: models.Gauge,
			Value: Ptr(float64((i + 1) * 10)),
		})
	}
	return metrics
}
