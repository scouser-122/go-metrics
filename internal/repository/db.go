package repository

import (
	"context"
	"errors"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/scouser-122/go-metrics/internal/config/db"
	"github.com/scouser-122/go-metrics/internal/logger"
	models "github.com/scouser-122/go-metrics/internal/model"
)

type DataBaseStorage struct {
	Database *db.Database
}

func (storage *DataBaseStorage) UpdateOrCreateCounter(ctx context.Context, name string, value int64) (int64, error) {
	row, err := storage.Database.QueryRow(
		ctx,
		"SELECT id, type, delta FROM metrics WHERE id = $1 AND type = $2",
		name, models.Counter,
	)
	if err != nil {
		return value, models.MetricSaveError{Message: err.Error()}
	}

	metric := models.Metrics{
		Delta: new(int64),
	}
	err = row.Scan(&metric.ID, &metric.MType, &metric.Delta)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			metric := models.Metrics{
				ID:    name,
				MType: models.Counter,
				Delta: new(int64),
			}
			*metric.Delta = value
			_, err = storage.Database.Exec(
				ctx,
				"INSERT INTO metrics (id,type,delta) VALUES ($1,$2,$3)",
				metric.ID, metric.MType, metric.Delta,
			)
			if err != nil {
				return value, models.MetricSaveError{Message: err.Error()}
			}
			return value, nil
		} else {
			return value, models.MetricSaveError{Message: err.Error()}
		}
	}

	*metric.Delta += value

	_, err = storage.Database.Exec(
		ctx,
		"UPDATE metrics SET delta = $1 WHERE id = $2 AND type = $3",
		*metric.Delta, metric.ID, metric.MType,
	)
	if err != nil {
		return value, models.MetricSaveError{Message: err.Error()}
	}

	return *metric.Delta, nil
}

func (storage *DataBaseStorage) UpdateOrCreateGauge(ctx context.Context, name string, value float64) (float64, error) {
	row, err := storage.Database.QueryRow(
		ctx,
		"SELECT id, type FROM metrics WHERE id = $1 AND type = $2",
		name, models.Gauge,
	)
	if err != nil {
		return value, models.MetricSaveError{Message: err.Error()}
	}

	metric := models.Metrics{}
	err = row.Scan(&metric.ID, &metric.MType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			metric := models.Metrics{
				ID:    name,
				MType: models.Gauge,
				Value: new(float64),
			}
			*metric.Value = value
			_, err = storage.Database.Exec(
				ctx,
				"INSERT INTO metrics (id,type,value) VALUES ($1,$2,$3)",
				metric.ID, metric.MType, metric.Value,
			)
			if err != nil {
				return value, models.MetricSaveError{Message: err.Error()}
			}
			return value, nil
		} else {
			return value, models.MetricSaveError{Message: err.Error()}
		}
	}

	_, err = storage.Database.Exec(
		ctx,
		"UPDATE metrics SET value = $1 WHERE id = $2 AND type = $3",
		value, metric.ID, metric.MType,
	)
	if err != nil {
		return value, models.MetricSaveError{Message: err.Error()}
	}

	return value, nil
}

func (storage *DataBaseStorage) UpdateOrCreateMetric(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	row, err := storage.Database.QueryRow(
		ctx,
		"SELECT delta, value FROM metrics WHERE id = $1 AND type = $2",
		metric.ID, metric.MType,
	)
	if err != nil {
		return metric, models.MetricSaveError{Message: err.Error()}
	}

	dbMetric := models.Metrics{}
	err = row.Scan(&dbMetric.Delta, &dbMetric.Value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_, err = storage.Database.Exec(
				ctx,
				"INSERT INTO metrics (id,type,delta,value) VALUES ($1,$2,$3,$4)",
				metric.ID, metric.MType, metric.Delta, metric.Value,
			)
			if err != nil {
				return metric, models.MetricSaveError{Message: err.Error()}
			}
			return metric, nil
		} else {
			return metric, models.MetricSaveError{Message: err.Error()}
		}
	}

	switch metric.MType {
	case models.Counter:
		*metric.Delta += *dbMetric.Delta
		_, err = storage.Database.Exec(
			ctx,
			"UPDATE metrics SET delta = $1 WHERE id = $2 AND type = $3",
			*metric.Delta, metric.ID, metric.MType,
		)
		if err != nil {
			return metric, models.MetricSaveError{Message: err.Error()}
		}
	case models.Gauge:
		_, err = storage.Database.Exec(
			ctx,
			"UPDATE metrics SET value = $1 WHERE id = $2 AND type = $3",
			*metric.Value, metric.ID, metric.MType,
		)
		if err != nil {
			return metric, models.MetricSaveError{Message: err.Error()}
		}
	}

	return metric, nil
}

