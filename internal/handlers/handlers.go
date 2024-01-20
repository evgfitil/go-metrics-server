package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type Storage interface {
	Update(metric metrics.Metrics)
	Get(metricName string) (metrics.Metrics, bool)
	GetAllMetrics() map[string]metrics.Metrics
}

func updateCounter(storage Storage, metricName string, metricValue int64) error {
	metric := metrics.Metrics{ID: metricName, MType: "counter", Delta: &metricValue}
	storage.Update(metric)
	return nil
}

func updateGauge(storage Storage, metricName string, metricValue float64) error {
	metric := metrics.Metrics{ID: metricName, MType: "gauge", Value: &metricValue}
	storage.Update(metric)
	return nil
}

func GetMetricsJson(storage Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(res, "Invalid request method", http.StatusBadRequest)
			return
		}
		var requestMetric metrics.Metrics

		if err := json.NewDecoder(req.Body).Decode(&requestMetric); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		metricName := requestMetric.ID
		metricType := requestMetric.MType

		if metricType != "counter" && metricType != "gauge" {
			http.Error(res, "Unsupported metric type", http.StatusNotFound)
			return
		}
		metric, ok := storage.Get(metricName)
		if !ok {
			http.Error(res, "Metric not found", http.StatusNotFound)
			return
		}

		jsonResponse, err := json.Marshal(metric)
		if err != nil {
			http.Error(res, "Error marshaling json", http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.Write(jsonResponse)
	}
}

func GetMetricsPlain(storage Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

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

func UpdateMetricsJson(storage Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Content-Type") != "application/json" {
			http.Error(res, "Invalid Content-type, expected 'application/json'", http.StatusUnsupportedMediaType)
		}

		if req.Method != http.MethodPost {
			http.Error(res, "Invalid request method", http.StatusBadRequest)
			return
		}

		var incomingMetric metrics.Metrics

		if err := json.NewDecoder(req.Body).Decode(&incomingMetric); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		metricType := incomingMetric.MType
		metricName := incomingMetric.ID

		switch metricType {
		case "counter":
			metricValue := incomingMetric.Delta
			if metricValue == nil {
				http.Error(res, "Missing metric value", http.StatusBadRequest)
				return
			}
			if err := updateCounter(storage, metricName, *metricValue); err != nil {
				http.Error(res, "Error updating counter: "+err.Error(), http.StatusBadRequest)
				return
			}
		case "gauge":
			metricValue := incomingMetric.Value
			if metricValue == nil {
				http.Error(res, "Missing metric value", http.StatusBadRequest)
				return
			}
			if err := updateGauge(storage, metricName, *metricValue); err != nil {
				http.Error(res, "Error updating gauge: "+err.Error(), http.StatusBadRequest)
				return
			}
		default:
			http.Error(res, "Unsupported metric type", http.StatusBadRequest)
			return
		}

		updateMetric, ok := storage.Get(metricName)
		if !ok {
			http.Error(res, "Error retrieving updated metric", http.StatusInternalServerError)
			return
		}

		jsonResponse, err := json.Marshal(updateMetric)
		if err != nil {
			http.Error(res, "Error marshaling JSON", http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.Write(jsonResponse)
	}
}

func UpdateMetricsPlain(storage Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(res, "Invalid request method", http.StatusBadRequest)
			return
		}

		metricType := chi.URLParam(req, "type")
		metricName := chi.URLParam(req, "name")
		metricValueStr := chi.URLParam(req, "value")

		switch metricType {
		case "counter":
			metricValue, err := strconv.ParseInt(metricValueStr, 10, 64)
			if err != nil {
				http.Error(res, "Error updating counter: "+err.Error(), http.StatusBadRequest)
				return
			}
			if err := updateCounter(storage, metricName, metricValue); err != nil {
				http.Error(res, "Error updating counter: "+err.Error(), http.StatusBadRequest)
				return
			}
		case "gauge":
			metricValue, err := strconv.ParseFloat(metricValueStr, 64)
			if err != nil {
				http.Error(res, "Error updating gauge: "+err.Error(), http.StatusBadRequest)
				return
			}
			if err := updateGauge(storage, metricName, metricValue); err != nil {
				http.Error(res, "Error updating gauge: "+err.Error(), http.StatusBadRequest)
				return
			}
		default:
			http.Error(res, "Unsupported metric type", http.StatusBadRequest)
			return
		}
	}
}

func GetAllMetrics(storage Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(res, "<html><body>\n")
		allMetrics := storage.GetAllMetrics()
		for _, metric := range allMetrics {
			valueStr, err := metric.GetValueAsString()
			if err != nil {
				fmt.Fprintf(res, "<div>Error getting value for metric %s: %s</div>\n", metric.ID, err)
			} else {
				fmt.Fprintf(res, "<div>%s: %s</div>\n", metric.ID, valueStr)
			}
		}
		fmt.Fprintf(res, "</body></html>\n")
	}
}
