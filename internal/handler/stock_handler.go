package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/company/stock-api/internal/domain"
	"github.com/company/stock-api/internal/usecase"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// StockHandler handles HTTP requests for stock operations
type StockHandler struct {
	useCase     *usecase.StockUseCase
	brokerageUC *usecase.BrokerageUseCase
	actionUC    *usecase.ActionUseCase
	ratingUC    *usecase.RatingUseCase
	logger      *zap.Logger
}

// NewStockHandler creates a new StockHandler
func NewStockHandler(useCase *usecase.StockUseCase, brokerageUC *usecase.BrokerageUseCase, actionUC *usecase.ActionUseCase, ratingUC *usecase.RatingUseCase, logger *zap.Logger) *StockHandler {
	return &StockHandler{
		useCase:     useCase,
		brokerageUC: brokerageUC,
		actionUC:    actionUC,
		ratingUC:    ratingUC,
		logger:      logger,
	}
}

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Meta    MetaData    `json:"meta"`
}

// MetaData contains pagination metadata
type MetaData struct {
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

// SyncStocks godoc
// @Summary Sync stocks from external API
// @Description Fetches all stocks from the external API and stores them in the database
// @Tags stocks
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/stocks/sync [post]
func (h *StockHandler) SyncStocks(c *gin.Context) {
	count, err := h.useCase.SyncStocksFromAPI(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to sync stocks", zap.Error(err))
		h.respondWithError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Stocks synced successfully",
		Data: map[string]interface{}{
			"synced_count": count,
		},
	})
}

// GetStocks godoc
// @Summary Get stocks
// @Description Retrieves stocks with optional filtering and pagination
// @Tags stocks
// @Accept json
// @Produce json
// @Param ticker query string false "Filter by ticker"
// @Param company query string false "Filter by company name (partial match)"
// @Param brokerage query string false "Filter by brokerage name (partial match)"
// @Param action query string false "Filter by action"
// @Param rating_from query string false "Filter by rating_from"
// @Param rating_to query string false "Filter by rating_to"
// @Param sortBy query string false "Sort by field (ticker, company, time, rating_to, action)" default(time)
// @Param sortOrder query string false "Sort order (asc, desc)" default(desc)
// @Param limit query int false "Number of items per page" default(50)
// @Param offset query int false "Number of items to skip" default(0)
// @Success 200 {object} PaginatedResponse
// @Failure 500 {object} Response
// @Router /api/v1/stocks [get]
func (h *StockHandler) GetStocks(c *gin.Context) {
	filter := domain.StockFilter{
		Ticker:     c.Query("ticker"),
		Company:    c.Query("company"),
		Brokerage:  c.Query("brokerage"),
		Action:     c.Query("action"),
		RatingFrom: c.Query("rating_from"),
		RatingTo:   c.Query("rating_to"),
		SortBy:     c.DefaultQuery("sortBy", "time"),
		SortOrder:  c.DefaultQuery("sortOrder", "desc"),
		Limit:      h.parseIntQuery(c, "limit", 50),
		Offset:     h.parseIntQuery(c, "offset", 0),
	}

	stocks, err := h.useCase.GetStocks(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to get stocks", zap.Error(err))
		h.respondWithError(c, http.StatusInternalServerError, err)
		return
	}

	total, _ := h.useCase.GetStockCount(c.Request.Context(), filter)

	c.JSON(http.StatusOK, PaginatedResponse{
		Success: true,
		Data:    stocks,
		Meta: MetaData{
			Total:  total,
			Limit:  filter.Limit,
			Offset: filter.Offset,
		},
	})
}

// GetStockByID godoc
// @Summary Get stock by ID
// @Description Retrieves a single stock by its ID
// @Tags stocks
// @Accept json
// @Produce json
// @Param id path int true "Stock ID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/stocks/{id} [get]
func (h *StockHandler) GetStockByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, domain.ErrInvalidInput)
		return
	}

	stock, err := h.useCase.GetStockByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			h.respondWithError(c, http.StatusNotFound, err)
			return
		}
		h.logger.Error("Failed to get stock", zap.Int64("id", id), zap.Error(err))
		h.respondWithError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    stock,
	})
}

// GetStocksByTicker godoc
// @Summary Get all historical versions of a stock by ticker
// @Description Retrieves all stock records for a given ticker symbol, ordered by time (newest first)
// @Tags stocks
// @Accept json
// @Produce json
// @Param ticker path string true "Stock ticker symbol (e.g., AAPL, GOOGL)"
// @Success 200 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/stock/{ticker} [get]
func (h *StockHandler) GetStocksByTicker(c *gin.Context) {
	ticker := c.Param("ticker")
	if ticker == "" {
		h.respondWithError(c, http.StatusBadRequest, domain.ErrInvalidInput)
		return
	}

	stocks, err := h.useCase.GetStocksByTicker(c.Request.Context(), ticker)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			h.respondWithError(c, http.StatusNotFound, err)
			return
		}
		h.logger.Error("Failed to get stocks by ticker", zap.String("ticker", ticker), zap.Error(err))
		h.respondWithError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    stocks,
	})
}

