package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logger returns a middleware that logs HTTP requests
func Logger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		statusCode := c.Writer.Status()

		// Log request details
		logger.Info("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)

		// Log errors if any
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				logger.Error("Request error", zap.Error(e))
			}
		}
	}
}

// CORS returns a middleware that adds CORS headers
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
