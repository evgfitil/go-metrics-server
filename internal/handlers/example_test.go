package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"

	"github.com/evgfitil/go-metrics-server.git/internal/handlers"
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
	"github.com/evgfitil/go-metrics-server.git/internal/storage"
)

// ExampleGetAllMetrics demonstrates how to fetch all metrics in HTML format.
// Note: The actual output may differ based on the state of the storage and other factors.
func ExampleGetAllMetrics() {
	// Setup
	storage := storage.NewMemStorage()
	router := chi.NewRouter()
	router.Get("/", handlers.GetAllMetrics(storage))

	// Create a request
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	// Execute the request
	router.ServeHTTP(rr, req)

	// Output the response status code
	fmt.Println(rr.Code)

	// Output:
	// 200
}

// ExampleGetMetricsJSON demonstrates how to fetch a specific metric in JSON format.
// Note: The actual output may differ based on the state of the storage and other factors.
func ExampleGetMetricsJSON() {
	// Setup
	storage := storage.NewMemStorage()
	router := chi.NewRouter()
	router.Post("/value", handlers.GetMetricsJSON(storage))

	// Create a metric
	metric := metrics.NewGauge("example_gauge", 123.45)
	storage.Update(context.Background(), &metric)

	// Create a request with the metric data
	metricData, _ := json.Marshal(metric)
	req, _ := http.NewRequest(http.MethodPost, "/value", bytes.NewBuffer(metricData))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Execute the request
	router.ServeHTTP(rr, req)

	// Output the response
	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String())

	// Output:
	// 200
	// {"id":"example_gauge","type":"gauge","value":123.45}
}

// ExampleUpdateMetricsJSON demonstrates how to update a metric with JSON data.
// Note: The actual output may differ based on the state of the storage and other factors.
func ExampleUpdateMetricsJSON() {
	// Setup
	storage := storage.NewMemStorage()
	router := chi.NewRouter()
	router.Post("/update", handlers.UpdateMetricsJSON(storage))

	// Create a metric
	metric := metrics.NewGauge("example_gauge", 678.9)
	metricData, _ := json.Marshal(metric)

	// Create a request with the metric data
	req, _ := http.NewRequest(http.MethodPost, "/update", bytes.NewBuffer(metricData))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Execute the request
	router.ServeHTTP(rr, req)

	// Output the response
	fmt.Println(rr.Code)
	fmt.Println(rr.Body.String())

	// Output:
	// 200
	// {"id":"example_gauge","type":"gauge","value":678.9}
}

// ExampleUpdateMetricsCollection demonstrates how to update multiple metrics with JSON data.
// Note: The actual output may differ based on the state of the storage and other factors.
func ExampleUpdateMetricsCollection() {
	// Setup
	storage := storage.NewMemStorage()
	router := chi.NewRouter()
	router.Post("/updates", handlers.UpdateMetricsCollection(storage))

	// Create multiple metrics
	metric1 := metrics.NewGauge("example_gauge_1", 111.11)
	metric2 := metrics.NewGauge("example_gauge_2", 222.22)
	metricsList := []*metrics.Metrics{&metric1, &metric2}
	metricsData, _ := json.Marshal(metricsList)

	// Create a request with the metrics data
	req, _ := http.NewRequest(http.MethodPost, "/updates", bytes.NewBuffer(metricsData))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Execute the request
	router.ServeHTTP(rr, req)

	// Output the response status code
	fmt.Println(rr.Code)

	// Output:
	// 200
}

// ExamplePing demonstrates how to use the ping endpoint.
// Note: The actual output may differ based on the state of the storage and other factors.
func ExamplePing() {
	// Setup
	storage := storage.NewMemStorage()
	router := chi.NewRouter()
	router.Get("/ping", handlers.Ping(storage))

	// Create a request
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()

	// Execute the request
	router.ServeHTTP(rr, req)

	// Output the response status code
	fmt.Println(rr.Code)

	// Output:
	// 200
}
