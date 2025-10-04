package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/company/stock-api/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockStockRepository is a mock implementation of domain.StockRepository
type MockStockRepository struct {
	mock.Mock
}

func (m *MockStockRepository) CreateBatch(stocks []*domain.Stock) error {
	args := m.Called(stocks)
	return args.Error(0)
}

func (m *MockStockRepository) FindByID(id int64) (*domain.Stock, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Stock), args.Error(1)
}

func (m *MockStockRepository) FindAll(filter domain.StockFilter) ([]*domain.Stock, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Stock), args.Error(1)
}

func (m *MockStockRepository) Count(filter domain.StockFilter) (int64, error) {
	args := m.Called(filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStockRepository) FindByTicker(ticker string) ([]*domain.Stock, error) {
	args := m.Called(ticker)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Stock), args.Error(1)
}

// MockStockAPIClient is a mock implementation of the stock API client
type MockStockAPIClient struct {
	mock.Mock
}

func (m *MockStockAPIClient) FetchAllStocks(ctx context.Context) ([]*domain.Stock, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Stock), args.Error(1)
}

func TestStockUseCase_GetStockByID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(MockStockRepository)

	useCase := NewStockUseCase(mockRepo, nil, logger)

	t.Run("Success", func(t *testing.T) {
		expectedStock := &domain.Stock{
			ID:      1,
			Ticker:  "AAPL",
			Company: "Apple Inc.",
			Time:    time.Now(),
		}

		mockRepo.On("FindByID", int64(1)).Return(expectedStock, nil).Once()

		stock, err := useCase.GetStockByID(context.Background(), 1)

		assert.NoError(t, err)
		assert.NotNil(t, stock)
		assert.Equal(t, expectedStock.ID, stock.ID)
		assert.Equal(t, expectedStock.Ticker, stock.Ticker)
		mockRepo.AssertExpectations(t)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockRepo.On("FindByID", int64(999)).Return(nil, domain.ErrNotFound).Once()

		stock, err := useCase.GetStockByID(context.Background(), 999)

		assert.Error(t, err)
		assert.Nil(t, stock)
		assert.Equal(t, domain.ErrNotFound, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestStockUseCase_GetStocks(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(MockStockRepository)

	useCase := NewStockUseCase(mockRepo, nil, logger)

	t.Run("Success with default pagination", func(t *testing.T) {
		expectedStocks := []*domain.Stock{
			{ID: 1, Ticker: "AAPL", Company: "Apple Inc.", Time: time.Now()},
			{ID: 2, Ticker: "GOOGL", Company: "Google", Time: time.Now()},
		}

		filter := domain.StockFilter{Limit: 50}
		mockRepo.On("FindAll", filter).Return(expectedStocks, nil).Once()

		stocks, err := useCase.GetStocks(context.Background(), filter)

		assert.NoError(t, err)
		assert.Len(t, stocks, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success with custom filter", func(t *testing.T) {
		expectedStocks := []*domain.Stock{
			{ID: 1, Ticker: "AAPL", Company: "Apple Inc.", Time: time.Now()},
		}

		filter := domain.StockFilter{
			Ticker: "AAPL",
			Limit:  10,
			Offset: 0,
		}
		mockRepo.On("FindAll", filter).Return(expectedStocks, nil).Once()

		stocks, err := useCase.GetStocks(context.Background(), filter)

		assert.NoError(t, err)
		assert.Len(t, stocks, 1)
		assert.Equal(t, "AAPL", stocks[0].Ticker)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success with rating filters", func(t *testing.T) {
		expectedStocks := []*domain.Stock{
			{ID: 1, Ticker: "AAPL", Company: "Apple Inc.", RatingFrom: "Neutral", RatingTo: "Overweight", Time: time.Now()},
		}

		filter := domain.StockFilter{
			RatingFrom: "Neutral",
			RatingTo:   "Overweight",
			Limit:      50,
		}
		mockRepo.On("FindAll", filter).Return(expectedStocks, nil).Once()

		stocks, err := useCase.GetStocks(context.Background(), filter)

		assert.NoError(t, err)
		assert.Len(t, stocks, 1)
		assert.Equal(t, "Neutral", stocks[0].RatingFrom)
		assert.Equal(t, "Overweight", stocks[0].RatingTo)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success with sorting parameters", func(t *testing.T) {
		expectedStocks := []*domain.Stock{
			{ID: 1, Ticker: "AAPL", Company: "Apple Inc.", Time: time.Now()},
			{ID: 2, Ticker: "GOOGL", Company: "Google", Time: time.Now()},
		}

		filter := domain.StockFilter{
			SortBy:    "ticker",
			SortOrder: "asc",
			Limit:     50,
		}
		mockRepo.On("FindAll", filter).Return(expectedStocks, nil).Once()

		stocks, err := useCase.GetStocks(context.Background(), filter)

		assert.NoError(t, err)
		assert.Len(t, stocks, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success with complex filters", func(t *testing.T) {
		expectedStocks := []*domain.Stock{
			{ID: 1, Ticker: "AAPL", Company: "Apple Inc.", RatingTo: "Overweight", Time: time.Now()},
		}

		filter := domain.StockFilter{
			Company:   "Apple",
			RatingTo:  "Overweight",
			SortBy:    "time",
			SortOrder: "desc",
			Limit:     10,
			Offset:    0,
		}
		mockRepo.On("FindAll", filter).Return(expectedStocks, nil).Once()

		stocks, err := useCase.GetStocks(context.Background(), filter)

		assert.NoError(t, err)
		assert.Len(t, stocks, 1)
		assert.Equal(t, "Apple Inc.", stocks[0].Company)
		assert.Equal(t, "Overweight", stocks[0].RatingTo)
		mockRepo.AssertExpectations(t)
	})
}

func TestStockUseCase_GetStockCount(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(MockStockRepository)

	useCase := NewStockUseCase(mockRepo, nil, logger)

	t.Run("Success", func(t *testing.T) {
		filter := domain.StockFilter{Ticker: "AAPL"}
		mockRepo.On("Count", filter).Return(int64(42), nil).Once()

		count, err := useCase.GetStockCount(context.Background(), filter)

		assert.NoError(t, err)
		assert.Equal(t, int64(42), count)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		filter := domain.StockFilter{}
		mockRepo.On("Count", filter).Return(int64(0), errors.New("database error")).Once()

		count, err := useCase.GetStockCount(context.Background(), filter)

		assert.Error(t, err)
		assert.Equal(t, int64(0), count)
		mockRepo.AssertExpectations(t)
	})
}

func TestStockUseCase_SyncStocksFromAPI(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockRepo := new(MockStockRepository)
	mockAPIClient := new(MockStockAPIClient)

	useCase := NewStockUseCase(mockRepo, mockAPIClient, logger)

	t.Run("Success", func(t *testing.T) {
		stocks := []*domain.Stock{
			{Ticker: "AAPL", Company: "Apple Inc.", Time: time.Now()},
			{Ticker: "GOOGL", Company: "Google", Time: time.Now()},
		}

		mockAPIClient.On("FetchAllStocks", mock.Anything).Return(stocks, nil).Once()
		mockRepo.On("CreateBatch", stocks).Return(nil).Once()

		count, err := useCase.SyncStocksFromAPI(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, 2, count)
		mockAPIClient.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("API Fetch Error", func(t *testing.T) {
		mockAPIClient.On("FetchAllStocks", mock.Anything).Return(nil, domain.ErrExternalAPI).Once()

		count, err := useCase.SyncStocksFromAPI(context.Background())

		assert.Error(t, err)
		assert.Equal(t, 0, count)
		mockAPIClient.AssertExpectations(t)
	})

	t.Run("Database Insert Error", func(t *testing.T) {
		stocks := []*domain.Stock{
			{Ticker: "AAPL", Company: "Apple Inc.", Time: time.Now()},
		}

		mockAPIClient.On("FetchAllStocks", mock.Anything).Return(stocks, nil).Once()
		mockRepo.On("CreateBatch", stocks).Return(errors.New("database error")).Once()

		count, err := useCase.SyncStocksFromAPI(context.Background())

		assert.Error(t, err)
		assert.Equal(t, 0, count)
		mockAPIClient.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}
