package handlers

import (
	"fmt"
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
	"github.com/evgfitil/go-metrics-server.git/pkg/repositories"
	"net/http"
	"strconv"
	"strings"
)

func UpdateCounter(storage repositories.Storage, metricName, metricValue string) error {
	value, err := strconv.ParseInt(metricValue, 10, 64)
	if err != nil {
		return err
	}
	metric := metrics.Counter{Name: metricName, Value: value}
	storage.Update(metric)
	return nil
}

func UpdateGauge(storage repositories.Storage, metricName, metricValue string) error {
	value, err := strconv.ParseFloat(metricValue, 64)
	if err != nil {
		return err
	}
	metric := metrics.Gauge{Name: metricName, Value: value}
	storage.Update(metric)
	return nil
}

func GetMetricsHandler(storage repositories.Storage) http.HandlerFunc {
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
		fmt.Fprintln(res, metric)
	}
}

func UpdateMetricsHandler(storage repositories.Storage) http.HandlerFunc {
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
