package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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
	return nil
}

func (db *Database) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

func (db *Database) Ping() error {
	if db.pool != nil {
		return db.pool.Ping(context.Background())
	}
	return fmt.Errorf("database connection was not opened")
}

func (db *Database) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	if db.pool != nil {
		return db.pool.Exec(ctx, query, args...)
	}
	return pgconn.CommandTag{}, fmt.Errorf("database connection was not opened")
}

func (db *Database) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	if db.pool != nil {
		return db.pool.Query(ctx, query, args...)
	}
	return nil, fmt.Errorf("database connection was not opened")
}

func (db *Database) QueryRow(ctx context.Context, query string, args ...any) (pgx.Row, error) {
	if db.pool != nil {
		return db.pool.QueryRow(ctx, query, args...), nil
	}
	return nil, fmt.Errorf("database connection was not opened")
}
