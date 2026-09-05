package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/havilz/caelus-cloud/backend/pkg/config"
	"github.com/havilz/caelus-cloud/backend/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Client struct {
	Pool *pgxpool.Pool
}

func NewClient(ctx context.Context, cfg *config.DatabaseConfig) (*Client, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("gagal mem-parsing konfigurasi database: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)
	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime
	poolConfig.MaxConnIdleTime = 15 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat connection pool database: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("gagal melakukan ping ke database PostgreSQL: %w", err)
	}

	logger.Info("PostgreSQL database connection established successfully",
		"host", cfg.Host,
		"port", cfg.Port,
		"database", cfg.DBName,
	)

	return &Client{Pool: pool}, nil
}

func (c *Client) Close() {
	if c.Pool != nil {
		c.Pool.Close()
		logger.Info("PostgreSQL database connection pool closed")
	}
}

func (c *Client) Ping(ctx context.Context) error {
	return c.Pool.Ping(ctx)
}
