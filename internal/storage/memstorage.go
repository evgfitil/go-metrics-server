package storage

import "github.com/evgfitil/go-metrics-server.git/internal/metrics"

type MemStorage struct {
	metrics map[string]*metrics.Metrics
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]*metrics.Metrics),
	}
}

func (m *MemStorage) Update(metric *metrics.Metrics) {
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

func (m *MemStorage) Get(metricName string) (*metrics.Metrics, bool) {
	metric, ok := m.metrics[metricName]
	return metric, ok
}

func (m *MemStorage) GetAllMetrics() map[string]*metrics.Metrics {
	return m.metrics
}
