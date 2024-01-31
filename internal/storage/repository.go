package storage

import "github.com/evgfitil/go-metrics-server.git/internal/metrics"

type Storage interface {
	Update(metric *metrics.Metrics)
	Get(metricName string) (*metrics.Metrics, bool)
	GetAllMetrics() map[string]*metrics.Metrics
	SaveMetrics() error
}
