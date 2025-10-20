package cockroachdb

import (
	"context"
	"fmt"
	"time"

	"github.com/company/stock-api/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ActionRepository implements domain.ActionRepository for CockroachDB
type ActionRepository struct {
	db *pgxpool.Pool
}

// NewActionRepository creates a new instance of ActionRepository
func NewActionRepository(db *pgxpool.Pool) *ActionRepository {
	return &ActionRepository{
		db: db,
	}
}

// Create inserts a new action record
func (r *ActionRepository) Create(action *domain.Action) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO actions (name)
		VALUES ($1)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query, action.Name).Scan(
		&action.ID,
		&action.CreatedAt,
		&action.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create action: %w", err)
	}

	return nil
}

// FindByID retrieves an action by its ID
func (r *ActionRepository) FindByID(id int64) (*domain.Action, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, name, created_at, updated_at
		FROM actions
		WHERE id = $1
	`

	action := &domain.Action{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&action.ID,
		&action.Name,
		&action.CreatedAt,
		&action.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to find action: %w", err)
	}

	return action, nil
}

// FindByName retrieves an action by its name
func (r *ActionRepository) FindByName(name string) (*domain.Action, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, name, created_at, updated_at
		FROM actions
		WHERE name = $1
	`

	action := &domain.Action{}
	err := r.db.QueryRow(ctx, query, name).Scan(
		&action.ID,
		&action.Name,
		&action.CreatedAt,
		&action.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to find action by name: %w", err)
	}

	return action, nil
}

// FindAll retrieves all actions
func (r *ActionRepository) FindAll(ctx context.Context) ([]*domain.Action, error) {
	queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	query := `
		SELECT id, name, created_at, updated_at
		FROM actions
		ORDER BY name ASC
	`

	rows, err := r.db.Query(queryCtx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query actions: %w", err)
	}
	defer rows.Close()

	var actions []*domain.Action
	for rows.Next() {
		action := &domain.Action{}
		err := rows.Scan(
			&action.ID,
			&action.Name,
			&action.CreatedAt,
			&action.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan action: %w", err)
		}
		actions = append(actions, action)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating actions: %w", err)
	}

	return actions, nil
}
