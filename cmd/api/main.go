// Package main is the entry point for the Stock API service.
//
//	@title			Stock API
//	@version		1.0
//	@description	API for managing stock data from external sources
//	@host			localhost:8080
//	@BasePath		/
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/truora/stock-api/docs"
	"github.com/truora/stock-api/internal/client"
	"github.com/truora/stock-api/internal/config"
	"github.com/truora/stock-api/internal/handler"
	"github.com/truora/stock-api/internal/repository/cockroachdb"
	"github.com/truora/stock-api/internal/router"
	"github.com/truora/stock-api/internal/usecase"
	"github.com/truora/stock-api/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.NewLogger(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting Stock API service",
		zap.String("env", cfg.Server.Env),
		zap.String("port", cfg.Server.Port))

	// Initialize database connection
	db, err := cockroachdb.NewConnection(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	log.Info("Database connection established")

	// Initialize database schema
	if err := cockroachdb.InitSchema(db); err != nil {
		log.Fatal("Failed to initialize database schema", zap.Error(err))
	}

	log.Info("Database schema initialized")

	// Initialize layers
	stockRepo := cockroachdb.NewStockRepository(db)
	stockAPIClient := client.NewStockAPIClient(&cfg.StockAPI)
	stockUseCase := usecase.NewStockUseCase(stockRepo, stockAPIClient, log)
	stockHandler := handler.NewStockHandler(stockUseCase, log)

	// Setup router
	r := router.SetupRouter(stockHandler, log)

	// Configure HTTP server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:           addr,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start server in a goroutine
	go func() {
		log.Info("Server started", zap.String("address", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown with 30 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Server exited")
}
