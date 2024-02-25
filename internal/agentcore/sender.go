package agentcore

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/evgfitil/go-metrics-server.git/internal/logger"
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
	"github.com/go-resty/resty/v2"
	"sync"
	"time"
)

const (
	retryCount       = 3
	retryWait        = 2 * time.Second
	retryMaxWaitTime = 5 * time.Second
)

type metricsCache struct {
	cache map[string]*metrics.Metrics
	mu    sync.Mutex
}

func newMetricsCache() *metricsCache {
	return &metricsCache{
		cache: make(map[string]*metrics.Metrics),
		mu:    sync.Mutex{},
	}
}

func (mc *metricsCache) Update(metric metrics.Metrics) bool {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	currentMetric, ok := mc.cache[metric.ID]
	switch metric.MType {
	case "gauge":
		if !ok || currentMetric.Value != metric.Value {
			mc.cache[metric.ID] = &metric
			return true
		}
	case "counter":
		if !ok {
			mc.cache[metric.ID] = &metric
			return true
		} else {
			if metric.Delta != nil {
				if currentMetric.Delta == nil {
					currentMetric.Delta = new(int64)
				}
				*currentMetric.Delta += *metric.Delta
				return true
			}
		}
	}
	return false
}

func (mc *metricsCache) ResetCounters() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	for _, metric := range mc.cache {
		if metric.MType == "counter" {
			*metric.Delta = 0
		}
	}
}

func SendMetrics(key string, metrics []metrics.Metrics, serverURL string) {
	for _, metric := range metrics {
		sendingMetric, err := json.Marshal(metric)
		if err != nil {
			logger.Sugar.Errorln("error marshaling json: %v", err)
		}
		urlFormat := "/update/"
		url := serverURL + urlFormat

		client := resty.New()
		client.
			SetRetryCount(retryCount).
			SetRetryWaitTime(retryWait).
			SetRetryMaxWaitTime(retryMaxWaitTime)
		req := client.R().
			SetHeader("Content-type", "application/json").
			SetBody(sendingMetric)
		if key != "" {
			hash := computeHash(key, sendingMetric)
			req.SetHeader("HashSHA256", hash)
		}
		_, err = req.Post(url)
		if err != nil {
			logger.Sugar.Errorln("error sending metric: %v", err)
			continue
		}
	}
}

func SendBatchMetrics(key string, metrics []metrics.Metrics, serverURL string) {
	sendingMetrics, err := json.Marshal(metrics)
	if err != nil {
		logger.Sugar.Errorln("error marshaling json: %v", err)
	}
	urlFormat := "/updates/"
	url := serverURL + urlFormat

	client := resty.New()
	client.
		SetRetryCount(retryCount).
		SetRetryWaitTime(retryWait).
		SetRetryMaxWaitTime(retryMaxWaitTime)
	req := client.R().
		SetHeader("Content-type", "application/json").
		SetBody(sendingMetrics)
	if key != "" {
		hash := computeHash(key, sendingMetrics)
		req.SetHeader("HashSHA256", hash)
	}
	_, err = req.Post(url)

	if err != nil {
		logger.Sugar.Errorf("error sending metrics: %v", err)
	}
}

func computeHash(key string, data []byte) string {
	secretKey := []byte(key)
	h := hmac.New(sha256.New, secretKey)
	h.Write(data)
	dst := h.Sum(nil)
	return fmt.Sprintf("%x", dst)
}

func StartSender(metricsChan <-chan []metrics.Metrics, serverURL string, reportInterval time.Duration, batchMode bool, key string, rateLimit int) {
	cache := newMetricsCache()
	var metricsBatch []metrics.Metrics
	uniqueIDs := make(map[string]struct{})

	ticker := time.NewTicker(reportInterval)
	defer ticker.Stop()

	for {
		select {
		case incomingMetrics := <-metricsChan:
			for _, metric := range incomingMetrics {
				_, ok := uniqueIDs[metric.ID]
				if !ok && cache.Update(metric) {
					metricsBatch = append(metricsBatch, metric)
					uniqueIDs[metric.ID] = struct{}{}
				}
			}
		case <-ticker.C:
			if batchMode && len(metricsBatch) > 0 {
				SendBatchMetrics(key, metricsBatch, serverURL)
			}
			if !batchMode && len(metricsBatch) > 0 {
				SendMetrics(key, metricsBatch, serverURL)
			}
			cache.ResetCounters()
			metricsBatch = []metrics.Metrics{}
			uniqueIDs = make(map[string]struct{})
		}
	}
}
