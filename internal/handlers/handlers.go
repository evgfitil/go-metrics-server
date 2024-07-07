// Package handlers provides HTTP handlers for the server to receive and process
// metrics from agents. It includes functionalities to parse incoming requests,
// store metrics, and return responses.
package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/evgfitil/go-metrics-server.git/internal/logger"
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
)

// Storage defines the interface for a metrics storage system.
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

var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func decodeJSON(data []byte, v interface{}) error {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	buf.Write(data)
	return json.NewDecoder(buf).Decode(v)
}

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

// GetAllMetrics returns an HTTP handler that responds with all metrics in HTML format.
func GetAllMetrics(storage Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		requestContext, cancel := context.WithTimeout(req.Context(), requestTimeout)
		defer cancel()

		allMetrics := storage.GetAllMetrics(requestContext)

		res.Header().Set("Content-Type", "text/html; charset=utf-8")

		_, err := fmt.Fprintf(res, "<html><body>\n")
		if err != nil {
			logger.Sugar.Errorf("Error writing initial response: %v", err)
			http.Error(res, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if len(allMetrics) == 0 {
			res.WriteHeader(http.StatusOK)
			_, err := fmt.Fprintf(res, "<div>No metrics available</div>\n</body></html>")
			if err != nil {
				logger.Sugar.Errorf("Error writing no metrics message to response: %v", err)
				http.Error(res, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		for _, metric := range allMetrics {
			valueStr, err := metric.GetValueAsString()
			if err != nil {
				_, err := fmt.Fprintf(res, "<div>Error getting value for metric %s: %s</div>\n", metric.ID, err)
				if err != nil {
					logger.Sugar.Errorf("Error writing metric error to response: %v", err)
					http.Error(res, "Internal Server Error", http.StatusInternalServerError)
					return
				}
			} else {
				_, err := fmt.Fprintf(res, "<div>%s: %s</div>\n", metric.ID, valueStr)
				if err != nil {
					logger.Sugar.Errorf("Error writing metric to response: %v", err)
					http.Error(res, "Internal Server Error", http.StatusInternalServerError)
					return
				}
			}
		}

		_, err = fmt.Fprintf(res, "</body></html>\n")
		if err != nil {
			logger.Sugar.Errorf("Error finalizing response: %v", err)
			http.Error(res, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// GetMetricsJSON returns an HTTP handler that responds with a specific metric in JSON format.
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
		_, err = res.Write(jsonResponse)
		if err != nil {
			logger.Sugar.Errorf("Error writing JSON response: %v", err)
			http.Error(res, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

// GetMetricsPlain returns an HTTP handler that responds with a specific metric in plain text format.
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
			logger.Sugar.Errorf("error getting a string value: %v", err)
			return
		}
		_, err = fmt.Fprintln(res, valueStr)
		if err != nil {
			logger.Sugar.Errorf("(error writing value to response: %v", err)
			return
		}
	}
}

// UpdateMetricsJSON returns an HTTP handler that updates a metric with data from a JSON request body.
func UpdateMetricsJSON(storage Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		requestContext, cancel := context.WithTimeout(req.Context(), requestTimeout)
		defer cancel()

		if req.Header.Get("Content-Type") != "application/json" {
			http.Error(res, "Invalid Content-type, expected 'application/json'", http.StatusUnsupportedMediaType)
		}

		var incomingMetric metrics.Metrics

		buf := bufferPool.Get().(*bytes.Buffer)
		buf.Reset()
		defer bufferPool.Put(buf)

		if _, err := buf.ReadFrom(req.Body); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		if err := decodeJSON(buf.Bytes(), &incomingMetric); err != nil {
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
		_, err = res.Write(jsonResponse)
		if err != nil {
			logger.Sugar.Errorf("Error writing JSON response: %v", err)
			http.Error(res, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}
}

// UpdateMetricsCollection returns an HTTP handler that updates multiple metrics with data from a JSON request body.
func UpdateMetricsCollection(storage Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		requestContext, cancel := context.WithTimeout(req.Context(), requestTimeout)
		defer cancel()

		var incomingMetrics []*metrics.Metrics

		if err := json.NewDecoder(req.Body).Decode(&incomingMetrics); err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}

		if len(incomingMetrics) == 0 {
			http.Error(res, "empty input", http.StatusOK)
			return
		}
		if err := storage.UpdateMetrics(requestContext, incomingMetrics); err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
		}
	}
}

// UpdateMetricsPlain returns an HTTP handler that updates a metric with data from URL parameters.
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

// Ping returns an HTTP handler that checks the connectivity to the storage system.
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