func (storage *DataBaseStorage) UpdateOrCreateMetrics(ctx context.Context, metrics []models.Metrics) (int64, error) {
	count := int64(0)
	tx, err := storage.Database.Begin(ctx)
	if err != nil {
		return 0, models.MetricSaveError{Message: err.Error()}
	}
	defer tx.Rollback(ctx)
	existingMetrics := []models.Metrics{}
	nonExistingMetrics := []models.Metrics{}
	for _, m := range metrics {
		row := tx.QueryRow(
			ctx,
			"SELECT delta, value FROM metrics WHERE id = $1 AND type = $2",
			m.ID, m.MType,
		)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return 0, models.MetricSaveError{Message: err.Error()}
			}
		}
		dbMetric := models.Metrics{}
		err = row.Scan(&dbMetric.Delta, &dbMetric.Value)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				if slices.ContainsFunc(nonExistingMetrics, func(fm models.Metrics) bool {
					return fm.ID == m.ID && fm.MType == m.MType
				}) {
					existingMetrics = append(existingMetrics, m)
				} else {
					nonExistingMetrics = append(nonExistingMetrics, m)
				}
			} else {
				return 0, models.MetricSaveError{Message: err.Error()}
			}
		} else {
			if m.MType == models.Counter {
				*m.Delta += *dbMetric.Delta
			}
			existingMetrics = append(existingMetrics, m)
		}
	}
	for _, m := range existingMetrics {
		_, err = tx.Exec(
			ctx,
			"UPDATE metrics SET delta = $1, value = $2 WHERE id = $3 AND type = $4",
			m.Delta, m.Value, m.ID, m.MType,
		)
		if err != nil {
			return 0, models.MetricSaveError{Message: err.Error()}
		}
		count++
	}
	for _, m := range nonExistingMetrics {
		_, err = tx.Exec(
			ctx,
			"INSERT INTO metrics (id,type,delta,value) VALUES ($1,$2,$3,$4)",
			m.ID, m.MType, m.Delta, m.Value,
		)
		if err != nil {
			return 0, models.MetricSaveError{Message: err.Error()}
		}
		count++
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, models.MetricSaveError{Message: err.Error()}
	}
	return count, nil
}

func (storage *DataBaseStorage) GetAllMetrics(ctx context.Context) []models.Metrics {
	page := 0
	limit := 100
	result := []models.Metrics{}

	for {
		offset := page * limit
		rows, err := storage.Database.Query(
			ctx,
			"SELECT id, type, delta, value FROM metrics ORDER BY id LIMIT $1 OFFSET $2",
			limit, offset,
		)
		if err != nil {
			logger.Log.Sugar().Error(err)
			return []models.Metrics{}
		}
		defer rows.Close()
		count := 0
		for rows.Next() {
			metric := models.Metrics{}
			err = rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
			if err != nil {
				logger.Log.Sugar().Error(err)
				return []models.Metrics{}
			}
			result = append(result, metric)
			count++
		}
		if count == 0 {
			break
		}
		err = rows.Err()
		if err != nil {
			logger.Log.Sugar().Error(err)
			return []models.Metrics{}
		}
		page++
	}

	return result
}

func (storage *DataBaseStorage) SaveMetrics(ctx context.Context, metrics []models.Metrics) error {
	for _, metric := range metrics {

		row, err := storage.Database.QueryRow(
			ctx,
			"SELECT delta, value FROM metrics WHERE id = $1 AND type = $2",
			metric.ID, metric.MType,
		)
		if err != nil {
			return models.MetricSaveError{Message: err.Error()}
		}

		dbMetric := models.Metrics{}
		err = row.Scan(&dbMetric.Delta, &dbMetric.Value)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				_, err = storage.Database.Exec(
					ctx,
					"INSERT INTO metrics (id,type,delta,value) VALUES ($1,$2,$3,$4)",
					metric.ID, metric.MType, metric.Delta, metric.Value,
				)
				if err != nil {
					return models.MetricSaveError{Message: err.Error()}
				}
				continue
			} else {
				return models.MetricSaveError{Message: err.Error()}
			}
		}

		switch metric.MType {
		case models.Counter:
			_, err = storage.Database.Exec(
				ctx,
				"UPDATE metrics SET delta = $1 WHERE id = $2 AND type = $3",
				*metric.Delta, metric.ID, metric.MType,
			)
			if err != nil {
				return models.MetricSaveError{Message: err.Error()}
			}
		case models.Gauge:
			_, err = storage.Database.Exec(
				ctx,
				"UPDATE metrics SET value = $1 WHERE id = $2 AND type = $3",
				*metric.Value, metric.ID, metric.MType,
			)
			if err != nil {
				return models.MetricSaveError{Message: err.Error()}
			}
		}
	}
	return nil
}

func (storage *DataBaseStorage) GetCounter(ctx context.Context, name string) (int64, error) {
	row, err := storage.Database.QueryRow(
		ctx,
		"SELECT id, type, delta FROM metrics WHERE id = $1 AND type = $2",
		name, models.Counter,
	)
	if err != nil {
		return 0, models.MetricGetError{Message: err.Error()}
	}

	metric := models.Metrics{}
	err = row.Scan(&metric.ID, &metric.MType, &metric.Delta)
	if err != nil {
		return 0, models.MetricGetError{Message: err.Error()}
	}

	return *metric.Delta, nil
}

func (storage *DataBaseStorage) GetGauge(ctx context.Context, name string) (float64, error) {
	row, err := storage.Database.QueryRow(
		ctx,
		"SELECT id, type, value FROM metrics WHERE id = $1 AND type = $2",
		name, models.Gauge,
	)
	if err != nil {
		return 0, models.MetricGetError{Message: err.Error()}
	}

	metric := models.Metrics{}
	err = row.Scan(&metric.ID, &metric.MType, &metric.Value)
	if err != nil {
		return 0, models.MetricGetError{Message: err.Error()}
	}

	return *metric.Value, nil
}

func (storage *DataBaseStorage) GetMetricWithValue(ctx context.Context, metric *models.Metrics) (*models.Metrics, error) {
	row, err := storage.Database.QueryRow(
		ctx,
		"SELECT id, type, delta, value FROM metrics WHERE id = $1 AND type = $2",
		metric.ID, metric.MType,
	)
	if err != nil {
		return nil, models.MetricGetError{Message: err.Error()}
	}

	metricDB := models.Metrics{}
	err = row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
	if err != nil {
		return nil, models.MetricGetError{Message: err.Error()}
	}

	return &metricDB, nil
}
