package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps a pgx pool to provide a common interface.
type DB struct {
	pool *pgxpool.Pool
}

// Pool returns the underlying pgx pool.
func (d *DB) Pool() *pgxpool.Pool {
	return d.pool
}

// Close closes the database connection pool.
func (d *DB) Close() {
	d.pool.Close()
}

// NewDB creates a new database connection.
func NewDB(ctx context.Context, dsn string) (*DB, error) {
	pool, err := NewPool(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &DB{pool: pool}, nil
}

// NewPool creates a new pgx connection pool with production-ready settings.
func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return pool, nil
}
