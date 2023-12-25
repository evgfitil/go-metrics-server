package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Metric interface {
	GetName() string
}

type Counter struct {
	Name  string
	Value int64
}

func (c Counter) GetName() string {
	return c.Name
}

type Gauge struct {
	Name  string
	Value float64
}

func (g Gauge) GetName() string {
	return g.Name
}

type MemStorage struct {
	metrics map[string]Metric
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]Metric),
	}
}

type Storage interface {
	Update(metric Metric)
	Get(name string) (Metric, bool)
}

func (m *MemStorage) Update(metric Metric) {
	switch v := metric.(type) {
	case Counter:
		if exists, ok := m.metrics[metric.GetName()].(Counter); ok {
			v.Value += exists.Value
		}
		m.metrics[metric.GetName()] = v
	case Gauge:
		m.metrics[metric.GetName()] = metric
	}
}

func (m *MemStorage) Get(name string) (Metric, bool) {
	metric, ok := m.metrics[name]
	return metric, ok
}
func updateCounter(storage Storage, metricName, metricValue string) error {
	value, err := strconv.ParseInt(metricValue, 10, 64)
	if err != nil {
		return err
	}
	metric := Counter{Name: metricName, Value: value}
	storage.Update(metric)
	return nil
}

func updateGauge(storage Storage, metricName, metricValue string) error {
	value, err := strconv.ParseFloat(metricValue, 64)
	if err != nil {
		return err
	}
	metric := Gauge{Name: metricName, Value: value}
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
		metric, ok := storage.Get(metricName)
		if !ok {
			http.Error(res, "Metric not found", http.StatusNotFound)
			return
		}
		fmt.Fprintf(res, "Metric: %s, Value: %v\n", metric.GetName(), metric)
	}
}

func updateMetricsHandler(storage Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// request checking
		switch {
		case req.Method != http.MethodPost:
			http.Error(res, "Invalid request method", http.StatusBadRequest)
			return
		case req.Header.Get("Content-Type") != "text/plain":
			http.Error(res, "Content-Type must be text/plain", http.StatusUnsupportedMediaType)
			return
		}
		// request path checking
		urlParts := strings.Split(req.URL.Path, "/")
		if len(urlParts) != 5 || urlParts[1] != "update" {
			http.Error(res, "Invalid path", http.StatusBadRequest)
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
		fmt.Fprintf(res, "Metric updated: %s: %s\n ", metricName, metricValue)
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
