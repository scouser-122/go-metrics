package repository

import (
	"context"
	"fmt"
	"slices"

	models "github.com/scouser-122/go-metrics/internal/model"
)

type MemStorage struct {
	Metrics []models.Metrics
}

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

func (memStorage *MemStorage) GetAllMetrics(ctx context.Context) []models.Metrics {
	return memStorage.Metrics
}

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

func (memStorage *MemStorage) GetMetricWithValue(ctx context.Context, metric *models.Metrics) (*models.Metrics, error) {
	index := slices.IndexFunc(memStorage.Metrics, func(m models.Metrics) bool {
		return m.MType == metric.MType && m.ID == metric.ID
	})
	if index != -1 {
		return &memStorage.Metrics[index], nil
	}
	return nil, models.MetricGetError{Message: fmt.Sprintf("unknown metric %s", metric.ID)}
}
