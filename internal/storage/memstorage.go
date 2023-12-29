package storage

import "github.com/evgfitil/go-metrics-server.git/internal/metrics"

type MemStorage struct {
	metrics map[string]metrics.Metric
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]metrics.Metric),
	}
}

func (m *MemStorage) Update(metric metrics.Metric) {
	switch v := metric.(type) {
	case metrics.Counter:
		if exists, ok := m.metrics[metric.GetName()].(metrics.Counter); ok {
			v.Value += exists.Value
		}
		m.metrics[metric.GetName()] = v
	case metrics.Gauge:
		m.metrics[metric.GetName()] = v
	}
}

func (m *MemStorage) Get(name string) (metrics.Metric, bool) {
	metric, ok := m.metrics[name]
	return metric, ok
}
