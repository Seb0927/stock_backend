package cockroachdb

import (
	"context"
	"fmt"
	"time"

	"github.com/company/stock-api/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewConnection creates a new database connection pool
func NewConnection(cfg *config.DatabaseConfig) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configure connection pool
	poolConfig.MaxConns = int32(cfg.MaxConns)
	poolConfig.MinConns = int32(cfg.MinConns)
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

// InitSchema initializes the database schema
func InitSchema(db *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	schema := `
		CREATE TABLE IF NOT EXISTS stocks (
			id SERIAL PRIMARY KEY,
			ticker VARCHAR(20) NOT NULL,
			target_from VARCHAR(50),
			target_to VARCHAR(50),
			company VARCHAR(255) NOT NULL,
			action VARCHAR(100),
			brokerage VARCHAR(255),
			rating_from VARCHAR(50),
			rating_to VARCHAR(50),
			time TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW(),
			UNIQUE(ticker, company, time)
		);

		CREATE INDEX IF NOT EXISTS idx_stocks_ticker ON stocks(ticker);
		CREATE INDEX IF NOT EXISTS idx_stocks_company ON stocks(company);
		CREATE INDEX IF NOT EXISTS idx_stocks_time ON stocks(time DESC);
		CREATE INDEX IF NOT EXISTS idx_stocks_brokerage ON stocks(brokerage);
	`

	_, err := db.Exec(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}
