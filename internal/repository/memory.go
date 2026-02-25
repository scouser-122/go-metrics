package repository

import (
	"fmt"
	"slices"
	"strconv"

	models "github.com/scouser-122/go-metrics/internal/model"
)

type MemStorage struct {
	Metrics []models.Metrics
}

func (memStorage *MemStorage) SaveCounter(name string, value int64) {
	index := slices.IndexFunc(memStorage.Metrics, func(m models.Metrics) bool {
		return m.MType == models.Counter && m.ID == name
	})
	if index != -1 {
		memStorage.Metrics[index].Delta = &value
	} else {
		metric := models.Metrics{
			ID:    name,
			MType: models.Counter,
			Delta: &value,
		}
		memStorage.Metrics = append(memStorage.Metrics, metric)
	}
}

func (memStorage *MemStorage) SaveGauge(name string, value float64) {
	index := slices.IndexFunc(memStorage.Metrics, func(m models.Metrics) bool {
		return m.MType == models.Gauge && m.ID == name
	})
	if index != -1 {
		memStorage.Metrics[index].Value = &value
	} else {
		metric := models.Metrics{
			ID:    name,
			MType: models.Gauge,
			Value: &value,
		}
		memStorage.Metrics = append(memStorage.Metrics, metric)
	}
}

func (memStorage *MemStorage) Print() {
	fmt.Printf("Metrics: [\n")
	for i, v := range memStorage.Metrics {
		var delta string = "<nil>"
		if v.Delta != nil {
			delta = strconv.FormatInt(*v.Delta, 10)
		}
		var value string = "<nil>"
		if v.Value != nil {
			value = strconv.FormatFloat(*v.Value, 'f', 2, 64)
		}
		fmt.Printf("\t%d: %q %q %q %q %q\n", i, v.MType, v.ID, delta, value, v.Hash)
	}
	fmt.Printf("]\n")
}

var MemoryStorage = MemStorage{}
