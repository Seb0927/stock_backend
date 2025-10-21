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
	db            *pgxpool.Pool
	brokerageRepo *BrokerageRepository
	actionRepo    *ActionRepository
	ratingRepo    *RatingRepository
}

// getStringValue safely dereferences a *string returning empty string if nil
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// NewStockRepository creates a new instance of StockRepository
func NewStockRepository(db *pgxpool.Pool, brokerageRepo *BrokerageRepository, actionRepo *ActionRepository, ratingRepo *RatingRepository) *StockRepository {
	return &StockRepository{
		db:            db,
		brokerageRepo: brokerageRepo,
		actionRepo:    actionRepo,
		ratingRepo:    ratingRepo,
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
		INSERT INTO stocks (ticker, target_from, target_to, company, action_id, brokerage_id, rating_from_id, rating_to_id, time)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (ticker, company, time) DO NOTHING
		RETURNING id, created_at, updated_at
	`

	for _, stock := range stocks {
		// Convert 0 to NULL for foreign keys (0 means no value)
		var actionID, brokerageID, ratingFromID, ratingToID interface{}
		if stock.ActionID > 0 {
			actionID = stock.ActionID
		}
		if stock.BrokerageID > 0 {
			brokerageID = stock.BrokerageID
		}
		if stock.RatingFromID > 0 {
			ratingFromID = stock.RatingFromID
		}
		if stock.RatingToID > 0 {
			ratingToID = stock.RatingToID
		}

		err := tx.QueryRow(ctx, query,
			stock.Ticker,
			stock.TargetFrom,
			stock.TargetTo,
			stock.Company,
			actionID,
			brokerageID,
			ratingFromID,
			ratingToID,
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

// FindByID retrieves a stock by its ID with all joined details
func (r *StockRepository) FindByID(id int64) (*domain.StockWithDetails, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT 
			s.id, s.ticker, s.target_from, s.target_to, s.company,
			s.action_id, a.name as action_name,
			s.brokerage_id, b.name as brokerage_name,
			s.rating_from_id, rf.term as rating_from_term,
			s.rating_to_id, rt.term as rating_to_term,
			s.time, s.created_at, s.updated_at
		FROM stocks s
		LEFT JOIN actions a ON s.action_id = a.id
		LEFT JOIN brokerages b ON s.brokerage_id = b.id
		LEFT JOIN ratings rf ON s.rating_from_id = rf.id
		LEFT JOIN ratings rt ON s.rating_to_id = rt.id
		WHERE s.id = $1
	`

	stock := &domain.StockWithDetails{}

	// Use nullable types for scanning
	var actionID, brokerageID, ratingFromID, ratingToID *int64
	var actionName, brokerageName, ratingFromTerm, ratingToTerm *string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&stock.ID,
		&stock.Ticker,
		&stock.TargetFrom,
		&stock.TargetTo,
		&stock.Company,
		&actionID,
		&actionName,
		&brokerageID,
		&brokerageName,
		&ratingFromID,
		&ratingFromTerm,
		&ratingToID,
		&ratingToTerm,
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

	// Assign nullable fields
	stock.ActionID = actionID
	stock.ActionName = getStringValue(actionName)
	stock.BrokerageID = brokerageID
	stock.BrokerageName = getStringValue(brokerageName)
	stock.RatingFromID = ratingFromID
	stock.RatingFromTerm = getStringValue(ratingFromTerm)
	stock.RatingToID = ratingToID
	stock.RatingToTerm = getStringValue(ratingToTerm)

	return stock, nil
}

// FindByTicker retrieves all stock records for a given ticker (all historical versions)
func (r *StockRepository) FindByTicker(ticker string) ([]*domain.StockWithDetails, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT 
			s.id, s.ticker, s.target_from, s.target_to, s.company,
			s.action_id, a.name as action_name,
			s.brokerage_id, b.name as brokerage_name,
			s.rating_from_id, rf.term as rating_from_term,
			s.rating_to_id, rt.term as rating_to_term,
			s.time, s.created_at, s.updated_at
		FROM stocks s
		LEFT JOIN actions a ON s.action_id = a.id
		LEFT JOIN brokerages b ON s.brokerage_id = b.id
		LEFT JOIN ratings rf ON s.rating_from_id = rf.id
		LEFT JOIN ratings rt ON s.rating_to_id = rt.id
		WHERE s.ticker = $1
		ORDER BY s.time DESC
	`

	rows, err := r.db.Query(ctx, query, ticker)
	if err != nil {
		return nil, fmt.Errorf("failed to query stocks by ticker: %w", err)
	}
	defer rows.Close()

	var stocks []*domain.StockWithDetails
	for rows.Next() {
		stock := &domain.StockWithDetails{}

		// Use nullable types for scanning
		var actionID, brokerageID, ratingFromID, ratingToID *int64
		var actionName, brokerageName, ratingFromTerm, ratingToTerm *string

		err := rows.Scan(
			&stock.ID,
			&stock.Ticker,
			&stock.TargetFrom,
			&stock.TargetTo,
			&stock.Company,
			&actionID,
			&actionName,
			&brokerageID,
			&brokerageName,
			&ratingFromID,
			&ratingFromTerm,
			&ratingToID,
			&ratingToTerm,
			&stock.Time,
			&stock.CreatedAt,
			&stock.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stock: %w", err)
		}

		// Assign nullable fields
		stock.ActionID = actionID
		stock.ActionName = getStringValue(actionName)
		stock.BrokerageID = brokerageID
		stock.BrokerageName = getStringValue(brokerageName)
		stock.RatingFromID = ratingFromID
		stock.RatingFromTerm = getStringValue(ratingFromTerm)
		stock.RatingToID = ratingToID
		stock.RatingToTerm = getStringValue(ratingToTerm)

		stocks = append(stocks, stock)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stocks: %w", err)
	}

	if len(stocks) == 0 {
		return nil, domain.ErrNotFound
	}

	return stocks, nil
}

// FindAll retrieves stocks based on filters
func (r *StockRepository) FindAll(filter domain.StockFilter) ([]*domain.StockWithDetails, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use a subquery to get only the latest stock per ticker
	// This prevents duplicates when stocks are updated over time
	query := `
		WITH latest_stocks AS (
			SELECT DISTINCT ON (s.ticker) 
				s.id, s.ticker, s.target_from, s.target_to, s.company,
				s.action_id, a.name as action_name,
				s.brokerage_id, b.name as brokerage_name,
				s.rating_from_id, rf.term as rating_from_term,
				s.rating_to_id, rt.term as rating_to_term,
				s.time, s.created_at, s.updated_at
			FROM stocks s
			LEFT JOIN actions a ON s.action_id = a.id
			LEFT JOIN brokerages b ON s.brokerage_id = b.id
			LEFT JOIN ratings rf ON s.rating_from_id = rf.id
			LEFT JOIN ratings rt ON s.rating_to_id = rt.id
			ORDER BY s.ticker, s.time DESC
		)
		SELECT id, ticker, target_from, target_to, company,
		       action_id, action_name, brokerage_id, brokerage_name,
		       rating_from_id, rating_from_term, rating_to_id, rating_to_term,
		       time, created_at, updated_at
		FROM latest_stocks
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
		// Fuzzy search with similarity matching (handles typos like "Aple" -> "Apple")
		// Using trigram similarity: matches if similarity > 0.3 (configurable threshold)
		query += fmt.Sprintf(" AND (company ILIKE $%d OR company %% $%d)", argPos, argPos+1)
		searchTerm := filter.Company
		args = append(args, "%"+searchTerm+"%") // ILIKE pattern matching
		args = append(args, searchTerm)         // Trigram similarity matching
		argPos += 2
	}

	if filter.Brokerage != "" {
		// Fuzzy search with similarity matching (handles typos)
		query += fmt.Sprintf(" AND (brokerage_name ILIKE $%d OR brokerage_name %% $%d)", argPos, argPos+1)
		searchTerm := filter.Brokerage
		args = append(args, "%"+searchTerm+"%") // ILIKE pattern matching
		args = append(args, searchTerm)         // Trigram similarity matching
		argPos += 2
	}

	if filter.Action != "" {
		query += fmt.Sprintf(" AND action_name = $%d", argPos)
		args = append(args, filter.Action)
		argPos++
	}

	if filter.RatingFrom != "" {
		query += fmt.Sprintf(" AND rating_from_term = $%d", argPos)
		args = append(args, filter.RatingFrom)
		argPos++
	}

	if filter.RatingTo != "" {
		query += fmt.Sprintf(" AND rating_to_term = $%d", argPos)
		args = append(args, filter.RatingTo)
		argPos++
	}

	// Build ORDER BY clause
	sortBy := "time"
	if filter.SortBy != "" {
		// Validate sortBy to prevent SQL injection
		validSortFields := map[string]bool{
			"ticker":         true,
			"company":        true,
			"time":           true,
			"rating_to_term": true,
			"action_name":    true,
			"brokerage_name": true,
			"target_to":      true,
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

	stocks := []*domain.StockWithDetails{}
	for rows.Next() {
		stock := &domain.StockWithDetails{}

		// Use nullable types for scanning
		var actionID, brokerageID, ratingFromID, ratingToID *int64
		var actionName, brokerageName, ratingFromTerm, ratingToTerm *string

		err := rows.Scan(
			&stock.ID,
			&stock.Ticker,
			&stock.TargetFrom,
			&stock.TargetTo,
			&stock.Company,
			&actionID,
			&actionName,
			&brokerageID,
			&brokerageName,
			&ratingFromID,
			&ratingFromTerm,
			&ratingToID,
			&ratingToTerm,
			&stock.Time,
			&stock.CreatedAt,
			&stock.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stock: %w", err)
		}

		// Assign nullable fields
		stock.ActionID = actionID
		stock.ActionName = getStringValue(actionName)
		stock.BrokerageID = brokerageID
		stock.BrokerageName = getStringValue(brokerageName)
		stock.RatingFromID = ratingFromID
		stock.RatingFromTerm = getStringValue(ratingFromTerm)
		stock.RatingToID = ratingToID
		stock.RatingToTerm = getStringValue(ratingToTerm)

		stocks = append(stocks, stock)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stocks: %w", err)
	}

	return stocks, nil
}

// Count returns the total number of unique stocks (latest per ticker) matching the filter
func (r *StockRepository) Count(filter domain.StockFilter) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Count only the latest version of each ticker
	query := `
		WITH latest_stocks AS (
			SELECT DISTINCT ON (s.ticker) 
				s.id, s.ticker, s.target_from, s.target_to, s.company,
				s.action_id, a.name as action_name,
				s.brokerage_id, b.name as brokerage_name,
				s.rating_from_id, rf.term as rating_from_term,
				s.rating_to_id, rt.term as rating_to_term,
				s.time, s.created_at, s.updated_at
			FROM stocks s
			LEFT JOIN actions a ON s.action_id = a.id
			LEFT JOIN brokerages b ON s.brokerage_id = b.id
			LEFT JOIN ratings rf ON s.rating_from_id = rf.id
			LEFT JOIN ratings rt ON s.rating_to_id = rt.id
			ORDER BY s.ticker, s.time DESC
		)
		SELECT COUNT(*) FROM latest_stocks WHERE 1=1
	`
	args := []interface{}{}
	argPos := 1

	if filter.Ticker != "" {
		query += fmt.Sprintf(" AND ticker = $%d", argPos)
		args = append(args, filter.Ticker)
		argPos++
	}

	if filter.Company != "" {
		// Fuzzy search with similarity matching (handles typos like "Aple" -> "Apple")
		// Using trigram similarity: matches if similarity > 0.3 (configurable threshold)
		query += fmt.Sprintf(" AND (company ILIKE $%d OR company %% $%d)", argPos, argPos+1)
		searchTerm := filter.Company
		args = append(args, "%"+searchTerm+"%") // ILIKE pattern matching
		args = append(args, searchTerm)         // Trigram similarity matching
		argPos += 2
	}

	if filter.Brokerage != "" {
		// Fuzzy search with similarity matching (handles typos)
		query += fmt.Sprintf(" AND (brokerage_name ILIKE $%d OR brokerage_name %% $%d)", argPos, argPos+1)
		searchTerm := filter.Brokerage
		args = append(args, "%"+searchTerm+"%") // ILIKE pattern matching
		args = append(args, searchTerm)         // Trigram similarity matching
		argPos += 2
	}

	if filter.Action != "" {
		query += fmt.Sprintf(" AND action_name = $%d", argPos)
		args = append(args, filter.Action)
		argPos++
	}

	if filter.RatingFrom != "" {
		query += fmt.Sprintf(" AND rating_from_term = $%d", argPos)
		args = append(args, filter.RatingFrom)
		argPos++
	}

	if filter.RatingTo != "" {
		query += fmt.Sprintf(" AND rating_to_term = $%d", argPos)
		args = append(args, filter.RatingTo)
	}

	var count int64
	err := r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count stocks: %w", err)
	}

	return count, nil
}
