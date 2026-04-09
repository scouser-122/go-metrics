package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/logger"
	models "github.com/scouser-122/go-metrics/internal/model"
	"github.com/scouser-122/go-metrics/internal/repository/postgres"
)

type PostgresDBStorage struct {
	Database *postgres.PostgresDatabase
}

func (storage *PostgresDBStorage) UpdateOrCreateCounter(ctx context.Context, name string, value int64) (int64, error) {
	row, err := storage.Database.QueryRow(
		ctx,
		"SELECT id, type, delta FROM metrics WHERE id = $1 AND type = $2",
		name, models.Counter,
	)
	if err != nil {
		return value, models.MetricSaveError{Err: err}
	}

	metric := models.Metrics{}
	err = config.DataBaseRequestRetry(
		ctx,
		storage.Database.Config.RetryConfig,
		func() error {
			logger.Sugar.Info("try scan db")
			return row.Scan(&metric.ID, &metric.MType, &metric.Delta)
		},
	)
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
				return value, models.MetricSaveError{Err: err}
			}
			return value, nil
		} else {
			return value, models.MetricSaveError{Err: err}
		}
	}

	_, err = storage.Database.Exec(
		ctx,
		"UPDATE metrics SET delta = delta + $1 WHERE id = $2 AND type = $3",
		value, metric.ID, metric.MType,
	)
	if err != nil {
		return value, models.MetricSaveError{Err: err}
	}

	return *metric.Delta + value, nil
}

func (storage *PostgresDBStorage) UpdateOrCreateGauge(ctx context.Context, name string, value float64) (float64, error) {
	row, err := storage.Database.QueryRow(
		ctx,
		"SELECT id, type FROM metrics WHERE id = $1 AND type = $2",
		name, models.Gauge,
	)
	if err != nil {
		return value, models.MetricSaveError{Err: err}
	}

	metric := models.Metrics{}
	err = config.DataBaseRequestRetry(
		ctx,
		storage.Database.Config.RetryConfig,
		func() error {
			return row.Scan(&metric.ID, &metric.MType)
		},
	)
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
				return value, models.MetricSaveError{Err: err}
			}
			return value, nil
		} else {
			return value, models.MetricSaveError{Err: err}
		}
	}

	_, err = storage.Database.Exec(
		ctx,
		"UPDATE metrics SET value = $1 WHERE id = $2 AND type = $3",
		value, metric.ID, metric.MType,
	)
	if err != nil {
		return value, models.MetricSaveError{Err: err}
	}

	return value, nil
}

func (storage *PostgresDBStorage) UpdateOrCreateMetric(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	row, err := storage.Database.QueryRow(
		ctx,
		"SELECT delta, value FROM metrics WHERE id = $1 AND type = $2",
		metric.ID, metric.MType,
	)
	if err != nil {
		return metric, models.MetricSaveError{Err: err}
	}

	dbMetric := models.Metrics{}
	err = config.DataBaseRequestRetry(
		ctx,
		storage.Database.Config.RetryConfig,
		func() error {
			return row.Scan(&dbMetric.Delta, &dbMetric.Value)
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_, err = storage.Database.Exec(
				ctx,
				"INSERT INTO metrics (id,type,delta,value) VALUES ($1,$2,$3,$4)",
				metric.ID, metric.MType, metric.Delta, metric.Value,
			)
			if err != nil {
				return metric, models.MetricSaveError{Err: err}
			}
			return metric, nil
		} else {
			return metric, models.MetricSaveError{Err: err}
		}
	}

	switch metric.MType {
	case models.Counter:
		_, err = storage.Database.Exec(
			ctx,
			"UPDATE metrics SET delta = delta + $1 WHERE id = $2 AND type = $3",
			*metric.Delta, metric.ID, metric.MType,
		)
		if err != nil {
			return metric, models.MetricSaveError{Err: err}
		}
	case models.Gauge:
		_, err = storage.Database.Exec(
			ctx,
			"UPDATE metrics SET value = $1 WHERE id = $2 AND type = $3",
			*metric.Value, metric.ID, metric.MType,
		)
		if err != nil {
			return metric, models.MetricSaveError{Err: err}
		}
	}

	return metric, nil
}

func (storage *PostgresDBStorage) UpdateOrCreateMetrics(ctx context.Context, metrics []models.Metrics) (int64, error) {
	count := int64(0)
	tx, err := storage.Database.Begin(ctx)
	if err != nil {
		return 0, models.MetricSaveError{Err: err}
	}
	defer tx.Rollback(ctx)
	for _, m := range metrics {
		row := tx.QueryRow(
			ctx,
			"SELECT delta, value FROM metrics WHERE id = $1 AND type = $2",
			m.ID, m.MType,
		)
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return 0, models.MetricSaveError{Err: err}
			}
		}
		dbMetric := models.Metrics{}
		err = row.Scan(&dbMetric.Delta, &dbMetric.Value)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				_, err = tx.Exec(
					ctx,
					"INSERT INTO metrics (id,type,delta,value) VALUES ($1,$2,$3,$4)",
					m.ID, m.MType, m.Delta, m.Value,
				)
				if err != nil {
					return 0, models.MetricSaveError{Err: err}
				}
				count++
			} else {
				return 0, models.MetricSaveError{Err: err}
			}
		} else {
			if m.MType == models.Counter {
				*m.Delta += *dbMetric.Delta
			}
			_, err = tx.Exec(
				ctx,
				"UPDATE metrics SET delta = $1, value = $2 WHERE id = $3 AND type = $4",
				m.Delta, m.Value, m.ID, m.MType,
			)
			if err != nil {
				return 0, models.MetricSaveError{Err: err}
			}
			count++
		}
	}
	err = config.DataBaseRequestRetry(
		ctx,
		storage.Database.Config.RetryConfig,
		func() error {
			return tx.Commit(ctx)
		},
	)
	if err != nil {
		return 0, models.MetricSaveError{Err: err}
	}
	return count, nil
}

