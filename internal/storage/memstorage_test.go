package storage

import (
	"context"
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemStorage(t *testing.T) {
	storage := NewMemStorage()
	var testGaugeMetricValue float64
	var testCounterMetricValue int64
	testGaugeMetricValue = 46.4
	testCounterMetricValue = 1

	testMetricGauge := metrics.Metrics{ID: "test", MType: "gauge", Value: &testGaugeMetricValue}
	storage.Update(context.TODO(), &testMetricGauge)
	retrievedMetric, ok := storage.Get(context.TODO(), "test", "gauge")
	assert.True(t, ok, "the metric must exists")
	assert.Equal(t, testMetricGauge, *retrievedMetric, "metrics must be equal")

	testMetricCounter := metrics.Metrics{ID: "counter", MType: "counter", Delta: &testCounterMetricValue}
	storage.Update(context.TODO(), &testMetricCounter)
	storage.Update(context.TODO(), &testMetricCounter)

	retrievedCounter, _ := storage.Get(context.TODO(), "counter", "counter")
	assert.Equal(t, int64(2), *retrievedCounter.Delta, "counter value must increment")
}
