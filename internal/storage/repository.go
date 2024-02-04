package storage

import (
	"context"
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
)

type Storage interface {
	Update(metric *metrics.Metrics)
	Get(metricName string) (*metrics.Metrics, bool)
	GetAllMetrics() map[string]*metrics.Metrics
	SaveMetrics() error
	Ping(ctx context.Context) error
}
