// Package main provides the entry point for the metrics server application.
// This server collects, processes, and stores various metrics, offering multiple
// endpoints to interact with the metrics data.
// The server supports in-memory storage, file-based storage, and database storage
// for metrics. Configuration options are available via environment variables or
// command-line flags, allowing flexibility in deployment.
// The server also includes optional pprof support for profiling.

// Configuration settings:
// - ADDRESS: Bind address for the server in the format host:port (e.g., "localhost:8080").
// - DATABASE_DSN: Data Source Name for connecting to a database.
// - ENABLE_PPROF: Enable pprof for profiling if set to true (pprof will be available on localhost:6060).
// - FILE_STORAGE_PATH: Path to the file used for file-based storage of metrics.
// - RESTORE: Whether to restore previously saved metrics from the file.
// - STORE_INTERVAL: Interval in seconds for periodically saving metrics to the file (0 to disable).

package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/evgfitil/go-metrics-server.git/internal/handlers"
	"github.com/evgfitil/go-metrics-server.git/internal/logger"
	"github.com/evgfitil/go-metrics-server.git/internal/storage"
)

func MetricsRouter(s storage.Storage) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Compress(5))
	r.Get("/", handlers.GetAllMetrics(s))
	r.Route("/value", func(r chi.Router) {
		r.Post("/", handlers.GetMetricsJSON(s))
		r.Get("/{type}/{name}", handlers.GetMetricsPlain(s))
	})
	r.Route("/update", func(r chi.Router) {
		r.Post("/", handlers.UpdateMetricsJSON(s))
		r.Post("/{type}/{name}/{value}", handlers.UpdateMetricsPlain(s))
	})
	r.Get("/ping", handlers.Ping(s))
	r.Post("/updates/", handlers.UpdateMetricsCollection(s))
	return r
}

func main() {
	logger.InitLogger()
	defer logger.Sugar.Sync()
	if err := Execute(); err != nil {
		logger.Sugar.Fatalf("error starting server: %v", err)
	}
}
