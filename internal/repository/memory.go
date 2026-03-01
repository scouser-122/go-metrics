package repository

import (
	"slices"

	"github.com/scouser-122/go-metrics/internal/config"
	models "github.com/scouser-122/go-metrics/internal/model"
)

type MemStorage struct {
	Metrics []models.Metrics
}

func (memStorage *MemStorage) FillMetrics(config *config.ServerConfig) {
	for _, name := range config.GaugeMetricNames {
		metric := models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: new(float64),
		}
		memStorage.Metrics = append(memStorage.Metrics, metric)
	}
	for _, name := range config.CounterMetricNames {
		metric := models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: new(int64),
		}
		memStorage.Metrics = append(memStorage.Metrics, metric)
	}
}

func (memStorage *MemStorage) SaveCounter(name string, value int64) (int64, error) {
	index := slices.IndexFunc(memStorage.Metrics, func(m models.Metrics) bool {
		return m.MType == models.Counter && m.ID == name
	})
	if index != -1 {
		*memStorage.Metrics[index].Delta += value
		return *memStorage.Metrics[index].Delta, nil
	} else {
		return 0, models.MetricSaveError{Message: "Unknown metric"}
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
		return 0, models.MetricSaveError{Message: "Unknown metric"}
	}
}

func (memStorage *MemStorage) GetAllMetrics() []models.Metrics {
	return memStorage.Metrics
}
