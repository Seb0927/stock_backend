package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/truora/stock-api/internal/client"
	"github.com/truora/stock-api/internal/domain"
	"go.uber.org/zap"
)

// StockUseCase handles business logic for stock operations
type StockUseCase struct {
	repo      domain.StockRepository
	apiClient *client.StockAPIClient
	logger    *zap.Logger
}

// NewStockUseCase creates a new StockUseCase
func NewStockUseCase(repo domain.StockRepository, apiClient *client.StockAPIClient, logger *zap.Logger) *StockUseCase {
	return &StockUseCase{
		repo:      repo,
		apiClient: apiClient,
		logger:    logger,
	}
}

// SyncStocksFromAPI fetches stocks from external API and stores them in the database
func (uc *StockUseCase) SyncStocksFromAPI(ctx context.Context) (int, error) {
	uc.logger.Info("Starting stock sync from external API")
	startTime := time.Now()

	stocks, err := uc.apiClient.FetchAllStocks(ctx)
	if err != nil {
		uc.logger.Error("Failed to fetch stocks from API", zap.Error(err))
		return 0, fmt.Errorf("failed to fetch stocks: %w", err)
	}

	uc.logger.Info("Fetched stocks from API", zap.Int("count", len(stocks)))

	// Store stocks in batches
	uc.logger.Info("Starting database insert", zap.Int("total_stocks", len(stocks)))
	if err := uc.repo.CreateBatch(stocks); err != nil {
		uc.logger.Error("Failed to store stocks in database", zap.Error(err))
		return 0, fmt.Errorf("failed to store stocks: %w", err)
	}
	uc.logger.Info("Database insert completed successfully")

	duration := time.Since(startTime)
	uc.logger.Info("Stock sync completed",
		zap.Int("count", len(stocks)),
		zap.Duration("duration", duration))

	return len(stocks), nil
}

// GetStocks retrieves stocks with filters
func (uc *StockUseCase) GetStocks(ctx context.Context, filter domain.StockFilter) ([]*domain.Stock, error) {
	// Set default pagination if not provided
	if filter.Limit == 0 {
		filter.Limit = 50
	}
	if filter.Limit > 1000 {
		filter.Limit = 1000
	}

	stocks, err := uc.repo.FindAll(filter)
	if err != nil {
		uc.logger.Error("Failed to retrieve stocks", zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve stocks: %w", err)
	}

	return stocks, nil
}

// GetStockByID retrieves a single stock by ID
func (uc *StockUseCase) GetStockByID(ctx context.Context, id int64) (*domain.Stock, error) {
	stock, err := uc.repo.FindByID(id)
	if err != nil {
		uc.logger.Error("Failed to retrieve stock", zap.Int64("id", id), zap.Error(err))
		return nil, err
	}

	return stock, nil
}

// GetStockCount returns the total count of stocks matching the filter
func (uc *StockUseCase) GetStockCount(ctx context.Context, filter domain.StockFilter) (int64, error) {
	count, err := uc.repo.Count(filter)
	if err != nil {
		uc.logger.Error("Failed to count stocks", zap.Error(err))
		return 0, fmt.Errorf("failed to count stocks: %w", err)
	}

	return count, nil
}
