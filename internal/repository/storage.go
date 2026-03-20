package repository

import models "github.com/scouser-122/go-metrics/internal/model"

type MetricsStorage interface {
	SaveCounter(string, int64) (int64, error)
	SaveGauge(string, float64) (float64, error)
	SaveMetric(models.Metrics) (models.Metrics, error)
	GetAllMetrics() []models.Metrics
	SaveMetrics([]models.Metrics) error
	GetGauge(string) (float64, error)
	GetCounter(string) (int64, error)
	GetMetricWithValue(*models.Metrics) (*models.Metrics, error)
}
