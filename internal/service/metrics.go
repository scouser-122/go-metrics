package service

import (
	"fmt"
	"strconv"

	models "github.com/scouser-122/go-metrics/internal/model"
	"github.com/scouser-122/go-metrics/internal/repository"
)

type MetricsService struct {
	Storage repository.MetricsStorage
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
