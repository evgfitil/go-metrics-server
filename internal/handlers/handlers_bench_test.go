package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/evgfitil/go-metrics-server.git/internal/storage"
)

func BenchmarkGetAllMetrics(b *testing.B) {
	storage := storage.NewMemStorage()
	router := chi.NewRouter()
	router.Get("/metrics", GetAllMetrics(storage).ServeHTTP)

	request, _ := http.NewRequest("GET", "/metrics", nil)
	recorder := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(recorder, request)
	}
}

func BenchmarkGetMetricsJSON(b *testing.B) {
	storage := storage.NewMemStorage()
	router := chi.NewRouter()
	router.Post("/metrics", GetMetricsJSON(storage).ServeHTTP)

	requestBody := `{"ID":"testCounter","MType":"counter"}`
	request, _ := http.NewRequest("POST", "/metrics", bytes.NewReader([]byte(requestBody)))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(recorder, request)
	}
}

func BenchmarkUpdateMetricsJSON(b *testing.B) {
	storage := storage.NewMemStorage()
	router := chi.NewRouter()
	router.Post("/update/metrics", UpdateMetricsJSON(storage).ServeHTTP)

	requestBody := `{"ID":"testCounter","MType":"counter","Delta":100}`
	request, _ := http.NewRequest("POST", "/update/metrics", bytes.NewReader([]byte(requestBody)))
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(recorder, request)
	}
}

func BenchmarkUpdateMetricsJSONParallel(b *testing.B) {
	storage := storage.NewMemStorage()
	router := chi.NewRouter()
	router.Post("/update/metrics", UpdateMetricsJSON(storage).ServeHTTP)

	numMetrics := 1000
	metrics := make([]string, numMetrics)
	for i := 0; i < numMetrics; i++ {
		if i < numMetrics*2/3 {
			metrics[i] = `{"ID":"testMetric","MType":"gauge","Value":100}`
		} else {
			metrics[i] = fmt.Sprintf(`{"ID":"testMetric%d","MType":"gauge","Value":%f}`, i, float64(i)*0.1)
		}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var count int
		for pb.Next() {
			metricIndex := count % numMetrics
			count++
			requestBody := metrics[metricIndex]

			request, _ := http.NewRequest("POST", "/update/metrics", bytes.NewReader([]byte(requestBody)))
			request.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
		}
	})
}
