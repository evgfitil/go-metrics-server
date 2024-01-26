package storage

import (
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemStorage(t *testing.T) {
	storage := NewMemStorage(nil)
	var testGaugeMetricValue float64
	var testCounterMetricValue int64
	testGaugeMetricValue = 46.4
	testCounterMetricValue = 1

	testMetricGauge := metrics.Metrics{ID: "test", MType: "gauge", Value: &testGaugeMetricValue}
	storage.Update(&testMetricGauge)
	retrievedMetric, ok := storage.Get("test")
	assert.True(t, ok, "the metric must exists")
	assert.Equal(t, testMetricGauge, retrievedMetric, "metrics must be equal")

	testMetricCounter := metrics.Metrics{ID: "counter", MType: "counter", Delta: &testCounterMetricValue}
	storage.Update(&testMetricCounter)
	storage.Update(&testMetricCounter)

	retrievedCounter, _ := storage.Get("counter")
	assert.Equal(t, int64(2), *retrievedCounter.Delta, "counter value must increment")
}
