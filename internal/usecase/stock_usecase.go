package usecase

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/company/stock-api/internal/domain"
	"go.uber.org/zap"
)

// StockUseCase handles business logic for stock operations
type StockUseCase struct {
	repo      domain.StockRepository
	apiClient domain.StockAPIClient
	logger    *zap.Logger
}

// NewStockUseCase creates a new StockUseCase
func NewStockUseCase(repo domain.StockRepository, apiClient domain.StockAPIClient, logger *zap.Logger) *StockUseCase {
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

// GetStocksByTicker retrieves all historical versions of a stock by ticker
func (uc *StockUseCase) GetStocksByTicker(ctx context.Context, ticker string) ([]*domain.Stock, error) {
	stocks, err := uc.repo.FindByTicker(ticker)
	if err != nil {
		uc.logger.Error("Failed to retrieve stocks by ticker", zap.String("ticker", ticker), zap.Error(err))
		return nil, err
	}

	return stocks, nil
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

// GetRecommendations analyzes stocks and returns the best investment recommendations
func (uc *StockUseCase) GetRecommendations(ctx context.Context, limit int) ([]*domain.StockRecommendation, error) {
	uc.logger.Info("Generating stock recommendations", zap.Int("limit", limit))

	// Get all latest stocks (deduplicated by ticker)
	filter := domain.StockFilter{
		Limit: 1000, // Get a large set to analyze
	}
	stocks, err := uc.repo.FindAll(filter)
	if err != nil {
		uc.logger.Error("Failed to retrieve stocks for recommendations", zap.Error(err))
		return nil, fmt.Errorf("failed to retrieve stocks: %w", err)
	}

	if len(stocks) == 0 {
		return []*domain.StockRecommendation{}, nil
	}

	// Calculate scores for each stock
	recommendations := make([]*domain.StockRecommendation, 0, len(stocks))
	for _, stock := range stocks {
		score, reason, targetIncrease := uc.calculateStockScore(stock)

		recommendations = append(recommendations, &domain.StockRecommendation{
			Stock:          stock,
			Score:          score,
			Reason:         reason,
			TargetIncrease: targetIncrease,
		})
	}

	// Sort by score (descending)
	for i := 0; i < len(recommendations)-1; i++ {
		for j := i + 1; j < len(recommendations); j++ {
			if recommendations[j].Score > recommendations[i].Score {
				recommendations[i], recommendations[j] = recommendations[j], recommendations[i]
			}
		}
	}

	// Return top N recommendations
	if limit > 0 && limit < len(recommendations) {
		recommendations = recommendations[:limit]
	}

	uc.logger.Info("Generated recommendations",
		zap.Int("total_analyzed", len(stocks)),
		zap.Int("returned", len(recommendations)))

	return recommendations, nil
}

// calculateStockScore calculates a score for a stock based on multiple factors
func (uc *StockUseCase) calculateStockScore(stock *domain.Stock) (float64, string, float64) {
	var score float64
	reasons := []string{}

	// 1. Action Score (30% weight) - upgrade is best
	actionScore := uc.getActionScore(stock.Action)
	score += actionScore * 0.30
	if actionScore > 3 {
		reasons = append(reasons, fmt.Sprintf("Recent %s", stock.Action))
	}

	// 2. Rating Improvement Score (25% weight)
	ratingScore := uc.getRatingImprovementScore(stock.RatingFrom, stock.RatingTo)
	score += ratingScore * 0.25
	if ratingScore > 3 {
		reasons = append(reasons, fmt.Sprintf("Rating improved to %s", stock.RatingTo))
	}

	// 3. Target Price Increase (20% weight)
	targetIncrease := uc.calculateTargetPriceIncrease(stock.TargetFrom, stock.TargetTo)
	if targetIncrease != 0 {
		// Normalize: 10% increase = 5 points, 20% = 10 points, etc.
		targetScore := (targetIncrease / 2.0)
		if targetScore > 10 {
			targetScore = 10 // Cap at 10
		}
		if targetScore < -10 {
			targetScore = -10 // Floor at -10
		}
		score += targetScore * 0.20
		if targetIncrease > 5 {
			reasons = append(reasons, fmt.Sprintf("%.1f%% price target increase", targetIncrease))
		} else if targetIncrease < -5 {
			reasons = append(reasons, fmt.Sprintf("%.1f%% price target decrease", targetIncrease))
		}
	}

	// 4. Recency Score (15% weight) - more recent is better
	recencyScore := uc.getRecencyScore(stock.Time)
	score += recencyScore * 0.15

	// 5. Brokerage Reputation (10% weight)
	brokerageScore := uc.getBrokerageScore(stock.Brokerage)
	score += brokerageScore * 0.10
	if brokerageScore >= 8 && stock.Brokerage != "" {
		reasons = append(reasons, fmt.Sprintf("Rated by %s", stock.Brokerage))
	}

	// Build reason string
	reason := strings.Join(reasons, "; ")
	if reason == "" {
		reason = "Positive outlook"
	}

	return score, reason, targetIncrease
}

// getActionScore returns a score based on the action type
func (uc *StockUseCase) getActionScore(action string) float64 {
	action = strings.ToLower(action)
	switch {
	case strings.Contains(action, "upgrade"):
		return 10.0
	case strings.Contains(action, "initiated") || strings.Contains(action, "initiate"):
		return 8.0
	case strings.Contains(action, "target") && strings.Contains(action, "raised"):
		return 7.0
	case strings.Contains(action, "reiterate") || strings.Contains(action, "maintain"):
		return 6.0
	case strings.Contains(action, "target") && strings.Contains(action, "lowered"):
		return 3.0 // Negative signal
	case strings.Contains(action, "downgrade"):
		return 2.0
	default:
		return 5.0 // neutral
	}
}

// getRatingImprovementScore compares rating_from to rating_to
func (uc *StockUseCase) getRatingImprovementScore(ratingFrom, ratingTo string) float64 {
	ratingValues := map[string]float64{
		"strong-buy":        5.0,
		"strong buy":        5.0,
		"buy":               4.0,
		"speculative buy":   4.0,
		"overweight":        4.0,
		"outperform":        4.0,
		"market outperform": 4.0,
		"sector outperform": 4.0,
		"positive":          4.0,
		"hold":              3.0,
		"neutral":           3.0,
		"in-line":           3.0,
		"market perform":    3.0,
		"sector perform":    3.0,
		"equal weight":      3.0,
		"equal-weight":      3.0,
		"underweight":       2.0,
		"underperform":      2.0,
		"reduce":            2.0,
		"sell":              1.0,
	}

	fromValue := uc.getRatingValue(ratingFrom, ratingValues)
	toValue := uc.getRatingValue(ratingTo, ratingValues)

	// Calculate improvement bonus
	improvementBonus := 0.0
	if toValue > fromValue {
		// Bigger improvement gets larger bonus
		improvementBonus = (toValue - fromValue) * 2.0
	} else if toValue < fromValue {
		// Downgrade penalty
		improvementBonus = (toValue - fromValue) * 2.0 // This will be negative
	}

	// Return the final rating value plus improvement bonus
	// Scale to 0-10 range: multiply by 2 to convert 1-5 scale to 2-10 scale
	return (toValue * 2.0) + improvementBonus
}

// getRatingValue gets the numeric value for a rating
func (uc *StockUseCase) getRatingValue(rating string, ratingValues map[string]float64) float64 {
	rating = strings.ToLower(strings.TrimSpace(rating))

	// Handle empty rating
	if rating == "" {
		return 3.0 // Default to neutral
	}

	if val, ok := ratingValues[rating]; ok {
		return val
	}
	// Default to neutral if unknown
	return 3.0
}

// calculateTargetPriceIncrease calculates the percentage increase from target_from to target_to
func (uc *StockUseCase) calculateTargetPriceIncrease(targetFrom, targetTo string) float64 {
	from := uc.parsePrice(targetFrom)
	to := uc.parsePrice(targetTo)

	if from <= 0 || to <= 0 {
		return 0
	}

	increase := ((to - from) / from) * 100
	return increase
}

// parsePrice extracts numeric value from price strings like "$200.00", "$2,700.00" or "$85"
func (uc *StockUseCase) parsePrice(priceStr string) float64 {
	// Remove currency symbols and commas
	priceStr = strings.TrimSpace(priceStr)
	priceStr = strings.ReplaceAll(priceStr, "$", "")
	priceStr = strings.ReplaceAll(priceStr, "€", "")
	priceStr = strings.ReplaceAll(priceStr, ",", "") // Handle $2,700.00 format
	priceStr = strings.ReplaceAll(priceStr, " ", "")

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0
	}
	return price
}

// getRecencyScore scores based on how recent the stock data is
func (uc *StockUseCase) getRecencyScore(t time.Time) float64 {
	daysSince := time.Since(t).Hours() / 24

	switch {
	case daysSince <= 1:
		return 10.0 // Today
	case daysSince <= 7:
		return 8.0 // This week
	case daysSince <= 30:
		return 6.0 // This month
	case daysSince <= 90:
		return 4.0 // Last 3 months
	default:
		return 2.0 // Older
	}
}

// getBrokerageScore scores based on brokerage reputation
func (uc *StockUseCase) getBrokerageScore(brokerage string) float64 {
	brokerage = strings.ToLower(strings.TrimSpace(brokerage))

	// Handle empty brokerage
	if brokerage == "" {
		return 5.0 // Neutral score for unknown brokerage
	}

	// Top-tier brokerages
	topTier := []string{"goldman sachs", "morgan stanley", "jp morgan", "jpmorgan", "barclays"}
	for _, top := range topTier {
		if strings.Contains(brokerage, top) {
			return 10.0
		}
	}

	// Mid-tier brokerages
	midTier := []string{"citigroup", "credit suisse", "deutsche bank", "ubs", "wells fargo"}
	for _, mid := range midTier {
		if strings.Contains(brokerage, mid) {
			return 8.0
		}
	}

	// Default score for other brokerages
	return 6.0
}
