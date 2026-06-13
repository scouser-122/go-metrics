package repository

import (
	"context"

	models "github.com/scouser-122/go-metrics/internal/model"
)

// MetricsStorage defines the interface for metrics storage operations.
type MetricsStorage interface {
	// UpdateOrCreateCounter updates an existing counter or creates a new one.
	UpdateOrCreateCounter(context.Context, string, int64) (*models.Metrics, error)

	// UpdateOrCreateGauge updates an existing gauge or creates a new one.
	UpdateOrCreateGauge(context.Context, string, float64) (*models.Metrics, error)

	// UpdateOrCreateMetric updates an existing metric or creates a new one.
	UpdateOrCreateMetric(context.Context, models.Metrics) (models.Metrics, error)

	// UpdateOrCreateMetrics updates an existing metrics or creates new metrics.
	UpdateOrCreateMetrics(context.Context, []models.Metrics) (int64, error)

	// GetAllMetrics returns all stored metrics.
	GetAllMetrics(context.Context) []models.Metrics

	// SaveMetrics store passed metrics.
	SaveMetrics(context.Context, []models.Metrics) error

	// GetCounter returns counter by name if exists. If not - returns error.
	GetCounter(context.Context, string) (int64, error)

	// GetGauge returns gauge by name if exists. If not - returns error.
	GetGauge(context.Context, string) (float64, error)

	// GetMetricWithValue returns metric by specified parameters if exists. If not - returns error.
	GetMetricWithValue(context.Context, *models.Metrics) (*models.Metrics, error)
}