func (storage *PostgresDBStorage) GetAllMetrics(ctx context.Context) []models.Metrics {
	page := 0
	limit := 10
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
			rows.Close()
			return []models.Metrics{}
		}
		count := 0
		for rows.Next() {
			metric := models.Metrics{}
			err = rows.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
			if err != nil {
				logger.Log.Sugar().Error(err)
				rows.Close()
				return []models.Metrics{}
			}
			result = append(result, metric)
			count++
		}
		err = rows.Err()
		if err != nil {
			logger.Log.Sugar().Error(err)
			rows.Close()
			return []models.Metrics{}
		}
		rows.Close()
		if count == 0 {
			break
		}
		page++
	}

	return result
}

func (storage *PostgresDBStorage) SaveMetrics(ctx context.Context, metrics []models.Metrics) error {
	tx, err := storage.Database.Begin(ctx)
	if err != nil {
		return models.MetricSaveError{Err: err}
	}
	defer tx.Rollback(ctx)
	for _, metric := range metrics {

		row := tx.QueryRow(
			ctx,
			"SELECT delta, value FROM metrics WHERE id = $1 AND type = $2",
			metric.ID, metric.MType,
		)
		if err != nil {
			return models.MetricSaveError{Err: err}
		}

		dbMetric := models.Metrics{}
		err = row.Scan(&dbMetric.Delta, &dbMetric.Value)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				_, err = tx.Exec(
					ctx,
					"INSERT INTO metrics (id,type,delta,value) VALUES ($1,$2,$3,$4)",
					metric.ID, metric.MType, metric.Delta, metric.Value,
				)
				if err != nil {
					return models.MetricSaveError{Err: err}
				}
				continue
			} else {
				return models.MetricSaveError{Err: err}
			}
		}
		_, err = tx.Exec(
			ctx,
			"UPDATE metrics SET delta = $1, value = $2 WHERE id = $3 AND type = $4",
			metric.Delta, metric.Value, metric.ID, metric.MType,
		)
	}
	err = config.DataBaseRequestRetry(
		ctx,
		storage.Database.Config.RetryConfig,
		func() error {
			return tx.Commit(ctx)
		},
	)
	if err != nil {
		return models.MetricSaveError{Err: err}
	}
	return nil
}

func (storage *PostgresDBStorage) GetCounter(ctx context.Context, name string) (int64, error) {
	row, err := storage.Database.QueryRow(
		ctx,
		"SELECT id, type, delta FROM metrics WHERE id = $1 AND type = $2",
		name, models.Counter,
	)
	if err != nil {
		return 0, models.MetricGetError{Err: err}
	}

	metric := models.Metrics{}
	err = config.DataBaseRequestRetry(
		ctx,
		storage.Database.Config.RetryConfig,
		func() error {
			return row.Scan(&metric.ID, &metric.MType, &metric.Delta)
		},
	)
	if err != nil {
		return 0, models.MetricGetError{Err: err}
	}

	return *metric.Delta, nil
}

func (storage *PostgresDBStorage) GetGauge(ctx context.Context, name string) (float64, error) {
	row, err := storage.Database.QueryRow(
		ctx,
		"SELECT id, type, value FROM metrics WHERE id = $1 AND type = $2",
		name, models.Gauge,
	)
	if err != nil {
		return 0, models.MetricGetError{Err: err}
	}

	metric := models.Metrics{}
	err = config.DataBaseRequestRetry(
		ctx,
		storage.Database.Config.RetryConfig,
		func() error {
			return row.Scan(&metric.ID, &metric.MType, &metric.Value)
		},
	)
	if err != nil {
		return 0, models.MetricGetError{Err: err}
	}

	return *metric.Value, nil
}

func (storage *PostgresDBStorage) GetMetricWithValue(ctx context.Context, metric *models.Metrics) (*models.Metrics, error) {
	row, err := storage.Database.QueryRow(
		ctx,
		"SELECT id, type, delta, value FROM metrics WHERE id = $1 AND type = $2",
		metric.ID, metric.MType,
	)
	if err != nil {
		return nil, models.MetricGetError{Err: err}
	}

	metricDB := models.Metrics{}
	err = config.DataBaseRequestRetry(
		ctx,
		storage.Database.Config.RetryConfig,
		func() error {
			return row.Scan(&metricDB.ID, &metricDB.MType, &metricDB.Delta, &metricDB.Value)
		},
	)
	if err != nil {
		return nil, models.MetricGetError{Err: err}
	}

	return &metricDB, nil
}
