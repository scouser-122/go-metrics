package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/botchris/go-pubsub"
	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/logger"
	models "github.com/scouser-122/go-metrics/internal/model"
	"github.com/scouser-122/go-metrics/internal/repository"
	"github.com/scouser-122/go-metrics/internal/repository/db"
)

type MetricsService struct {
	Storage      repository.MetricsStorage
	serverConfig *config.ServerConfig
	fsStorage    repository.MetricsStorage
	eventBroker  pubsub.Broker
}

// NewMetricsService creates new MetricsService instance
func NewMetricsService(
	serverConfig *config.ServerConfig,
	db *db.PostgresDatabase,
	eventBroker pubsub.Broker,
) *MetricsService {
	service := MetricsService{
		serverConfig: serverConfig,
		eventBroker:  eventBroker,
	}
	service.createStorage(db)
	return &service
}

func (service *MetricsService) createStorage(db *db.PostgresDatabase) {
	if err := db.Ping(context.Background()); err == nil {
		logger.Sugar.Infof("use database storage")
		service.Storage = &repository.PostgresDBStorage{
			Database: db,
		}
		if service.serverConfig.StorePath != "" {
			logger.Sugar.Infof("additionaly use filesystem storage")
			service.fsStorage = repository.CreateFileSystemStorage(service.serverConfig)
			service.restoreMetricsIfRequired()
			service.storeMetricsInFsIfRequired()
		}
		return
	}
	if service.serverConfig.StorePath != "" {
		logger.Sugar.Infof("use filesystem storage")
		service.fsStorage = repository.CreateFileSystemStorage(service.serverConfig)
		service.Storage = service.fsStorage
		service.restoreMetricsIfRequired()
		service.storeMetricsInFsIfRequired()
		return
	}
	logger.Sugar.Infof("use in-memory storage")
	service.Storage = &repository.MemStorage{}
}

func (service *MetricsService) SaveMetric(ctx context.Context, metricType string, name string, value string) (string, error) {
	var result string
	if metricType != models.Counter && metricType != models.Gauge {
		return result, models.IncorrectMetricType{
			Message: fmt.Sprintf("metric type incorrect: %q", metricType),
		}
	}

	switch metricType {
	case models.Counter:
		counterValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return result, models.MetricFormatError{
				Message: "metric counter incorrect format",
				Err:     err,
			}
		}
		metric, err := service.Storage.UpdateOrCreateCounter(ctx, name, counterValue)
		if err != nil {
			return result, err
		}
		service.logMetricsReceiveEventToAudit(ctx, []models.Metrics{*metric})
		result = strconv.FormatInt(*metric.Delta, 10)

	case models.Gauge:
		gaugeValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return result, models.MetricFormatError{
				Message: "metric gauge incorrect format",
				Err:     err,
			}
		}
		metric, err := service.Storage.UpdateOrCreateGauge(ctx, name, gaugeValue)
		if err != nil {
			return result, err
		}
		service.logMetricsReceiveEventToAudit(ctx, []models.Metrics{*metric})
		result = strconv.FormatFloat(*metric.Value, 'f', -1, 64)
	}

	return result, nil
}

func (service *MetricsService) SaveMetricModel(ctx context.Context, metric *models.Metrics) (models.Metrics, error) {
	var result models.Metrics
	switch metric.MType {
	case models.Counter:
		if metric.Delta == nil {
			return result, models.MetricFormatError{
				Message: "metric counter missing delta",
			}
		}
	case models.Gauge:
		if metric.Value == nil {
			return result, models.MetricFormatError{
				Message: "metric gauge missing value",
			}
		}
	default:
		return result, models.IncorrectMetricType{
			Message: fmt.Sprintf("metric type incorrect: %q", metric.MType),
		}
	}
	result, err := service.Storage.UpdateOrCreateMetric(ctx, *metric)
	service.logMetricsReceiveEventToAudit(ctx, []models.Metrics{result})
	return result, err
}

func (service *MetricsService) SaveMetricsModel(ctx context.Context, metrics []models.Metrics) (int64, error) {
	var result int64
	for _, m := range metrics {
		switch m.MType {
		case models.Counter:
			if m.Delta == nil {
				return 0, models.MetricFormatError{
					Message: fmt.Sprintf("metric counter %s missing delta", m.ID),
				}
			}
		case models.Gauge:
			if m.Value == nil {
				return 0, models.MetricFormatError{
					Message: fmt.Sprintf("metric gauge %s missing value", m.ID),
				}
			}
		default:
			return 0, models.IncorrectMetricType{
				Message: fmt.Sprintf("metric type incorrect: %s %q", m.ID, m.MType),
			}
		}
	}
	var err error
	result, err = service.Storage.UpdateOrCreateMetrics(ctx, metrics)
	service.logMetricsReceiveEventToAudit(ctx, metrics)
	return result, err
}

func (service *MetricsService) GetAllMetrics(ctx context.Context) []models.Metrics {
	return service.Storage.GetAllMetrics(ctx)
}

func (service *MetricsService) GetValue(ctx context.Context, metricType string, name string) (string, error) {
	var result string
	if metricType != models.Counter && metricType != models.Gauge {
		return result, models.IncorrectMetricType{
			Message: fmt.Sprintf("metric type incorrect: %q", metricType),
		}
	}

	switch metricType {
	case models.Counter:
		getResult, err := service.Storage.GetCounter(ctx, name)
		if err != nil {
			return result, err
		}
		result = strconv.FormatInt(getResult, 10)

	case models.Gauge:
		getResult, err := service.Storage.GetGauge(ctx, name)
		if err != nil {
			return result, err
		}
		result = strconv.FormatFloat(getResult, 'f', -1, 64)
	}

	return result, nil
}

func (service *MetricsService) ReadMetric(ctx context.Context, metric *models.Metrics) (*models.Metrics, error) {
	if metric.MType != models.Counter && metric.MType != models.Gauge {
		return nil, models.IncorrectMetricType{
			Message: fmt.Sprintf("metric type incorrect: %q", metric.MType),
		}
	}
	result, err := service.Storage.GetMetricWithValue(ctx, metric)
	return result, err
}

func (service *MetricsService) restoreMetricsIfRequired() {
	if !service.serverConfig.Restore {
		return
	}
	metrics := service.fsStorage.GetAllMetrics(context.Background())
	service.Storage.SaveMetrics(context.Background(), metrics)
}

func (service *MetricsService) storeMetricsInFsIfRequired() {
	if service.serverConfig.StoreInterval == 0 || service.serverConfig.StoreInterval == -1 {
		return
	}
	logger.Sugar.Infof("start storing metrics in time interval %d seconds", service.serverConfig.StoreInterval)
	go service.storeMetricsInFsWorker()
}

func (service *MetricsService) storeMetricsInFsWorker() {
	ticker := time.NewTicker(time.Duration(service.serverConfig.StoreInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		service.fsStorage.SaveMetrics(ctx, service.Storage.GetAllMetrics(ctx))
	}
}

func (service *MetricsService) logMetricsReceiveEventToAudit(ctx context.Context, metrics []models.Metrics) {
	if service.eventBroker == nil {
		return
	}
	event := models.MetricsReceivedEvent{}
	event.TS = time.Now().UnixMilli()
	if ipAddress, ok := ctx.Value(models.IPAddressContextKey).(string); ok {
		event.IPAddress = ipAddress
	}
	for _, m := range metrics {
		event.Metrics = append(event.Metrics, m.ID)
	}
	service.eventBroker.Publish(ctx, models.MetricEventTopic, event)
}
