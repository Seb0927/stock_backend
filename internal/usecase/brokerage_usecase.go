package usecase

import (
	"context"
	"fmt"

	"github.com/company/stock-api/internal/domain"
	"go.uber.org/zap"
)

// BrokerageUseCase handles business logic for brokerage operations
type BrokerageUseCase struct {
	repo   domain.BrokerageRepository
	logger *zap.Logger
}

// NewBrokerageUseCase creates a new BrokerageUseCase
func NewBrokerageUseCase(repo domain.BrokerageRepository, logger *zap.Logger) *BrokerageUseCase {
	return &BrokerageUseCase{
		repo:   repo,
		logger: logger,
	}
}

// GetAll retrieves all brokerages
func (uc *BrokerageUseCase) GetAll(ctx context.Context) ([]*domain.Brokerage, error) {
	brokerages, err := uc.repo.FindAll(ctx)
	if err != nil {
		uc.logger.Error("Failed to retrieve brokerages", zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve brokerages: %w", err)
	}

	return brokerages, nil
}

// GetByID retrieves a single brokerage by ID
func (uc *BrokerageUseCase) GetByID(ctx context.Context, id int64) (*domain.Brokerage, error) {
	brokerage, err := uc.repo.FindByID(id)
	if err != nil {
		uc.logger.Error("Failed to retrieve brokerage", zap.Int64("id", id), zap.Error(err))
		return nil, err
	}

	return brokerage, nil
}

// GetOrCreate retrieves a brokerage by name or creates it if it doesn't exist
func (uc *BrokerageUseCase) GetOrCreate(ctx context.Context, name string) (*domain.Brokerage, error) {
	// Try to find existing brokerage
	brokerage, err := uc.repo.FindByName(name)
	if err == nil {
		return brokerage, nil
	}

	// Create new brokerage if not found
	brokerage = &domain.Brokerage{
		Name: name,
	}

	err = uc.repo.Create(brokerage)
	if err != nil {
		uc.logger.Error("Failed to create brokerage", zap.String("name", name), zap.Error(err))
		return nil, fmt.Errorf("failed to create brokerage: %w", err)
	}

	uc.logger.Info("Created new brokerage", zap.String("name", name), zap.Int64("id", brokerage.ID))
	return brokerage, nil
}
