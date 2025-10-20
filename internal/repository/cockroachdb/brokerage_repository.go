package cockroachdb

import (
	"context"
	"fmt"
	"time"

	"github.com/company/stock-api/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BrokerageRepository implements domain.BrokerageRepository for CockroachDB
type BrokerageRepository struct {
	db *pgxpool.Pool
}

// NewBrokerageRepository creates a new instance of BrokerageRepository
func NewBrokerageRepository(db *pgxpool.Pool) *BrokerageRepository {
	return &BrokerageRepository{
		db: db,
	}
}

// Create inserts a new brokerage record
func (r *BrokerageRepository) Create(brokerage *domain.Brokerage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO brokerages (name)
		VALUES ($1)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query, brokerage.Name).Scan(
		&brokerage.ID,
		&brokerage.CreatedAt,
		&brokerage.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create brokerage: %w", err)
	}

	return nil
}

// FindByID retrieves a brokerage by its ID
func (r *BrokerageRepository) FindByID(id int64) (*domain.Brokerage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, name, created_at, updated_at
		FROM brokerages
		WHERE id = $1
	`

	brokerage := &domain.Brokerage{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&brokerage.ID,
		&brokerage.Name,
		&brokerage.CreatedAt,
		&brokerage.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to find brokerage: %w", err)
	}

	return brokerage, nil
}

// FindByName retrieves a brokerage by its name
func (r *BrokerageRepository) FindByName(name string) (*domain.Brokerage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, name, created_at, updated_at
		FROM brokerages
		WHERE name = $1
	`

	brokerage := &domain.Brokerage{}
	err := r.db.QueryRow(ctx, query, name).Scan(
		&brokerage.ID,
		&brokerage.Name,
		&brokerage.CreatedAt,
		&brokerage.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to find brokerage by name: %w", err)
	}

	return brokerage, nil
}

// FindAll retrieves all brokerages
func (r *BrokerageRepository) FindAll(ctx context.Context) ([]*domain.Brokerage, error) {
	queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	query := `
		SELECT id, name, created_at, updated_at
		FROM brokerages
		ORDER BY name ASC
	`

	rows, err := r.db.Query(queryCtx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query brokerages: %w", err)
	}
	defer rows.Close()

	var brokerages []*domain.Brokerage
	for rows.Next() {
		brokerage := &domain.Brokerage{}
		err := rows.Scan(
			&brokerage.ID,
			&brokerage.Name,
			&brokerage.CreatedAt,
			&brokerage.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan brokerage: %w", err)
		}
		brokerages = append(brokerages, brokerage)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating brokerages: %w", err)
	}

	return brokerages, nil
}
