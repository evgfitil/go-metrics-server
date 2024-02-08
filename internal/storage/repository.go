package storage

import (
	"context"
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
)

type Storage interface {
	Get(ctx context.Context, metricName, metricType string) (*metrics.Metrics, bool)
	GetAllMetrics(ctx context.Context) map[string]*metrics.Metrics
	Ping(ctx context.Context) error
	Update(ctx context.Context, metric *metrics.Metrics)
	UpdateMetrics(ctx context.Context, metrics []*metrics.Metrics) error
	SaveMetrics(ctx context.Context) error
}
