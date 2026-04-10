package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/config/db"
	models "github.com/scouser-122/go-metrics/internal/model"
	"github.com/stretchr/testify/mock"
)

type MockPostgresDBTestData struct {
	MockPool  *MockPostgresPool
	MockRow   pgx.Row
	MockTag   pgconn.CommandTag
	MockError error
}

func NewMockPostgresDB(serverConfig config.ServerConfig, mockPool *MockPostgresPool) PostgresDatabase {
	database := PostgresDatabase{
		Config: db.DBConnectionConfig{
			DSN:         serverConfig.DBDataSourceName,
			RetryConfig: config.DefaultRetryConfig(),
		},
		Pool: mockPool,
	}
	return database
}

// Мок для pgxpool.Pool
type MockPostgresPool struct {
	mock.Mock
	MockMethods func(tt MockPostgresDBTestData)
}

func (m *MockPostgresPool) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockPostgresPool) Close() {
	m.Called()
}

func (m *MockPostgresPool) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	args := m.Called(ctx, sql, arguments)
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}

func (m *MockPostgresPool) QueryRow(ctx context.Context, sql string, arguments ...interface{}) pgx.Row {
	args := m.Called(ctx, sql, arguments)
	return args.Get(0).(pgx.Row)
}

func (m *MockPostgresPool) Query(ctx context.Context, sql string, arguments ...interface{}) (pgx.Rows, error) {
	args := m.Called(ctx, sql, arguments)
	return args.Get(0).(pgx.Rows), args.Error(1)
}

func (m *MockPostgresPool) Begin(ctx context.Context) (pgx.Tx, error) {
	args := m.Called(ctx)
	return args.Get(0).(pgx.Tx), args.Error(1)
}

// Мок для pgx.Rows
type MockPostgresRows struct {
	mock.Mock
	metrics []models.Metrics
	index   int
}

func (m *MockPostgresRows) Next() bool {
	return m.index < len(m.metrics)
}

func (m *MockPostgresRows) Scan(dest ...interface{}) error {
	if m.index >= len(m.metrics) {
		return errors.New("no more rows")
	}

	metric := m.metrics[m.index]
	*dest[0].(*string) = metric.ID
	*dest[1].(*string) = metric.MType
	*dest[2].(**int64) = metric.Delta
	*dest[3].(**float64) = metric.Value
	m.index++
	return nil
}

func (m *MockPostgresRows) Close()     {}
func (m *MockPostgresRows) Err() error { return nil }

func (m *MockPostgresRows) CommandTag() pgconn.CommandTag {
	return pgconn.NewCommandTag("SELECT 3")
}

// Мок для pgx.Row
type MockPostgresRow struct {
	Metric *models.Metrics
	Err    error
}

func (m *MockPostgresRow) Scan(dest ...interface{}) error {
	if m.Err != nil {
		return m.Err
	}
	if m.Metric == nil {
		return pgx.ErrNoRows
	}

	*dest[0].(*string) = m.Metric.ID
	*dest[1].(*string) = m.Metric.MType
	if len(dest) == 3 {
		switch m.Metric.MType {
		case models.Counter:
			*dest[2].(**int64) = m.Metric.Delta
		case models.Gauge:
			*dest[2].(**float64) = m.Metric.Value
		}
	} else if len(dest) == 4 {
		*dest[2].(**int64) = m.Metric.Delta
		*dest[3].(**float64) = m.Metric.Value
	}
	return nil
}
