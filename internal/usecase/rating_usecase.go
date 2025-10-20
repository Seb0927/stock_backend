package usecase

import (
	"context"
	"fmt"

	"github.com/company/stock-api/internal/domain"
	"go.uber.org/zap"
)

// RatingUseCase handles business logic for rating operations
type RatingUseCase struct {
	repo   domain.RatingRepository
	logger *zap.Logger
}

// NewRatingUseCase creates a new RatingUseCase
func NewRatingUseCase(repo domain.RatingRepository, logger *zap.Logger) *RatingUseCase {
	return &RatingUseCase{
		repo:   repo,
		logger: logger,
	}
}

// GetAll retrieves all ratings
func (uc *RatingUseCase) GetAll(ctx context.Context) ([]*domain.Rating, error) {
	ratings, err := uc.repo.FindAll(ctx)
	if err != nil {
		uc.logger.Error("Failed to retrieve ratings", zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve ratings: %w", err)
	}

	return ratings, nil
}

// GetByID retrieves a single rating by ID
func (uc *RatingUseCase) GetByID(ctx context.Context, id int64) (*domain.Rating, error) {
	rating, err := uc.repo.FindByID(id)
	if err != nil {
		uc.logger.Error("Failed to retrieve rating", zap.Int64("id", id), zap.Error(err))
		return nil, err
	}

	return rating, nil
}

// GetOrCreate retrieves a rating by term or creates it if it doesn't exist
func (uc *RatingUseCase) GetOrCreate(ctx context.Context, term string) (*domain.Rating, error) {
	// Try to find existing rating
	rating, err := uc.repo.FindByTerm(term)
	if err == nil {
		return rating, nil
	}

	// Create new rating if not found
	rating = &domain.Rating{
		Term: term,
	}

	err = uc.repo.Create(rating)
	if err != nil {
		uc.logger.Error("Failed to create rating",
			zap.String("term", term),
			zap.Error(err))
		return nil, fmt.Errorf("failed to create rating: %w", err)
	}

	uc.logger.Info("Created new rating",
		zap.String("term", term),
		zap.Int64("id", rating.ID))
	return rating, nil
}
