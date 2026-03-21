package repository

import models "github.com/scouser-122/go-metrics/internal/model"

type MetricsStorage interface {
	UpdateOrCreateCounter(string, int64) (int64, error)
	UpdateOrCreateGauge(string, float64) (float64, error)
	UpdateOrCreateMetric(models.Metrics) (models.Metrics, error)
	GetAllMetrics() []models.Metrics
	SaveMetrics([]models.Metrics) error
	GetGauge(string) (float64, error)
	GetCounter(string) (int64, error)
	GetMetricWithValue(*models.Metrics) (*models.Metrics, error)
}
