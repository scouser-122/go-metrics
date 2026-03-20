package service

import (
	"fmt"
	"strconv"
	"time"

	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/logger"
	models "github.com/scouser-122/go-metrics/internal/model"
	"github.com/scouser-122/go-metrics/internal/repository"
)

type MetricsService struct {
	Storage      repository.MetricsStorage
	serverConfig *config.ServerConfig
	fsStorage    *repository.FileSystemStorage
}

func (service *MetricsService) Initialize(config *config.ServerConfig) {
	service.serverConfig = config
	service.fsStorage = repository.CreateFileSystemStorage(config)
	service.restoreMetricsIfRequired()
	service.storeMetricsInFsIfRequired()
}

func (service *MetricsService) SaveMetric(metricType string, name string, value string) (string, error) {
	var result string
	if metricType != models.Counter && metricType != models.Gauge {
		return result, models.IncorrectMetricType{
			Message: fmt.Sprintf("Metric type incorrect: %q", metricType),
		}
	}

	switch metricType {
	case models.Counter:
		counterValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return result, models.MetricFormatError{
				Message: fmt.Sprintf("Metric counter incorrect format: %v\n", err),
			}
		}
		saveResult, err := service.Storage.SaveCounter(name, counterValue)
		if err != nil {
			return result, err
		}
		if service.serverConfig.StoreInterval == 0 {
			if _, err := service.fsStorage.SaveCounter(name, counterValue); err != nil {
				return result, err
			}
		}
		result = strconv.FormatInt(saveResult, 10)

	case models.Gauge:
		gaugeValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return result, models.MetricFormatError{
				Message: fmt.Sprintf("Metric gauge incorrect format: %v\n", err),
			}
		}
		saveResult, err := service.Storage.SaveGauge(name, gaugeValue)
		if err != nil {
			return result, err
		}
		if service.serverConfig.StoreInterval == 0 {
			if _, err := service.fsStorage.SaveGauge(name, gaugeValue); err != nil {
				return result, err
			}
		}
		result = strconv.FormatFloat(saveResult, 'f', -1, 64)
	}

	return result, nil
}

func (service *MetricsService) SaveMetricModel(metric *models.Metrics) (models.Metrics, error) {
	var result models.Metrics
	switch metric.MType {
	case models.Counter:
		if metric.Delta == nil {
			return result, models.MetricFormatError{
				Message: "Metric counter missing delta",
			}
		}
	case models.Gauge:
		if metric.Value == nil {
			return result, models.MetricFormatError{
				Message: "Metric gauge missing value",
			}
		}
	default:
		return result, models.IncorrectMetricType{
			Message: fmt.Sprintf("Metric type incorrect: %q", metric.MType),
		}
	}
	result, err := service.Storage.SaveMetric(*metric)
	if service.serverConfig.StoreInterval == 0 {
		if result, err := service.fsStorage.SaveMetric(*metric); err != nil {
			return result, err
		}
	}
	return result, err
}

func (service *MetricsService) GetAllMetrics() []models.Metrics {
	return service.Storage.GetAllMetrics()
}

func (service *MetricsService) GetValue(metricType string, name string) (string, error) {
	var result string
	if metricType != models.Counter && metricType != models.Gauge {
		return result, models.IncorrectMetricType{
			Message: fmt.Sprintf("Metric type incorrect: %q", metricType),
		}
	}

	switch metricType {
	case models.Counter:
		getResult, err := service.Storage.GetCounter(name)
		if err != nil {
			return result, err
		}
		result = strconv.FormatInt(getResult, 10)

	case models.Gauge:
		getResult, err := service.Storage.GetGauge(name)
		if err != nil {
			return result, err
		}
		result = strconv.FormatFloat(getResult, 'f', -1, 64)
	}

	return result, nil
}

func (service *MetricsService) ReadMetric(metric *models.Metrics) (*models.Metrics, error) {
	if metric.MType != models.Counter && metric.MType != models.Gauge {
		return nil, models.IncorrectMetricType{
			Message: fmt.Sprintf("Metric type incorrect: %q", metric.MType),
		}
	}
	result, err := service.Storage.GetMetricWithValue(metric)
	return result, err
}

func (service *MetricsService) restoreMetricsIfRequired() {
	if !service.serverConfig.Restore {
		return
	}
	metrics := service.fsStorage.GetAllMetrics()
	service.Storage.SaveMetrics(metrics)
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
		service.fsStorage.SaveMetrics(service.Storage.GetAllMetrics())
	}
}
