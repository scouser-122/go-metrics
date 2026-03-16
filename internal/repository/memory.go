package repository

import (
	"fmt"
	"slices"

	models "github.com/scouser-122/go-metrics/internal/model"
)

type MemStorage struct {
	Metrics []models.Metrics
}

func (memStorage *MemStorage) SaveCounter(name string, value int64) (int64, error) {
	index := slices.IndexFunc(memStorage.Metrics, func(m models.Metrics) bool {
		return m.MType == models.Counter && m.ID == name
	})
	if index != -1 {
		*memStorage.Metrics[index].Delta += value
		return *memStorage.Metrics[index].Delta, nil
	} else {
		metric := models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: new(int64),
		}
		*metric.Delta = value
		memStorage.Metrics = append(memStorage.Metrics, metric)
		return value, nil
	}
}

func (memStorage *MemStorage) SaveGauge(name string, value float64) (float64, error) {
	index := slices.IndexFunc(memStorage.Metrics, func(m models.Metrics) bool {
		return m.MType == models.Gauge && m.ID == name
	})
	if index != -1 {
		*memStorage.Metrics[index].Value = value
		return *memStorage.Metrics[index].Value, nil
	} else {
		metric := models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: new(float64),
		}
		*metric.Value = value
		memStorage.Metrics = append(memStorage.Metrics, metric)
		return value, nil
	}
}

func (memStorage *MemStorage) SaveMetric(metric models.Metrics) (models.Metrics, error) {
	index := slices.IndexFunc(memStorage.Metrics, func(m models.Metrics) bool {
		return m.MType == metric.MType && m.ID == metric.ID
	})
	if index != -1 {
		foundMetric := memStorage.Metrics[index]
		switch metric.MType {
		case models.Counter:
			*foundMetric.Delta += int64(*metric.Delta)
			return foundMetric, nil
		case models.Gauge:
			*foundMetric.Value = float64(*metric.Value)
			return foundMetric, nil
		}
	} else {
		memStorage.Metrics = append(memStorage.Metrics, metric)
	}
	return metric, nil
}

func (memStorage *MemStorage) GetAllMetrics() []models.Metrics {
	return memStorage.Metrics
}

func (memStorage *MemStorage) GetCounter(name string) (int64, error) {
	index := slices.IndexFunc(memStorage.Metrics, func(m models.Metrics) bool {
		return m.MType == models.Counter && m.ID == name
	})
	if index != -1 {
		return *memStorage.Metrics[index].Delta, nil
	} else {
		return 0, models.MetricGetError{Message: fmt.Sprintf("unknown metric %s", name)}
	}
}

func (memStorage *MemStorage) GetGauge(name string) (float64, error) {
	index := slices.IndexFunc(memStorage.Metrics, func(m models.Metrics) bool {
		return m.MType == models.Gauge && m.ID == name
	})
	if index != -1 {
		return *memStorage.Metrics[index].Value, nil
	} else {
		return 0, models.MetricGetError{Message: fmt.Sprintf("unknown metric %s", name)}
	}
}

func (memStorage *MemStorage) GetMetricWithValue(metric *models.Metrics) (*models.Metrics, error) {
	index := slices.IndexFunc(memStorage.Metrics, func(m models.Metrics) bool {
		return m.MType == metric.MType && m.ID == metric.ID
	})
	if index != -1 {
		return &memStorage.Metrics[index], nil
	}
	return nil, models.MetricGetError{Message: fmt.Sprintf("unknown metric %s", metric.ID)}
}
