package cockroachdb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/company/stock-api/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// StockRepository implements domain.StockRepository for CockroachDB
type StockRepository struct {
	db *pgxpool.Pool
}

// NewStockRepository creates a new instance of StockRepository
func NewStockRepository(db *pgxpool.Pool) *StockRepository {
	return &StockRepository{
		db: db,
	}
}

// CreateBatch inserts multiple stock records in a single transaction
func (r *StockRepository) CreateBatch(stocks []*domain.Stock) error {
	if len(stocks) == 0 {
		return nil
	}

	// Process in chunks to avoid timeouts and improve performance
	const chunkSize = 100
	for i := 0; i < len(stocks); i += chunkSize {
		end := i + chunkSize
		if end > len(stocks) {
			end = len(stocks)
		}

		chunk := stocks[i:end]
		if err := r.insertChunk(chunk); err != nil {
			return fmt.Errorf("failed to insert batch chunk: %w", err)
		}
	}

	return nil
}

// insertChunk inserts a chunk of stocks in a single transaction
func (r *StockRepository) insertChunk(stocks []*domain.Stock) error {
	// Timeout per chunk (1 minute should be plenty)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO stocks (ticker, target_from, target_to, company, action, brokerage, rating_from, rating_to, time)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (ticker, company, time) DO NOTHING
		RETURNING id, created_at, updated_at
	`

	for _, stock := range stocks {
		err := tx.QueryRow(ctx, query,
			stock.Ticker,
			stock.TargetFrom,
			stock.TargetTo,
			stock.Company,
			stock.Action,
			stock.Brokerage,
			stock.RatingFrom,
			stock.RatingTo,
			stock.Time,
		).Scan(&stock.ID, &stock.CreatedAt, &stock.UpdatedAt)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("failed to insert stock: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// FindByID retrieves a stock by its ID
func (r *StockRepository) FindByID(id int64) (*domain.Stock, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, ticker, target_from, target_to, company, action, brokerage, 
		       rating_from, rating_to, time, created_at, updated_at
		FROM stocks
		WHERE id = $1
	`

	stock := &domain.Stock{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&stock.ID,
		&stock.Ticker,
		&stock.TargetFrom,
		&stock.TargetTo,
		&stock.Company,
		&stock.Action,
		&stock.Brokerage,
		&stock.RatingFrom,
		&stock.RatingTo,
		&stock.Time,
		&stock.CreatedAt,
		&stock.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find stock: %w", err)
	}

	return stock, nil
}

// FindAll retrieves stocks based on filters
func (r *StockRepository) FindAll(filter domain.StockFilter) ([]*domain.Stock, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		SELECT id, ticker, target_from, target_to, company, action, brokerage,
		       rating_from, rating_to, time, created_at, updated_at
		FROM stocks
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	if filter.Ticker != "" {
		query += fmt.Sprintf(" AND ticker = $%d", argPos)
		args = append(args, filter.Ticker)
		argPos++
	}

	if filter.Company != "" {
		query += fmt.Sprintf(" AND company ILIKE $%d", argPos)
		args = append(args, "%"+filter.Company+"%")
		argPos++
	}

	if filter.Brokerage != "" {
		query += fmt.Sprintf(" AND brokerage ILIKE $%d", argPos)
		args = append(args, "%"+filter.Brokerage+"%")
		argPos++
	}

	if filter.Action != "" {
		query += fmt.Sprintf(" AND action = $%d", argPos)
		args = append(args, filter.Action)
		argPos++
	}

	if filter.RatingFrom != "" {
		query += fmt.Sprintf(" AND rating_from = $%d", argPos)
		args = append(args, filter.RatingFrom)
		argPos++
	}

	if filter.RatingTo != "" {
		query += fmt.Sprintf(" AND rating_to = $%d", argPos)
		args = append(args, filter.RatingTo)
		argPos++
	}

	// Build ORDER BY clause
	sortBy := "time"
	if filter.SortBy != "" {
		// Validate sortBy to prevent SQL injection
		validSortFields := map[string]bool{
			"ticker":    true,
			"company":   true,
			"time":      true,
			"rating_to": true,
			"action":    true,
			"brokerage": true,
			"target_to": true,
		}
		if validSortFields[filter.SortBy] {
			sortBy = filter.SortBy
		}
	}

	sortOrder := "DESC"
	if filter.SortOrder == "asc" || filter.SortOrder == "ASC" {
		sortOrder = "ASC"
	}

	query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, filter.Limit)
		argPos++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query stocks: %w", err)
	}
	defer rows.Close()

	stocks := []*domain.Stock{}
	for rows.Next() {
		stock := &domain.Stock{}
		err := rows.Scan(
			&stock.ID,
			&stock.Ticker,
			&stock.TargetFrom,
			&stock.TargetTo,
			&stock.Company,
			&stock.Action,
			&stock.Brokerage,
			&stock.RatingFrom,
			&stock.RatingTo,
			&stock.Time,
			&stock.CreatedAt,
			&stock.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stock: %w", err)
		}
		stocks = append(stocks, stock)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stocks: %w", err)
	}

	return stocks, nil
}

// Count returns the total number of stocks matching the filter
func (r *StockRepository) Count(filter domain.StockFilter) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT COUNT(*) FROM stocks WHERE 1=1`
	args := []interface{}{}
	argPos := 1

	if filter.Ticker != "" {
		query += fmt.Sprintf(" AND ticker = $%d", argPos)
		args = append(args, filter.Ticker)
		argPos++
	}

	if filter.Company != "" {
		query += fmt.Sprintf(" AND company ILIKE $%d", argPos)
		args = append(args, "%"+filter.Company+"%")
		argPos++
	}

	if filter.Brokerage != "" {
		query += fmt.Sprintf(" AND brokerage ILIKE $%d", argPos)
		args = append(args, "%"+filter.Brokerage+"%")
		argPos++
	}

	if filter.Action != "" {
		query += fmt.Sprintf(" AND action = $%d", argPos)
		args = append(args, filter.Action)
		argPos++
	}

	if filter.RatingFrom != "" {
		query += fmt.Sprintf(" AND rating_from = $%d", argPos)
		args = append(args, filter.RatingFrom)
		argPos++
	}

	if filter.RatingTo != "" {
		query += fmt.Sprintf(" AND rating_to = $%d", argPos)
		args = append(args, filter.RatingTo)
	}

	var count int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count stocks: %w", err)
	}

	return count, nil
}