// GetRecommendations godoc
// @Summary Get stock recommendations
// @Description Analyzes stock data and returns the best investment recommendations based on ratings, actions, target prices, and recency
// @Tags stocks
// @Accept json
// @Produce json
// @Param limit query int false "Number of recommendations to return" default(10)
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/recommendations [get]
func (h *StockHandler) GetRecommendations(c *gin.Context) {
	limit := h.parseIntQuery(c, "limit", 10)

	// Cap limit at 50
	if limit > 50 {
		limit = 50
	}
	if limit < 1 {
		limit = 10
	}

	recommendations, err := h.useCase.GetRecommendations(c.Request.Context(), limit)
	if err != nil {
		h.logger.Error("Failed to get recommendations", zap.Error(err))
		h.respondWithError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    recommendations,
		Message: fmt.Sprintf("Top %d stock recommendations based on recent ratings, actions, and target prices", len(recommendations)),
	})
}

// HealthCheck godoc
// @Summary Health check
// @Description Check if the API is healthy
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Router /health [get]
func (h *StockHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "Service is healthy",
	})
}

// Helper methods

func (h *StockHandler) parseIntQuery(c *gin.Context, key string, defaultValue int) int {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

func (h *StockHandler) respondWithError(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, Response{
		Success: false,
		Error:   err.Error(),
	})
}

// GetBrokerages godoc
// @Summary Get all brokerages
// @Description Retrieves all brokerage firms
// @Tags brokerages
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/brokerages [get]
func (h *StockHandler) GetBrokerages(c *gin.Context) {
	brokerages, err := h.brokerageUC.GetAll(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get brokerages", zap.Error(err))
		h.respondWithError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    brokerages,
	})
}

// GetBrokerageByID godoc
// @Summary Get a brokerage by ID
// @Description Retrieves a single brokerage firm by ID
// @Tags brokerages
// @Accept json
// @Produce json
// @Param id path int true "Brokerage ID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/brokerages/{id} [get]
func (h *StockHandler) GetBrokerageByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, errors.New("invalid brokerage ID"))
		return
	}

	brokerage, err := h.brokerageUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			h.respondWithError(c, http.StatusNotFound, err)
			return
		}
		h.logger.Error("Failed to get brokerage", zap.Int64("id", id), zap.Error(err))
		h.respondWithError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    brokerage,
	})
}

// GetActions godoc
// @Summary Get all actions
// @Description Retrieves all analyst actions
// @Tags actions
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/actions [get]
func (h *StockHandler) GetActions(c *gin.Context) {
	actions, err := h.actionUC.GetAll(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get actions", zap.Error(err))
		h.respondWithError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    actions,
	})
}

// GetActionByID godoc
// @Summary Get an action by ID
// @Description Retrieves a single action by ID
// @Tags actions
// @Accept json
// @Produce json
// @Param id path int true "Action ID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/actions/{id} [get]
func (h *StockHandler) GetActionByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, errors.New("invalid action ID"))
		return
	}

	action, err := h.actionUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			h.respondWithError(c, http.StatusNotFound, err)
			return
		}
		h.logger.Error("Failed to get action", zap.Int64("id", id), zap.Error(err))
		h.respondWithError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    action,
	})
}

// GetRatings godoc
// @Summary Get all ratings
// @Description Retrieves all rating terms with brokerage information
// @Tags ratings
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/ratings [get]
func (h *StockHandler) GetRatings(c *gin.Context) {
	ratings, err := h.ratingUC.GetAll(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get ratings", zap.Error(err))
		h.respondWithError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    ratings,
	})
}

// GetRatingByID godoc
// @Summary Get a rating by ID
// @Description Retrieves a single rating by ID
// @Tags ratings
// @Accept json
// @Produce json
// @Param id path int true "Rating ID"
// @Success 200 {object} Response
// @Failure 400 {object} Response
// @Failure 404 {object} Response
// @Failure 500 {object} Response
// @Router /api/v1/ratings/{id} [get]
func (h *StockHandler) GetRatingByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.respondWithError(c, http.StatusBadRequest, errors.New("invalid rating ID"))
		return
	}

	rating, err := h.ratingUC.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			h.respondWithError(c, http.StatusNotFound, err)
			return
		}
		h.logger.Error("Failed to get rating", zap.Int64("id", id), zap.Error(err))
		h.respondWithError(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    rating,
	})
}
