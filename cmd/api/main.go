package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/truora/stock-api/internal/config"
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Info("Request received", zap.String("path", r.URL.Path), zap.String("method", r.Method))
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Stock API - Service is running")
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Info("Server started", zap.String("address", addr))
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("Failed to start server", zap.Error(err))
	}
}
