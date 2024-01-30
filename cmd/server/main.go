package main

import (
	"github.com/evgfitil/go-metrics-server.git/internal/handlers"
	"github.com/evgfitil/go-metrics-server.git/internal/logger"
	"github.com/evgfitil/go-metrics-server.git/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func MetricsRouter(s *storage.MemStorage) chi.Router {
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
	return r
}

func main() {
	logger.InitLogger()
	defer logger.Sugar.Sync()
	if err := Execute(); err != nil {
		logger.Sugar.Fatalf("error starting server: %v", err)
	}
}
