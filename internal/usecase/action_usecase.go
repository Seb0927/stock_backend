package usecase

import (
	"context"
	"fmt"

	"github.com/company/stock-api/internal/domain"
	"go.uber.org/zap"
)

// ActionUseCase handles business logic for action operations
type ActionUseCase struct {
	repo   domain.ActionRepository
	logger *zap.Logger
}

// NewActionUseCase creates a new ActionUseCase
func NewActionUseCase(repo domain.ActionRepository, logger *zap.Logger) *ActionUseCase {
	return &ActionUseCase{
		repo:   repo,
		logger: logger,
	}
}

// GetAll retrieves all actions
func (uc *ActionUseCase) GetAll(ctx context.Context) ([]*domain.Action, error) {
	actions, err := uc.repo.FindAll(ctx)
	if err != nil {
		uc.logger.Error("Failed to retrieve actions", zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve actions: %w", err)
	}

	return actions, nil
}

// GetByID retrieves a single action by ID
func (uc *ActionUseCase) GetByID(ctx context.Context, id int64) (*domain.Action, error) {
	action, err := uc.repo.FindByID(id)
	if err != nil {
		uc.logger.Error("Failed to retrieve action", zap.Int64("id", id), zap.Error(err))
		return nil, err
	}

	return action, nil
}

// GetOrCreate retrieves an action by name or creates it if it doesn't exist
func (uc *ActionUseCase) GetOrCreate(ctx context.Context, name string) (*domain.Action, error) {
	// Try to find existing action
	action, err := uc.repo.FindByName(name)
	if err == nil {
		return action, nil
	}

	// Create new action if not found
	action = &domain.Action{
		Name: name,
	}

	err = uc.repo.Create(action)
	if err != nil {
		uc.logger.Error("Failed to create action", zap.String("name", name), zap.Error(err))
		return nil, fmt.Errorf("failed to create action: %w", err)
	}

	uc.logger.Info("Created new action", zap.String("name", name), zap.Int64("id", action.ID))
	return action, nil
}
