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

func (service *MetricsService) SaveMetric(metricType string, name string, value string) (bool, error) {
	if metricType != models.Counter && metricType != models.Gauge {
		return false, models.IncorrectMetricType{
			Message: fmt.Sprintf("Metric type incorrect: %q", metricType),
		}
	}

	var saveResult bool
	switch metricType {
	case models.Counter:
		counterValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return false, models.MetricFormatError{
				Message: fmt.Sprintf("Metric counter incorrect format: %v\n", err),
			}
		}
		saveResult, err = service.Storage.SaveCounter(name, counterValue)
		if err != nil {
			return false, models.MetricSaveError{
				Message: fmt.Sprintf("Metric counter save failed: %v\n", err),
			}
		}

	case models.Gauge:
		gaugeValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return false, models.MetricFormatError{
				Message: fmt.Sprintf("Metric gauge incorrect format: %v\n", err),
			}
		}
		saveResult, err = service.Storage.SaveGauge(name, gaugeValue)
		if err != nil {
			return false, models.MetricSaveError{
				Message: fmt.Sprintf("Metric counter save failed: %v\n", err),
			}
		}
	}

	return saveResult, nil
}
