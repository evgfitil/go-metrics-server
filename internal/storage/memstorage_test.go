package storage

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
)

func float64Ptr(f float64) *float64 {
	return &f
}

func int64Ptr(i int64) *int64 {
	return &i
}

func TestMemStorage_Update(t *testing.T) {
	type fields struct {
		metrics map[string]*metrics.Metrics
	}
	type args struct {
		in0    context.Context
		metric *metrics.Metrics
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		expectedMetric *metrics.Metrics
	}{
		{
			name:   "successful add gauge",
			fields: fields{metrics: make(map[string]*metrics.Metrics)},
			args: args{
				in0:    context.Background(),
				metric: &metrics.Metrics{ID: "testGauge", MType: "gauge", Value: float64Ptr(46.4)},
			},
			expectedMetric: &metrics.Metrics{ID: "testGauge", MType: "gauge", Value: float64Ptr(46.4)},
		},
		{
			name:   "successful add counter",
			fields: fields{metrics: make(map[string]*metrics.Metrics)},
			args: args{
				in0:    context.Background(),
				metric: &metrics.Metrics{ID: "testCounter", MType: "counter", Delta: int64Ptr(2)},
			},
			expectedMetric: &metrics.Metrics{ID: "testCounter", MType: "counter", Delta: int64Ptr(2)},
		},
		{
			name:   "successful update existing counter",
			fields: fields{metrics: map[string]*metrics.Metrics{"testCounter": {ID: "testCounter", MType: "counter", Delta: int64Ptr(2)}}},
			args: args{
				in0:    context.Background(),
				metric: &metrics.Metrics{ID: "testCounter", MType: "counter", Delta: int64Ptr(1)},
			},
			expectedMetric: &metrics.Metrics{ID: "testCounter", MType: "counter", Delta: int64Ptr(3)},
		},
		{
			name: "successful update existing gauge",
			fields: fields{metrics: map[string]*metrics.Metrics{
				"testGauge": {ID: "testGauge", MType: "gauge", Value: float64Ptr(46.6)}}},
			args: args{
				in0:    context.Background(),
				metric: &metrics.Metrics{ID: "testGague", MType: "gauge", Value: float64Ptr(24.6)},
			},
			expectedMetric: &metrics.Metrics{ID: "testGague", MType: "gauge", Value: float64Ptr(24.6)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				metrics: tt.fields.metrics,
			}
			m.Update(tt.args.in0, tt.args.metric)
			storedMetric, exists := m.metrics[tt.args.metric.ID]
			assert.True(t, exists)
			assert.Equal(t, tt.expectedMetric, storedMetric)
		})
	}
}
