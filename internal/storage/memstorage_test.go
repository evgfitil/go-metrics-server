package storage

import (
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemStorage(t *testing.T) {
	storage := NewMemStorage()

	testMetricGauge := metrics.Metric{Name: "test", Type: "gauge", Value: float64(46.4)}
	storage.Update(testMetricGauge)
	retrievedMetric, ok := storage.Get("test")
	assert.True(t, ok, "the metric must exists")
	assert.Equal(t, testMetricGauge, retrievedMetric, "metrics must be equal")

	testMetricCounter := metrics.Metric{Name: "counter", Type: "counter", Value: int64(1)}
	storage.Update(testMetricCounter)
	storage.Update(testMetricCounter)

	retrievedCounter, _ := storage.Get("counter")
	assert.Equal(t, int64(2), retrievedCounter.Value, "counter value must increment")
}
