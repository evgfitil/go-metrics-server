package main

import (
	"github.com/evgfitil/go-metrics-server.git/internal/handlers"
	"github.com/evgfitil/go-metrics-server.git/internal/storage"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func MetricsRouter(s *storage.MemStorage) chi.Router {
	r := chi.NewRouter()
	r.Get("/", handlers.GetAllMetrics(s))
	r.Get("/value/{type}/{name}", handlers.GetMetrics(s))
	r.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", handlers.UpdateMetrics(s))
	})
	return r
}

func main() {
	s := storage.NewMemStorage()
	config := NewConfig()
	err := config.ParseFlags()
	if err != nil {
		log.Fatalf("error getting arguments: %v", err)
	}
	http.ListenAndServe(config.bindAddress, MetricsRouter(s))
}
