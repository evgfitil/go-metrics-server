package handlers

import (
	"fmt"
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
	"github.com/evgfitil/go-metrics-server.git/pkg/repositories"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

func UpdateCounter(storage repositories.Storage, metricName, metricValue string) error {
	value, err := strconv.ParseInt(metricValue, 10, 64)
	if err != nil {
		return err
	}
	metric := metrics.Metric{Name: metricName, Type: "counter", Value: value}
	storage.Update(metric)
	return nil
}

func UpdateGauge(storage repositories.Storage, metricName, metricValue string) error {
	value, err := strconv.ParseFloat(metricValue, 64)
	if err != nil {
		return err
	}
	metric := metrics.Metric{Name: metricName, Type: "gauge", Value: value}
	storage.Update(metric)
	return nil
}

func GetMetrics(storage repositories.Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		// request checking
		if req.Method != http.MethodGet {
			http.Error(res, "Invalid request method", http.StatusBadRequest)
			return
		}
		metricName := chi.URLParam(req, "name")
		metricType := chi.URLParam(req, "type")

		if metricType != "counter" && metricType != "gauge" {
			http.Error(res, "Unsupported metric type", http.StatusNotFound)
			return
		}
		metric, ok := storage.Get(metricName)
		if !ok {
			http.Error(res, "Metric not found", http.StatusNotFound)
			return
		}
		valueStr, err := metric.GetValueAsString()
		if err != nil {
			return
		}
		fmt.Fprintln(res, valueStr)
	}
}

func UpdateMetrics(storage repositories.Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(res, "Invalid request method", http.StatusBadRequest)
			return
		}

		metricType := chi.URLParam(req, "type")
		metricName := chi.URLParam(req, "name")
		metricValue := chi.URLParam(req, "value")

		switch metricType {
		case "counter":
			if err := UpdateCounter(storage, metricName, metricValue); err != nil {
				http.Error(res, "Error updating counter: "+err.Error(), http.StatusBadRequest)
				return
			}
		case "gauge":
			if err := UpdateGauge(storage, metricName, metricValue); err != nil {
				http.Error(res, "Error updating gauge: "+err.Error(), http.StatusBadRequest)
			}
		default:
			http.Error(res, "Unsupported metric type", http.StatusBadRequest)
			return
		}
	}
}

func GetAllMetrics(storage repositories.Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(res, "<html><body>\n")
		metrics := storage.GetAllMetrics()
		for _, metric := range metrics {
			valueStr, err := metric.GetValueAsString()
			if err != nil {
				fmt.Fprintf(res, "<div>Error getting value for metric %s: %s</div>\n", metric.Name, err)
			} else {
				fmt.Fprintf(res, "<div>%s: %s</div>\n", metric.Name, valueStr)
			}
		}
		fmt.Fprintf(res, "</body></html>\n")
	}
}
