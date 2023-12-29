package main

import (
	"github.com/evgfitil/go-metrics-server.git/internal/handlers"
	"github.com/evgfitil/go-metrics-server.git/internal/storage"
	"net/http"
)

func main() {
	storage := storage.NewMemStorage()
	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, handlers.UpdateMetricsHandler(storage))
	mux.HandleFunc(`/get/`, handlers.GetMetricsHandler(storage))

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
