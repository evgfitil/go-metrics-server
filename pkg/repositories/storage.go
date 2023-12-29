package repositories

import "github.com/evgfitil/go-metrics-server.git/internal/metrics"

type Storage interface {
	Update(metric metrics.Metric)
	Get(name string) (metrics.Metric, bool)
}
