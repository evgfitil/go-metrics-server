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
	switch metric.MType {
	case "counter":
		if oldMetric, ok := m.metrics[metric.ID]; ok {
			if oldValue, ok := oldMetric.Value.(int64); ok {
				newMetric := metrics.Metric{
					ID:    metric.ID,
					MType: metric.MType,
					Value: metric.Value.(int64) + oldValue,
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

func (m *MemStorage) Get(metricName string) (metrics.Metric, bool) {
	metric, ok := m.metrics[metricName]
	return metric, ok
}

func (m *MemStorage) GetAllMetrics() map[string]metrics.Metric {
	return m.metrics
}
