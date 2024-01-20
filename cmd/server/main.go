package main

import (
	"github.com/evgfitil/go-metrics-server.git/internal/handlers"
	"github.com/evgfitil/go-metrics-server.git/internal/logger"
	"github.com/evgfitil/go-metrics-server.git/internal/storage"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func MetricsRouter(s *storage.MemStorage) chi.Router {
	r := chi.NewRouter()
	r.Get("/", handlers.GetAllMetrics(s))
	r.Route("/value", func(r chi.Router) {
		r.Post("/", handlers.GetMetricsJson(s))
		r.Get("/{type}/{name}", handlers.GetMetricsPlain(s))
	})
	r.Route("/update", func(r chi.Router) {
		r.Post("/", handlers.UpdateMetricsJson(s))
		r.Post("/{type}/{name}/{value}", handlers.UpdateMetricsPlain(s))
	})
	return r
}

func main() {
	logger.InitLogger()
	defer logger.Sugar.Sync()
	s := storage.NewMemStorage()
	config := NewConfig()
	err := config.ParseFlags()
	if err != nil {
		logger.Sugar.Fatalf("error getting arguments: %v", err)
	}
	logger.Sugar.Infoln("starting server")
	err = http.ListenAndServe(config.bindAddress, logger.WithLogging(MetricsRouter(s)))
	if err != nil {
		logger.Sugar.Fatalf("error starting server: %v", err)
	}
}
