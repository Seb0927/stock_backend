package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/truora/stock-api/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	log, err := logger.NewLogger("info", "console")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting Stock API service")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Info("Request received", zap.String("path", r.URL.Path), zap.String("method", r.Method))
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Stock API - Service is running")
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	port := "8080"
	log.Info("Server started", zap.String("port", port))
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("Failed to start server", zap.Error(err))
	}
}
