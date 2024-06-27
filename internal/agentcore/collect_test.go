package agentcore

import (
	"runtime"
	"testing"

	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
)

func TestCollectMetrics(t *testing.T) {
	type args struct {
		m *runtime.MemStats
	}
	tests := []struct {
		name string
		args args
		want []MetricInterface
	}{
		{
			name: "Collect basic metrics",
			args: args{
				m: &runtime.MemStats{
					Alloc:       1024,
					BuckHashSys: 2048,
					Frees:       100,
					GCSys:       500,
				},
			},
			want: []MetricInterface{
				metrics.NewGauge("Alloc", 1024),
				metrics.NewGauge("BuckHashSys", 2048),
				metrics.NewGauge("Frees", 100),
				metrics.NewGauge("GCSys", 500),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CollectMetrics(tt.args.m)
			for _, expectedMetric := range tt.want {
				found := false
				for _, actualMetric := range got {
					if actualMetric.GetName() == expectedMetric.GetName() &&
						actualMetric.GetType() == expectedMetric.GetType() {
						expectedValue, _ := expectedMetric.GetValueAsString()
						actualValue, _ := actualMetric.GetValueAsString()
						if expectedValue == actualValue {
							found = true
							break
						}
					}
				}
				if !found {
					t.Errorf("Expected metric %v not found in collected metrics", expectedMetric)
				}
			}
		})
	}
}
