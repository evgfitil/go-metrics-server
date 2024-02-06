package storage

import (
	"context"
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
)

type Storage interface {
	Update(ctx context.Context, metric *metrics.Metrics)
	Get(ctx context.Context, metricName string) (*metrics.Metrics, bool)
	GetAllMetrics(ctx context.Context) map[string]*metrics.Metrics
	SaveMetrics(ctx context.Context) error
	Ping(ctx context.Context) error
}
