package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/scouser-122/go-metrics/internal/logger"
)

type Database struct {
	Config DbConnectionConfig
	pool   *pgxpool.Pool
}

func (db *Database) Open() error {
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
	return db.pool.Ping(context.Background())
}
