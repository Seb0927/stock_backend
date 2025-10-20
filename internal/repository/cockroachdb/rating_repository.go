package cockroachdb

import (
	"context"
	"fmt"
	"time"

	"github.com/company/stock-api/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RatingRepository implements domain.RatingRepository for CockroachDB
type RatingRepository struct {
	db *pgxpool.Pool
}

// NewRatingRepository creates a new instance of RatingRepository
func NewRatingRepository(db *pgxpool.Pool) *RatingRepository {
	return &RatingRepository{
		db: db,
	}
}

// Create inserts a new rating record (term-only)
func (r *RatingRepository) Create(rating *domain.Rating) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO ratings (term)
		VALUES ($1)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query, rating.Term).Scan(
		&rating.ID,
		&rating.CreatedAt,
		&rating.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create rating: %w", err)
	}

	return nil
}

// FindByID retrieves a rating by its ID
func (r *RatingRepository) FindByID(id int64) (*domain.Rating, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, term, created_at, updated_at
		FROM ratings
		WHERE id = $1
	`

	rating := &domain.Rating{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&rating.ID,
		&rating.Term,
		&rating.CreatedAt,
		&rating.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to find rating: %w", err)
	}

	return rating, nil
}

// FindByTerm retrieves a rating by its term
func (r *RatingRepository) FindByTerm(term string) (*domain.Rating, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, term, created_at, updated_at
		FROM ratings
		WHERE term = $1
	`

	rating := &domain.Rating{}
	err := r.db.QueryRow(ctx, query, term).Scan(
		&rating.ID,
		&rating.Term,
		&rating.CreatedAt,
		&rating.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to find rating by term: %w", err)
	}

	return rating, nil
}

// FindAll retrieves all ratings
func (r *RatingRepository) FindAll(ctx context.Context) ([]*domain.Rating, error) {
	queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	query := `
		SELECT id, term, created_at, updated_at
		FROM ratings
		ORDER BY term ASC
	`

	rows, err := r.db.Query(queryCtx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query ratings: %w", err)
	}
	defer rows.Close()

	var ratings []*domain.Rating
	for rows.Next() {
		rating := &domain.Rating{}
		err := rows.Scan(
			&rating.ID,
			&rating.Term,
			&rating.CreatedAt,
			&rating.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rating: %w", err)
		}
		ratings = append(ratings, rating)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ratings: %w", err)
	}

	return ratings, nil
}

// (previous brokerage-scoped FindAll removed; ratings are global now)
