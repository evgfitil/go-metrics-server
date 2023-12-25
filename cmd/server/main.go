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
	m.metrics[metric.GetName()] = metric
}

func (m *MemStorage) Get(name string) (Metric, bool) {
	metric, ok := m.metrics[name]
	return metric, ok
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
		var metric Metric
		switch metricType {
		case "counter":
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(res, "Invalid metric value", http.StatusBadRequest)
				return
			}
			metric = Counter{Name: metricName, Value: value}
		default:
			http.Error(res, "Unsupported metric type", http.StatusBadRequest)
			return
		}

		storage.Update(metric)
		fmt.Fprintf(res, "Metric updated: %s\n", metric.GetName())
	}
}

func main() {
	storage := NewMemStorage()
	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, updateMetricsHandler(storage))

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
