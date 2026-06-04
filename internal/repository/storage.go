package repository

import (
	"context"

	models "github.com/scouser-122/go-metrics/internal/model"
)

type MetricsStorage interface {
	UpdateOrCreateCounter(context.Context, string, int64) (*models.Metrics, error)
	UpdateOrCreateGauge(context.Context, string, float64) (*models.Metrics, error)
	UpdateOrCreateMetric(context.Context, models.Metrics) (models.Metrics, error)
	UpdateOrCreateMetrics(context.Context, []models.Metrics) (int64, error)
	GetAllMetrics(context.Context) []models.Metrics
	SaveMetrics(context.Context, []models.Metrics) error
	GetGauge(context.Context, string) (float64, error)
	GetCounter(context.Context, string) (int64, error)
	GetMetricWithValue(context.Context, *models.Metrics) (*models.Metrics, error)
}
