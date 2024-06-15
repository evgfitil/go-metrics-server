package storage

import (
	"context"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"

	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
)

func float64Ptr(f float64) *float64 {
	return &f
}

func int64Ptr(i int64) *int64 {
	return &i
}

func TestNewMemStorage(t *testing.T) {
	m := NewMemStorage()

	assert.NotNil(t, m)
	assert.NotNil(t, m.metrics)
	assert.Equal(t, 0, len(m.metrics))
}

func TestMemStorage_Get(t *testing.T) {
	type fields struct {
		metrics map[string]*metrics.Metrics
	}
	type args struct {
		in0        context.Context
		metricName string
		in2        string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *metrics.Metrics
		want1  bool
	}{
		{
			name: "get existing metric",
			fields: fields{metrics: map[string]*metrics.Metrics{
				"testCounter": {ID: "testCounter", MType: "counter", Delta: int64Ptr(10)},
			}},
			args: args{
				in0:        context.Background(),
				metricName: "testCounter",
			},
			want:  &metrics.Metrics{ID: "testCounter", MType: "counter", Delta: int64Ptr(10)},
			want1: true,
		},
		{
			name:   "get non-existing metric",
			fields: fields{},
			args:   args{in0: context.Background(), metricName: "testCounter"},
			want:   nil,
			want1:  false,
		},
		{
			name: "get with empty metric name",
			fields: fields{metrics: map[string]*metrics.Metrics{
				"testCounter": {ID: "testCounter", MType: "counter", Delta: int64Ptr(10)},
			}},
			args: args{
				in0:        context.Background(),
				metricName: "",
			},
			want:  nil,
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				metrics: tt.fields.metrics,
			}
			got, got1 := m.Get(tt.args.in0, tt.args.metricName, tt.args.in2)
			assert.Equalf(t, tt.want, got, "Get(%v, %v, %v)", tt.args.in0, tt.args.metricName, tt.args.in2)
			assert.Equalf(t, tt.want1, got1, "Get(%v, %v, %v)", tt.args.in0, tt.args.metricName, tt.args.in2)
		})
	}
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
				metric: &metrics.Metrics{ID: "testGauge", MType: "gauge", Value: float64Ptr(24.6)},
			},
			expectedMetric: &metrics.Metrics{ID: "testGauge", MType: "gauge", Value: float64Ptr(24.6)},
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

func TestMemStorage_UpdateConcurrent(t *testing.T) {
	m := &MemStorage{
		metrics: map[string]*metrics.Metrics{"testCounter": {ID: "testCounter", MType: "counter", Delta: int64Ptr(0)}},
		mu:      sync.RWMutex{},
	}

	var testWG sync.WaitGroup

	for i := 0; i < 100; i++ {
		testWG.Add(1)
		go func() {
			defer testWG.Done()
			m.Update(context.Background(), &metrics.Metrics{ID: "testCounter", MType: "counter", Delta: int64Ptr(1)})
		}()
	}
	testWG.Wait()

	storedMetric, exists := m.metrics["testCounter"]
	assert.True(t, exists)
	assert.Equal(t, *storedMetric.Delta, *int64Ptr(100))
}
