package main

import (
	"fmt"
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
	"github.com/go-resty/resty/v2"
	"log"
	"math/rand"
	"runtime"
	"strconv"
	"time"
)

func collectMetrics(m *runtime.MemStats) []metrics.Metric {
	collectedMetrics := []metrics.Metric{
		metrics.Gauge{Name: "Alloc", Value: float64(m.Alloc)},
		metrics.Gauge{Name: "BuckHashSys", Value: float64(m.BuckHashSys)},
		metrics.Gauge{Name: "Frees", Value: float64(m.Frees)},
		metrics.Gauge{Name: "GCCPUFraction", Value: m.GCCPUFraction},
		metrics.Gauge{Name: "GCSys", Value: float64(m.GCSys)},
		metrics.Gauge{Name: "HeapAlloc", Value: float64(m.HeapAlloc)},
		metrics.Gauge{Name: "HeapIdle", Value: float64(m.HeapIdle)},
		metrics.Gauge{Name: "HeapInuse", Value: float64(m.HeapInuse)},
		metrics.Gauge{Name: "HeapObjects", Value: float64(m.HeapObjects)},
		metrics.Gauge{Name: "HeapReleased", Value: float64(m.HeapReleased)},
		metrics.Gauge{Name: "HeapSys", Value: float64(m.HeapSys)},
		metrics.Gauge{Name: "LastGC", Value: float64(m.LastGC)},
		metrics.Gauge{Name: "Lookups", Value: float64(m.Lookups)},
		metrics.Gauge{Name: "MCacheInuse", Value: float64(m.MCacheInuse)},
		metrics.Gauge{Name: "MCacheSys", Value: float64(m.MCacheSys)},
		metrics.Gauge{Name: "MSpanInuse", Value: float64(m.MSpanInuse)},
		metrics.Gauge{Name: "MSpanSys", Value: float64(m.MSpanSys)},
		metrics.Gauge{Name: "Mallocs", Value: float64(m.Mallocs)},
		metrics.Gauge{Name: "NextGC", Value: float64(m.NextGC)},
		metrics.Gauge{Name: "NumForcedGC", Value: float64(m.NumForcedGC)},
		metrics.Gauge{Name: "NumGC", Value: float64(m.NumGC)},
		metrics.Gauge{Name: "OtherSys", Value: float64(m.OtherSys)},
		metrics.Gauge{Name: "PauseTotalNs", Value: float64(m.PauseTotalNs)},
		metrics.Gauge{Name: "StackInuse", Value: float64(m.StackInuse)},
		metrics.Gauge{Name: "Sys", Value: float64(m.Sys)},
		metrics.Gauge{Name: "TotalAlloc", Value: float64(m.TotalAlloc)},
		metrics.Gauge{Name: "RandomValue", Value: rand.Float64()},
	}

	return collectedMetrics
}

func sendMetrics(m []metrics.Metric, serverURL string) {
	for _, metric := range m {
		var metricType string
		switch metric.(type) {
		case metrics.Gauge:
			metricType = "gauge"
		case metrics.Counter:
			metricType = "counter"
		}
		urlFormat := "%s/update/{metricType}/{metricName}/{metricValue}"
		url := fmt.Sprintf(urlFormat, serverURL)

		client := resty.New()
		_, err := client.R().
			SetPathParams(map[string]string{
				"metricType":  metricType,
				"metricName":  metric.GetName(),
				"metricValue": metric.GetValueAsString(),
			}).Post(url)

		if err != nil {
			log.Println("error sending metric:", err)
			continue
		}
	}
}

func getIntervalSettings(interval string) (*time.Ticker, error) {
	i, err := strconv.Atoi(interval)
	if err != nil {
		return nil, fmt.Errorf("invalid interval: %v", err)
	}
	res := time.NewTicker(time.Duration(i) * time.Second)
	return res, nil
}

func main() {
	var pollCount int64
	args, err := ParseFlags()
	if err != nil {
		log.Fatalf("error getting arguments: %v", err)
	}
	serverAddr := args["addr"]
	serverURL := "http://" + serverAddr
	pollIntervalStr, reportIntervalStr := args["pollInterval"], args["reportInterval"]

	pollTicker, err := getIntervalSettings(pollIntervalStr)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	reportTicker, err := getIntervalSettings(reportIntervalStr)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	defer pollTicker.Stop()
	defer reportTicker.Stop()

	var collectedMetrics []metrics.Metric
	for {
		select {
		case <-pollTicker.C:
			pollCount++
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			collectedMetrics = collectMetrics(&m)
			collectedMetrics = append(collectedMetrics, metrics.Counter{Name: "PollCount", Value: pollCount})
		case <-reportTicker.C:
			sendMetrics(collectedMetrics, serverURL)
			collectedMetrics = []metrics.Metric{}
		}
	}
}
