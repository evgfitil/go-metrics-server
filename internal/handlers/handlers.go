package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"time"

	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
)

type Storage interface {
	Get(ctx context.Context, metricName, metricType string) (*metrics.Metrics, bool)
	GetAllMetrics(ctx context.Context) map[string]*metrics.Metrics
	Ping(ctx context.Context) error
	Update(ctx context.Context, metric *metrics.Metrics)
	UpdateMetrics(ctx context.Context, metrics []*metrics.Metrics) error
}

const (
	requestTimeout = 1 * time.Second
)

func updateCounter(ctx context.Context, storage Storage, metricName string, metricValue int64) error {
	metric := metrics.Metrics{ID: metricName, MType: "counter", Delta: &metricValue}
	storage.Update(ctx, &metric)
	return nil
}

func updateGauge(ctx context.Context, storage Storage, metricName string, metricValue float64) error {
	metric := metrics.Metrics{ID: metricName, MType: "gauge", Value: &metricValue}
	storage.Update(ctx, &metric)
	return nil
}

func GetAllMetrics(storage Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		requestContext, cancel := context.WithTimeout(req.Context(), requestTimeout)
		defer cancel()
		allMetrics := storage.GetAllMetrics(requestContext)

		if len(allMetrics) == 0 {
			res.WriteHeader(http.StatusOK)
			return
		}

		res.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(res, "<html><body>\n")
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

func GetMetricsJSON(storage Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		requestContext, cancel := context.WithTimeout(req.Context(), requestTimeout)
		defer cancel()

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
		metric, ok := storage.Get(requestContext, metricName, metricType)
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
		requestContext, cancel := context.WithTimeout(req.Context(), requestTimeout)
		defer cancel()

		metricName := chi.URLParam(req, "name")
		metricType := chi.URLParam(req, "type")

		if metricType != "counter" && metricType != "gauge" {
			http.Error(res, "Unsupported metric type", http.StatusNotFound)
			return
		}
		metric, ok := storage.Get(requestContext, metricName, metricType)
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

func UpdateMetricsJSON(storage Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		requestContext, cancel := context.WithTimeout(req.Context(), requestTimeout)
		defer cancel()

		if req.Header.Get("Content-Type") != "application/json" {
			http.Error(res, "Invalid Content-type, expected 'application/json'", http.StatusUnsupportedMediaType)
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
			if err := updateCounter(requestContext, storage, metricName, *metricValue); err != nil {
				http.Error(res, "Error updating counter: "+err.Error(), http.StatusBadRequest)
				return
			}
		case "gauge":
			metricValue := incomingMetric.Value
			if metricValue == nil {
				http.Error(res, "Missing metric value", http.StatusBadRequest)
				return
			}
			if err := updateGauge(requestContext, storage, metricName, *metricValue); err != nil {
				http.Error(res, "Error updating gauge: "+err.Error(), http.StatusBadRequest)
				return
			}
		default:
			http.Error(res, "Unsupported metric type", http.StatusBadRequest)
			return
		}

		updateMetric, ok := storage.Get(requestContext, metricName, metricType)
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

func UpdateMetricsCollection(storage Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		requestContext, cancel := context.WithTimeout(req.Context(), requestTimeout)
		defer cancel()

		var incomingMetrics []*metrics.Metrics

		if err := json.NewDecoder(req.Body).Decode(&incomingMetrics); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		if err := storage.UpdateMetrics(requestContext, incomingMetrics); err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
	}
}

func UpdateMetricsPlain(storage Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		requestContext, cancel := context.WithTimeout(req.Context(), requestTimeout)
		defer cancel()

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
			if err := updateCounter(requestContext, storage, metricName, metricValue); err != nil {
				http.Error(res, "Error updating counter: "+err.Error(), http.StatusBadRequest)
				return
			}
		case "gauge":
			metricValue, err := strconv.ParseFloat(metricValueStr, 64)
			if err != nil {
				http.Error(res, "Error updating gauge: "+err.Error(), http.StatusBadRequest)
				return
			}
			if err := updateGauge(requestContext, storage, metricName, metricValue); err != nil {
				http.Error(res, "Error updating gauge: "+err.Error(), http.StatusBadRequest)
				return
			}
		default:
			http.Error(res, "Unsupported metric type", http.StatusBadRequest)
			return
		}
	}
}

func Ping(storage Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		err := storage.Ping(req.Context())
		if err != nil {
			http.Error(res, "database connection failed", http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
	}
}
