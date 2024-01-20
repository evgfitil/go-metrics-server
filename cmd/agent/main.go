package main

import (
	"encoding/json"
	"fmt"
	"github.com/evgfitil/go-metrics-server.git/internal/logger"
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
	"github.com/go-resty/resty/v2"
	"log"
	"math/rand"
	"runtime"
	"strconv"
	"time"
)

type MetricInterface interface {
	GetName() string
	GetType() string
	GetValueAsString() (string, error)
}

var _ MetricInterface = (*metrics.Metrics)(nil)

func collectMetrics(m *runtime.MemStats) []MetricInterface {
	collectedMetrics := []MetricInterface{
		metrics.NewGauge("Alloc", float64(m.Alloc)),
		metrics.NewGauge("BuckHashSys", float64(m.BuckHashSys)),
		metrics.NewGauge("Frees", float64(m.Frees)),
		metrics.NewGauge("GCCPUFraction", m.GCCPUFraction),
		metrics.NewGauge("GCSys", float64(m.GCSys)),
		metrics.NewGauge("HeapAlloc", float64(m.HeapAlloc)),
		metrics.NewGauge("HeapIdle", float64(m.HeapIdle)),
		metrics.NewGauge("HeapInuse", float64(m.HeapInuse)),
		metrics.NewGauge("HeapObjects", float64(m.HeapObjects)),
		metrics.NewGauge("HeapReleased", float64(m.HeapReleased)),
		metrics.NewGauge("HeapSys", float64(m.HeapSys)),
		metrics.NewGauge("LastGC", float64(m.LastGC)),
		metrics.NewGauge("Lookups", float64(m.Lookups)),
		metrics.NewGauge("MCacheInuse", float64(m.MCacheInuse)),
		metrics.NewGauge("MCacheSys", float64(m.MCacheSys)),
		metrics.NewGauge("MSpanInuse", float64(m.MSpanInuse)),
		metrics.NewGauge("MSpanSys", float64(m.MSpanSys)),
		metrics.NewGauge("Mallocs", float64(m.Mallocs)),
		metrics.NewGauge("NextGC", float64(m.NextGC)),
		metrics.NewGauge("NumForcedGC", float64(m.NumForcedGC)),
		metrics.NewGauge("NumGC", float64(m.NumGC)),
		metrics.NewGauge("OtherSys", float64(m.OtherSys)),
		metrics.NewGauge("PauseTotalNs", float64(m.PauseTotalNs)),
		metrics.NewGauge("StackInuse", float64(m.StackInuse)),
		metrics.NewGauge("Sys", float64(m.Sys)),
		metrics.NewGauge("TotalAlloc", float64(m.TotalAlloc)),
		metrics.NewGauge("RandomValue", rand.Float64()),
	}

	return collectedMetrics
}

func sendMetrics(metrics []MetricInterface, serverURL string) {
	for _, metric := range metrics {
		sendingMetric, err := json.Marshal(metric)
		if err != nil {
			logger.Sugar.Errorln("error marshaling json: %v", err)
		}
		urlFormat := "/update"
		url := serverURL + urlFormat

		client := resty.New()
		_, err = client.R().
			SetHeader("Content-type", "application/json").
			SetBody(sendingMetric).
			Post(url)

		if err != nil {
			log.Println("error sending metric:", err)
			continue
		}
	}
}

func getIntervalSettings(interval string) (time.Duration, error) {
	i, err := strconv.Atoi(interval)
	if err != nil {
		return 0, fmt.Errorf("invalid interval: %v", err)
	}
	res := time.Duration(i) * time.Second
	return res, nil
}

func main() {
	logger.InitLogger()
	defer logger.Sugar.Sync()
	config := NewConfig()
	err := config.ParseFlags()
	if err != nil {
		log.Fatalf("error getting arguments: %v", err)
	}
	serverURL := "http://" + config.serverAddress
	pollIntervalStr, reportIntervalStr := config.pollInterval, config.reportInterval

	pollInterval, err := getIntervalSettings(pollIntervalStr)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	reportInterval, err := getIntervalSettings(reportIntervalStr)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	var collectedMetrics []MetricInterface
	var pollCount int64
	lastPollTime, lastReportTime := time.Now(), time.Now()

	for {
		now := time.Now()

		if now.Sub(lastPollTime) >= pollInterval {
			pollCount++
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			collectedMetrics = collectMetrics(&m)
			collectedMetrics = append(collectedMetrics, metrics.NewCounter("PollCount", pollCount))

			lastPollTime = now
		}

		if now.Sub(lastReportTime) > reportInterval {
			sendMetrics(collectedMetrics, serverURL)
			collectedMetrics = []MetricInterface{}

			lastReportTime = now
		}

		time.Sleep(1 * time.Second)
	}
}
