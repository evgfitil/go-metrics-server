package storage

import (
	"context"
	"github.com/stretchr/testify/assert"
	"strconv"
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

func TestMemStorage_GetAllMetrics(t *testing.T) {
	type fields struct {
		metrics map[string]*metrics.Metrics
		mu      *sync.RWMutex
	}
	type args struct {
		in0 context.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]*metrics.Metrics
	}{
		{
			name: "empty storage",
			fields: fields{
				metrics: map[string]*metrics.Metrics{},
				mu:      &sync.RWMutex{},
			},
			args: args{
				in0: context.Background(),
			},
			want: map[string]*metrics.Metrics{},
		},
		{
			name: "mixed metric types",
			fields: fields{
				metrics: map[string]*metrics.Metrics{
					"gaugeMetric": {
						ID:    "gaugeMetric",
						MType: "gauge",
						Value: float64Ptr(123.456),
					},
					"counterMetric": {
						ID:    "counterMetric",
						MType: "counter",
						Delta: int64Ptr(789),
					},
				},
				mu: &sync.RWMutex{},
			},
			args: args{
				in0: context.Background(),
			},
			want: map[string]*metrics.Metrics{
				"gaugeMetric": {
					ID:    "gaugeMetric",
					MType: "gauge",
					Value: float64Ptr(123.456),
				},
				"counterMetric": {
					ID:    "counterMetric",
					MType: "counter",
					Delta: int64Ptr(789),
				},
			},
		},
		{
			name: "concurrent access",
			fields: fields{
				metrics: map[string]*metrics.Metrics{
					"counterMetric": {
						ID:    "counterMetric",
						MType: "counter",
						Delta: int64Ptr(0),
					},
				},
				mu: &sync.RWMutex{},
			},
			args: args{
				in0: context.Background(),
			},
			want: map[string]*metrics.Metrics{
				"counterMetric": {
					ID:    "counterMetric",
					MType: "counter",
					Delta: int64Ptr(100),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				metrics: tt.fields.metrics,
			}

			if tt.name == "concurrent access" {
				var testWG sync.WaitGroup
				numGoroutines := 100
				for i := 0; i < numGoroutines; i++ {
					testWG.Add(1)
					go func() {
						defer testWG.Done()
						metric := &metrics.Metrics{
							ID:    "counterMetric",
							MType: "counter",
							Delta: int64Ptr(1),
						}
						m.Update(tt.args.in0, metric)
					}()
				}
				testWG.Wait()
			}

			got := m.GetAllMetrics(tt.args.in0)
			assert.Equalf(t, tt.want, got, "GetAllMetrics(%v)", tt.args.in0)
		})
	}
}

func generateBatchOfMetrics(size int, metricType string) []*metrics.Metrics {
	batch := make([]*metrics.Metrics, size)
	for i := 0; i < size; i++ {
		id := "metric" + strconv.Itoa(i)
		if metricType == "gauge" {
			batch[i] = &metrics.Metrics{
				ID:    id,
				MType: "gauge",
				Value: float64Ptr(float64(i) + 0.1),
			}
		} else if metricType == "counter" {
			batch[i] = &metrics.Metrics{
				ID:    id,
				MType: "counter",
				Delta: int64Ptr(int64(i) + 1),
			}
		}
	}
	return batch
}

func TestMemStorage_UpdateMetrics(t *testing.T) {
	type fields struct {
		metrics map[string]*metrics.Metrics
	}
	type args struct {
		ctx     context.Context
		metrics []*metrics.Metrics
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		expectedResult map[string]*metrics.Metrics
	}{
		{
			name: "successful add large batch of gauges",
			fields: fields{
				metrics: make(map[string]*metrics.Metrics),
			},
			args: args{
				ctx:     context.Background(),
				metrics: generateBatchOfMetrics(10000, "gauge"),
			},
			expectedResult: func() map[string]*metrics.Metrics {
				expected := make(map[string]*metrics.Metrics)
				for i := 0; i < 10000; i++ {
					id := "metric" + strconv.Itoa(i)
					expected[id] = &metrics.Metrics{
						ID:    id,
						MType: "gauge",
						Value: float64Ptr(float64(i) + 0.1),
					}
				}
				return expected
			}(),
		},
		{
			name: "successful add large batch of counters",
			fields: fields{
				metrics: make(map[string]*metrics.Metrics),
			},
			args: args{
				ctx:     context.Background(),
				metrics: generateBatchOfMetrics(10000, "counter"),
			},
			expectedResult: func() map[string]*metrics.Metrics {
				expected := make(map[string]*metrics.Metrics)
				for i := 0; i < 10000; i++ {
					id := "metric" + strconv.Itoa(i)
					expected[id] = &metrics.Metrics{
						ID:    id,
						MType: "counter",
						Delta: int64Ptr(int64(i) + 1),
					}
				}
				return expected
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				metrics: tt.fields.metrics,
			}
			err := m.UpdateMetrics(tt.args.ctx, tt.args.metrics)
			assert.NoError(t, err, "UpdateMetrics should not return an error")
			assert.Equal(t, tt.expectedResult, m.metrics)
		})
	}
}

func TestMemStorage_UpdateMetricsConcurrent(t *testing.T) {
	m := &MemStorage{
		metrics: map[string]*metrics.Metrics{"testCounter": {ID: "testCounter", MType: "counter", Delta: int64Ptr(0)}},
		mu:      sync.RWMutex{},
	}

	var testWG sync.WaitGroup
	numGoroutines := 100
	metricsBatch := make([]*metrics.Metrics, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		metricsBatch[i] = &metrics.Metrics{ID: "testCounter", MType: "counter", Delta: int64Ptr(1)}
	}

	for i := 0; i < numGoroutines; i++ {
		testWG.Add(1)
		go func(i int) {
			defer testWG.Done()
			m.Update(context.Background(), metricsBatch[i])
		}(i)
	}
	testWG.Wait()

	storedMetric, exists := m.metrics["testCounter"]
	assert.True(t, exists)
	assert.Equal(t, *storedMetric.Delta, *int64Ptr(100))
}
