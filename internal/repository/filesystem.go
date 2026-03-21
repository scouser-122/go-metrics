package repository

import (
	"encoding/json"
	"os"

	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/logger"
	models "github.com/scouser-122/go-metrics/internal/model"
)

type FileSystemStorage struct {
	MemoryStorage MemStorage
	config        *config.ServerConfig
}

func CreateFileSystemStorage(config *config.ServerConfig) *FileSystemStorage {
	memStorage := MemStorage{}
	return &FileSystemStorage{
		MemoryStorage: memStorage,
		config:        config,
	}
}

func (storage *FileSystemStorage) UpdateOrCreateCounter(name string, value int64) (int64, error) {
	value, err := storage.MemoryStorage.UpdateOrCreateCounter(name, value)
	if err != nil {
		return value, err
	}
	err = storage.saveMetricsInFS()
	return value, err
}

func (storage *FileSystemStorage) UpdateOrCreateGauge(name string, value float64) (float64, error) {
	value, err := storage.MemoryStorage.UpdateOrCreateGauge(name, value)
	if err != nil {
		return value, err
	}
	err = storage.saveMetricsInFS()
	return value, err
}

func (storage *FileSystemStorage) UpdateOrCreateMetric(metric models.Metrics) (models.Metrics, error) {
	metric, err := storage.MemoryStorage.UpdateOrCreateMetric(metric)
	if err != nil {
		return metric, err
	}
	err = storage.saveMetricsInFS()
	return metric, err
}

func (storage *FileSystemStorage) GetAllMetrics() []models.Metrics {
	if len(storage.MemoryStorage.Metrics) == 0 {
		storage.loadMetricsFromFS()
	}
	return storage.MemoryStorage.Metrics
}

func (storage *FileSystemStorage) SaveMetrics(metrics []models.Metrics) error {
	err := storage.MemoryStorage.SaveMetrics(metrics)
	if err != nil {
		return err
	}
	storage.saveMetricsInFS()
	return nil
}

func (storage *FileSystemStorage) GetCounter(name string) (int64, error) {
	if len(storage.MemoryStorage.Metrics) == 0 {
		storage.loadMetricsFromFS()
	}
	return storage.MemoryStorage.GetCounter(name)
}

func (storage *FileSystemStorage) GetGauge(name string) (float64, error) {
	if len(storage.MemoryStorage.Metrics) == 0 {
		storage.loadMetricsFromFS()
	}
	return storage.MemoryStorage.GetGauge(name)
}

func (storage *FileSystemStorage) GetMetricWithValue(metric *models.Metrics) (*models.Metrics, error) {
	if len(storage.MemoryStorage.Metrics) == 0 {
		storage.loadMetricsFromFS()
	}
	return storage.MemoryStorage.GetMetricWithValue(metric)
}

func (storage *FileSystemStorage) saveMetricsInFS() error {
	metrics := storage.MemoryStorage.Metrics
	if len(metrics) == 0 {
		return nil
	}
	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		logger.Sugar.Errorf("couldn't parse metrics to save in FS: %q", err)
		return err
	}
	err = os.WriteFile(storage.config.StorePath, data, 0666)
	if err != nil {
		logger.Sugar.Errorf("couldn't save metrics in FS: %q", err)
	}
	logger.Sugar.Infof("succesfully saved %d metrics to file %s", len(metrics), storage.config.StorePath)
	return nil
}

func (storage *FileSystemStorage) loadMetricsFromFS() {
	result := []models.Metrics{}
	data, err := os.ReadFile(storage.config.StorePath)
	if err != nil {
		logger.Sugar.Errorf("couldn't load metrics from FS: %q", err)
		return
	}
	if err := json.Unmarshal(data, &result); err != nil {
		logger.Sugar.Errorf("couldn't parse metrics loaded from FS: %q", err)
		return
	}
	if len(result) > 0 {
		logger.Sugar.Infof("succesfully loaded %d metrics from file %s", len(result), storage.config.StorePath)
		storage.MemoryStorage.Metrics = result
	}
}
