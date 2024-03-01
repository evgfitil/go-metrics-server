package agentcore

import (
	"context"
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

type metricsBatch struct {
	batch  []metrics.Metrics
	unique map[string]struct{}
	mu     sync.Mutex
}

func newMetricsBatch() *metricsBatch {
	return &metricsBatch{
		batch:  make([]metrics.Metrics, 0),
		unique: make(map[string]struct{}),
		mu:     sync.Mutex{},
	}
}

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

type sendTask struct {
	metrics   []metrics.Metrics
	serverURL string
	key       string
}

func (mc *metricsCache) Update(metric metrics.Metrics) bool {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	currentMetric, ok := mc.cache[metric.ID]
	switch metric.MType {
	case "gauge":
		if !ok {
			mc.cache[metric.ID] = &metric
			return true
		}
		if currentMetric.Value != nil && metric.Value != nil {
			if *currentMetric.Value != *metric.Value {
				mc.cache[metric.ID] = &metric
				return true
			}
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

func sendMetrics(ctx context.Context, key string, metrics []metrics.Metrics, serverURL string) {
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
			SetContext(ctx).
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

func sendBatchMetrics(ctx context.Context, key string, metrics []metrics.Metrics, serverURL string) {
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
		SetContext(ctx).
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

func runWorkers(ctx context.Context, workersCount int, tasksChan <-chan sendTask, batchMode bool, wg *sync.WaitGroup) {
	for i := 0; i < workersCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case task, ok := <-tasksChan:
					if !ok {
						return
					}
					if batchMode {
						sendBatchMetrics(ctx, task.key, task.metrics, task.serverURL)
					} else {
						sendMetrics(ctx, task.key, task.metrics, task.serverURL)
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}

func splitBatch(batch []metrics.Metrics, workersCount int) [][]metrics.Metrics {
	var batches [][]metrics.Metrics

	batchSize := (len(batch) + workersCount - 1) / workersCount
	for i := 0; i < len(batch); i += batchSize {
		end := i + batchSize
		if end > len(batch) {
			end = len(batch)
		}
		batches = append(batches, batch[i:end])
	}
	return batches
}

func StartSender(ctx context.Context, metricsChan <-chan []metrics.Metrics, serverURL string, reportInterval time.Duration, batchMode bool, key string, rateLimit int) {
	logger.Sugar.Infoln("sender has been started")
	cache := newMetricsCache()
	tasksChan := make(chan sendTask, 100)
	mb := newMetricsBatch()

	var wg sync.WaitGroup
	runWorkers(ctx, rateLimit, tasksChan, batchMode, &wg)

	ticker := time.NewTicker(reportInterval)
	defer ticker.Stop()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for incomingMetrics := range metricsChan {
			for _, metric := range incomingMetrics {
				_, ok := mb.unique[metric.ID]
				mb.mu.Lock()
				if !ok && cache.Update(metric) {
					mb.batch = append(mb.batch, metric)
					mb.unique[metric.ID] = struct{}{}
				}
				mb.mu.Unlock()
			}
			select {
			case <-ctx.Done():
				return
			default:
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				mb.mu.Lock()
				if len(mb.batch) > 0 {
					batches := splitBatch(mb.batch, rateLimit)
					for _, batch := range batches {
						select {
						case tasksChan <- sendTask{metrics: batch, serverURL: serverURL, key: key}:
						case <-ctx.Done():
							mb.mu.Unlock()
							return
						}
					}
					mb.batch = make([]metrics.Metrics, 0)
					mb.unique = make(map[string]struct{})
					cache.ResetCounters()
				}
				mb.mu.Unlock()
			}
		}
	}()

	go func() {
		<-ctx.Done()
		logger.Sugar.Info("shutdown signal received, terminating the sender")
		ticker.Stop()
		close(tasksChan)
	}()
	wg.Wait()
}
