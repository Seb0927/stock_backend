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
		-- Enable pg_trgm extension for fuzzy search
		CREATE EXTENSION IF NOT EXISTS pg_trgm;

		-- Brokerages table
		CREATE TABLE IF NOT EXISTS brokerages (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);

		-- Actions table
		CREATE TABLE IF NOT EXISTS actions (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL UNIQUE,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);

		-- Ratings table (global rating terms, independent of brokerage)
		CREATE TABLE IF NOT EXISTS ratings (
			id SERIAL PRIMARY KEY,
			term VARCHAR(50) NOT NULL UNIQUE,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		);

		-- Stocks table with foreign keys
		CREATE TABLE IF NOT EXISTS stocks (
			id SERIAL PRIMARY KEY,
			ticker VARCHAR(20) NOT NULL,
			target_from VARCHAR(50),
			target_to VARCHAR(50),
			company VARCHAR(255) NOT NULL,
			action_id INT,
			brokerage_id INT,
			rating_from_id INT,
			rating_to_id INT,
			time TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW(),
			FOREIGN KEY (action_id) REFERENCES actions(id),
			FOREIGN KEY (brokerage_id) REFERENCES brokerages(id),
			FOREIGN KEY (rating_from_id) REFERENCES ratings(id),
			FOREIGN KEY (rating_to_id) REFERENCES ratings(id),
			UNIQUE(ticker, company, time)
		);

		-- Indexes for stocks
		CREATE INDEX IF NOT EXISTS idx_stocks_ticker ON stocks(ticker);
		CREATE INDEX IF NOT EXISTS idx_stocks_company ON stocks(company);
		CREATE INDEX IF NOT EXISTS idx_stocks_time ON stocks(time DESC);
		CREATE INDEX IF NOT EXISTS idx_stocks_brokerage_id ON stocks(brokerage_id);
		CREATE INDEX IF NOT EXISTS idx_stocks_action_id ON stocks(action_id);

		-- Trigram indexes for fuzzy search
		CREATE INDEX IF NOT EXISTS idx_stocks_company_trgm ON stocks USING GIN (company gin_trgm_ops);
		CREATE INDEX IF NOT EXISTS idx_brokerages_name_trgm ON brokerages USING GIN (name gin_trgm_ops);
	`

	_, err := db.Exec(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}
