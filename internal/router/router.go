package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/truora/stock-api/internal/handler"
	"github.com/truora/stock-api/internal/middleware"
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
	}

	return router
}
