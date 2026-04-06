package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/scouser-122/go-metrics/internal/config"
	"github.com/scouser-122/go-metrics/internal/logger"
)

type Database struct {
	Config DBConnectionConfig
	pool   *pgxpool.Pool
}

func (db *Database) Open() error {
	if db.Config.DSN == "" {
		return fmt.Errorf("connection string is empty")
	}

	var err error
	config, err := pgxpool.ParseConfig(db.Config.DSN)
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %q, err: %w", db.Config.DSN, err)
	}

	db.pool, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := db.pool.Ping(context.Background()); err != nil {
		db.pool.Close()
		db.pool = nil
		return fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Sugar.Info("successfully connected to DB")

	err = db.runMigrations()
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func (db *Database) runMigrations() error {
	path, err := getMigrationsPath()
	if err != nil {
		return err
	}
	m, err := migrate.New(
		path,
		db.Config.DSN,
	)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	logger.Sugar.Info("DB migrations completed successfully")
	return nil
}

func getMigrationsPath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}

	execDir := filepath.Dir(execPath)

	migrationsPath := filepath.Join(execDir, "../../migrations")

	return "file://" + migrationsPath, nil
}

func (db *Database) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

func (db *Database) Ping(ctx context.Context) error {
	if db.pool != nil {
		return config.DataBaseRequestRetry(
			ctx,
			db.Config.RetryConfig,
			func() error {
				return db.pool.Ping(ctx)
			},
		)
	}
	return fmt.Errorf("database connection was not opened")
}

func (db *Database) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	if db.pool != nil {
		var commandTag pgconn.CommandTag
		var err error
		err = config.DataBaseRequestRetry(
			ctx,
			db.Config.RetryConfig,
			func() error {
				var err error
				commandTag, err = db.pool.Exec(ctx, query, args...)
				return err
			},
		)
		return commandTag, err
	}
	return pgconn.CommandTag{}, fmt.Errorf("database connection was not opened")
}

func (db *Database) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	if db.pool != nil {
		var rows pgx.Rows
		var err error
		err = config.DataBaseRequestRetry(
			ctx,
			db.Config.RetryConfig,
			func() error {
				var err error
				rows, err = db.pool.Query(ctx, query, args...)
				return err
			},
		)
		return rows, err
	}
	return nil, fmt.Errorf("database connection was not opened")
}

func (db *Database) QueryRow(ctx context.Context, query string, args ...any) (pgx.Row, error) {
	if db.pool != nil {
		return db.pool.QueryRow(ctx, query, args...), nil
	}
	return nil, fmt.Errorf("database connection was not opened")
}

func (db *Database) Begin(ctx context.Context) (pgx.Tx, error) {
	if db.pool != nil {
		var tx pgx.Tx
		var err error
		err = config.DataBaseRequestRetry(
			ctx,
			db.Config.RetryConfig,
			func() error {
				var err error
				tx, err = db.pool.Begin(ctx)
				return err
			},
		)
		return tx, err
	}
	return nil, fmt.Errorf("database connection was not opened")
}
