package main

import (
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
	"net/http"
	"strconv"
	"strings"
)

type MemStorage struct {
	metrics map[string]metrics.Metric
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]metrics.Metric),
	}
}

type Storage interface {
	Update(metric metrics.Metric)
	Get(name string) (metrics.Metric, bool)
}

func (m *MemStorage) Update(metric metrics.Metric) {
	switch v := metric.(type) {
	case metrics.Counter:
		if exists, ok := m.metrics[metric.GetName()].(metrics.Counter); ok {
			v.Value += exists.Value
		}
		m.metrics[metric.GetName()] = v
	case metrics.Gauge:
		m.metrics[metric.GetName()] = metric
	}
}

func (m *MemStorage) Get(name string) (metrics.Metric, bool) {
	metric, ok := m.metrics[name]
	return metric, ok
}
func updateCounter(storage Storage, metricName, metricValue string) error {
	value, err := strconv.ParseInt(metricValue, 10, 64)
	if err != nil {
		return err
	}
	metric := metrics.Counter{Name: metricName, Value: value}
	storage.Update(metric)
	return nil
}

func updateGauge(storage Storage, metricName, metricValue string) error {
	value, err := strconv.ParseFloat(metricValue, 64)
	if err != nil {
		return err
	}
	metric := metrics.Gauge{Name: metricName, Value: value}
	storage.Update(metric)
	return nil
}

func getMetricsHandler(storage Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// request checking
		if req.Method != http.MethodGet {
			http.Error(res, "Invalid request method", http.StatusBadRequest)
			return
		}
		// request path checking
		urlParts := strings.Split(req.URL.Path, "/")
		if len(urlParts) != 3 || urlParts[1] != "get" {
			http.Error(res, "Invalid path", http.StatusBadRequest)
			return
		}
		metricName := urlParts[2]
		_, ok := storage.Get(metricName)
		if !ok {
			http.Error(res, "Metric not found", http.StatusNotFound)
			return
		}
	}
}

func updateMetricsHandler(storage Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// request checking
		if req.Method != http.MethodPost {
			http.Error(res, "Invalid request method", http.StatusBadRequest)
			return
		}
		// request path checking
		urlParts := strings.Split(req.URL.Path, "/")
		if len(urlParts) != 5 || urlParts[1] != "update" {
			http.Error(res, "Invalid path", http.StatusNotFound)
			return
		}

		// update metrics
		metricType, metricName, metricValue := urlParts[2], urlParts[3], urlParts[4]
		switch metricType {
		case "counter":
			if err := updateCounter(storage, metricName, metricValue); err != nil {
				http.Error(res, "Error updating counter: "+err.Error(), http.StatusBadRequest)
				return
			}
		case "gauge":
			if err := updateGauge(storage, metricName, metricValue); err != nil {
				http.Error(res, "Error updating gauge: "+err.Error(), http.StatusBadRequest)
			}
		default:
			http.Error(res, "Unsupported metric type", http.StatusBadRequest)
			return
		}
	}
}

func main() {
	storage := NewMemStorage()
	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, updateMetricsHandler(storage))
	mux.HandleFunc(`/get/`, getMetricsHandler(storage))

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
