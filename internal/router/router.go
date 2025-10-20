package router

import (
	"github.com/company/stock-api/internal/handler"
	"github.com/company/stock-api/internal/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// SetupRouter configures and returns the HTTP router
func SetupRouter(stockHandler *handler.StockHandler, logger *zap.Logger) *gin.Engine {
	// Set Gin mode based on environment
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())

	// Health check endpoint
	router.GET("/health", stockHandler.HealthCheck)

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		stocks := v1.Group("/stocks")
		{
			stocks.GET("", stockHandler.GetStocks)
			stocks.GET("/:id", stockHandler.GetStockByID)
			stocks.POST("/sync", stockHandler.SyncStocks)
		}

		// Get all historical versions of a stock by ticker
		v1.GET("/stock/:ticker", stockHandler.GetStocksByTicker)

		// Get stock recommendations
		v1.GET("/recommendations", stockHandler.GetRecommendations)

		// Brokerage routes (read-only)
		brokerages := v1.Group("/brokerages")
		{
			brokerages.GET("", stockHandler.GetBrokerages)
			brokerages.GET("/:id", stockHandler.GetBrokerageByID)
		}

		// Action routes (read-only)
		actions := v1.Group("/actions")
		{
			actions.GET("", stockHandler.GetActions)
			actions.GET("/:id", stockHandler.GetActionByID)
		}

		// Rating routes (read-only)
		ratings := v1.Group("/ratings")
		{
			ratings.GET("", stockHandler.GetRatings)
			ratings.GET("/:id", stockHandler.GetRatingByID)
		}
	}

	return router
}
