package storage

import (
	"context"
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
	"sync"
)

type MemStorage struct {
	metrics map[string]*metrics.Metrics
	mu      sync.RWMutex
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]*metrics.Metrics),
	}
}

func (m *MemStorage) Update(_ context.Context, metric *metrics.Metrics) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch metric.MType {
	case "counter":
		if oldMetric, ok := m.metrics[metric.ID]; ok {
			if oldDelta := oldMetric.Delta; oldDelta != nil {
				newDelta := *metric.Delta + *oldDelta
				newMetric := &metrics.Metrics{
					ID:    metric.ID,
					MType: metric.MType,
					Delta: &newDelta,
				}
				m.metrics[metric.ID] = newMetric
			}
		} else {
			m.metrics[metric.ID] = metric
		}
	case "gauge":
		m.metrics[metric.ID] = metric
	}
}

func (m *MemStorage) Get(_ context.Context, metricName string) (*metrics.Metrics, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metric, ok := m.metrics[metricName]
	return metric, ok
}

func (m *MemStorage) GetAllMetrics(_ context.Context) map[string]*metrics.Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.metrics
}

func (m *MemStorage) SaveMetrics(_ context.Context) error {
	return nil
}

func (m *MemStorage) Ping(_ context.Context) error {
	return nil
}
