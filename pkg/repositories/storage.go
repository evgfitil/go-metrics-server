package repositories

import "github.com/evgfitil/go-metrics-server.git/internal/metrics"

type Storage interface {
	Update(metric metrics.Metric)
	Get(metricName string) (metrics.Metric, bool)
	GetAllMetrics() map[string]metrics.Metric
}
