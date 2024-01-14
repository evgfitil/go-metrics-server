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
	switch metric.Type {
	case "Counter":
		if oldMetric, ok := m.metrics[metric.Name]; ok {
			if oldValue, ok := oldMetric.Value.(int64); ok {
				newMetric := metrics.Metric{
					Name:  metric.Name,
					Type:  metric.Type,
					Value: metric.Value.(int64) + oldValue,
				}
				m.metrics[metric.Name] = newMetric
			}
		} else {
			m.metrics[metric.Name] = metric
		}
	case "Gauge":
		m.metrics[metric.Name] = metric
	}
}

func (m *MemStorage) Get(metricName string) (metrics.Metric, bool) {
	metric, ok := m.metrics[metricName]
	return metric, ok
}

func (m *MemStorage) GetAllMetrics() map[string]metrics.Metric {
	return m.metrics
}
