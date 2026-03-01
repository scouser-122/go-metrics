package repository

import (
	"slices"

	models "github.com/scouser-122/go-metrics/internal/model"
)

type MemStorage struct {
	Metrics []models.Metrics
}

func (memStorage *MemStorage) SaveCounter(name string, value int64) (bool, error) {
	index := slices.IndexFunc(memStorage.Metrics, func(m models.Metrics) bool {
		return m.MType == models.Counter && m.ID == name
	})
	if index != -1 {
		*memStorage.Metrics[index].Delta += value
	} else {
		metric := models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: new(int64),
		}
		*metric.Delta = value
		memStorage.Metrics = append(memStorage.Metrics, metric)
	}
	return true, nil
}

func (memStorage *MemStorage) SaveGauge(name string, value float64) (bool, error) {
	index := slices.IndexFunc(memStorage.Metrics, func(m models.Metrics) bool {
		return m.MType == models.Gauge && m.ID == name
	})
	if index != -1 {
		*memStorage.Metrics[index].Value = value
	} else {
		metric := models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: new(float64),
		}
		*metric.Value = value
		memStorage.Metrics = append(memStorage.Metrics, metric)
	}
	return true, nil
}
