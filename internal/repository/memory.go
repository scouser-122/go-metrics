package repository

import (
	"context"
	"fmt"
	"slices"

	models "github.com/scouser-122/go-metrics/internal/model"
)

// MemStorage provides an in-memory implementation of metrics storage.
type MemStorage struct {
	Metrics []models.Metrics
}

// UpdateOrCreateCounter updates an existing counter or creates a new one.
func (memStorage *MemStorage) UpdateOrCreateCounter(ctx context.Context, name string, value int64) (*models.Metrics, error) {
	index := slices.IndexFunc(memStorage.Metrics, func(m models.Metrics) bool {
		return m.MType == models.Counter && m.ID == name
	})
	if index != -1 {
		*memStorage.Metrics[index].Delta += value
		return &memStorage.Metrics[index], nil
	} else {
		metric := models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: new(int64),
		}
		*metric.Delta = value
		memStorage.Metrics = append(memStorage.Metrics, metric)
		return &metric, nil
	}
}

// UpdateOrCreateGauge updates an existing gauge or creates a new one.
func (memStorage *MemStorage) UpdateOrCreateGauge(ctx context.Context, name string, value float64) (*models.Metrics, error) {
	index := slices.IndexFunc(memStorage.Metrics, func(m models.Metrics) bool {
		return m.MType == models.Gauge && m.ID == name
	})
	if index != -1 {
		*memStorage.Metrics[index].Value = value
		return &memStorage.Metrics[index], nil
	} else {
		metric := models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: new(float64),
		}
		*metric.Value = value
		memStorage.Metrics = append(memStorage.Metrics, metric)
		return &metric, nil
	}
}

// UpdateOrCreateMetric updates an existing metric or creates a new one.
func (memStorage *MemStorage) UpdateOrCreateMetric(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
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
		metricCopy := metric
		if metric.Delta != nil {
			metricCopy.Delta = new(int64)
			*metricCopy.Delta = *metric.Delta
		}
		if metric.Value != nil {
			metricCopy.Value = new(float64)
			*metricCopy.Value = *metric.Value
		}
		memStorage.Metrics = append(memStorage.Metrics, metricCopy)
	}
	return metric, nil
}

// UpdateOrCreateMetrics updates an existing metrics or creates new metrics.
func (memStorage *MemStorage) UpdateOrCreateMetrics(ctx context.Context, metrics []models.Metrics) (int64, error) {
	count := int64(0)
	for _, m := range metrics {
		_, err := memStorage.UpdateOrCreateMetric(ctx, m)
		if err == nil {
			count++
		}
	}
	return count, nil
}

// GetAllMetrics returns all stored metrics.
func (memStorage *MemStorage) GetAllMetrics(ctx context.Context) ([]models.Metrics, error) {
	return memStorage.Metrics, nil
}

// SaveMetrics store passed metrics.
func (memStorage *MemStorage) SaveMetrics(ctx context.Context, metrics []models.Metrics) error {
	for _, metric := range metrics {
		index := slices.IndexFunc(memStorage.Metrics, func(m models.Metrics) bool {
			return m.MType == metric.MType && m.ID == metric.ID
		})
		if index != -1 {
			foundMetric := memStorage.Metrics[index]
			switch metric.MType {
			case models.Counter:
				*foundMetric.Delta = *metric.Delta
			case models.Gauge:
				*foundMetric.Value = *metric.Value
			}
		} else {
			metricCopy := metric
			if metric.Delta != nil {
				metricCopy.Delta = new(int64)
				*metricCopy.Delta = *metric.Delta
			}
			if metric.Value != nil {
				metricCopy.Value = new(float64)
				*metricCopy.Value = *metric.Value
			}
			memStorage.Metrics = append(memStorage.Metrics, metricCopy)
		}
	}
	return nil
}

// GetCounter returns counter by name if exists. If not - returns error.
func (memStorage *MemStorage) GetCounter(ctx context.Context, name string) (int64, error) {
	index := slices.IndexFunc(memStorage.Metrics, func(m models.Metrics) bool {
		return m.MType == models.Counter && m.ID == name
	})
	if index != -1 {
		return *memStorage.Metrics[index].Delta, nil
	} else {
		return 0, models.MetricGetError{Message: fmt.Sprintf("unknown metric %s", name)}
	}
}

// GetGauge returns gauge by name if exists. If not - returns error.
func (memStorage *MemStorage) GetGauge(ctx context.Context, name string) (float64, error) {
	index := slices.IndexFunc(memStorage.Metrics, func(m models.Metrics) bool {
		return m.MType == models.Gauge && m.ID == name
	})
	if index != -1 {
		return *memStorage.Metrics[index].Value, nil
	} else {
		return 0, models.MetricGetError{Message: fmt.Sprintf("unknown metric %s", name)}
	}
}

// GetMetricWithValue returns metric by specified parameters if exists. If not - returns error.
func (memStorage *MemStorage) GetMetricWithValue(ctx context.Context, metric *models.Metrics) (*models.Metrics, error) {
	index := slices.IndexFunc(memStorage.Metrics, func(m models.Metrics) bool {
		return m.MType == metric.MType && m.ID == metric.ID
	})
	if index != -1 {
		return &memStorage.Metrics[index], nil
	}
	return nil, models.MetricGetError{Message: fmt.Sprintf("unknown metric %s", metric.ID)}
}
