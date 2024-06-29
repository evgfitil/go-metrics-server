package storage

import (
	"context"
	"strconv"
	"testing"

	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
)

func BenchmarkUpdate(b *testing.B) {
	storage := NewMemStorage()
	metricValue := 0.5
	metric := metrics.Metrics{ID: "cpu", MType: "gauge", Value: &metricValue}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.Update(context.Background(), &metric)
	}
}

func BenchmarkGet(b *testing.B) {
	storage := NewMemStorage()
	metricValue := 0.5
	metric := metrics.Metrics{ID: "cpu", MType: "gauge", Value: &metricValue}
	storage.Update(context.Background(), &metric)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = storage.Get(context.Background(), "cpu", "gauge")
	}
}

func BenchmarkGetAllMetrics(b *testing.B) {
	storage := NewMemStorage()
	for i := 0; i < 1000; i++ {
		float64Value := float64(i)
		metric := metrics.Metrics{ID: "cpu_" + strconv.Itoa(i), MType: "gauge", Value: &float64Value}
		storage.Update(context.Background(), &metric)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.GetAllMetrics(context.Background())
	}
}

func BenchmarkUpdateMetrics(b *testing.B) {
	storage := NewMemStorage()
	metricsSlice := make([]*metrics.Metrics, 100)
	metricDelta := int64(1)
	for i := range metricsSlice {
		metricsSlice[i] = &metrics.Metrics{ID: "metric_" + strconv.Itoa(i), MType: "counter", Delta: &metricDelta}
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = storage.UpdateMetrics(context.Background(), metricsSlice)
		}
	})
}
