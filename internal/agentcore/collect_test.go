package agentcore

import (
	"runtime"
	"testing"
)

func TestCollectMetrics(t *testing.T) {

	type args struct {
		m *runtime.MemStats
	}
	tests := []struct {
		name    string
		args    args
		wantLen int
	}{
		{
			name: "Basic metrics test",
			args: args{
				m: &runtime.MemStats{
					Alloc:         1024,
					TotalAlloc:    2048,
					Sys:           4096,
					Lookups:       10,
					Mallocs:       100,
					Frees:         50,
					HeapAlloc:     1024,
					HeapSys:       2048,
					HeapIdle:      512,
					HeapInuse:     512,
					HeapReleased:  256,
					HeapObjects:   50,
					StackInuse:    128,
					StackSys:      256,
					MSpanInuse:    64,
					MSpanSys:      128,
					MCacheInuse:   32,
					MCacheSys:     64,
					BuckHashSys:   16,
					GCSys:         256,
					OtherSys:      512,
					NextGC:        5000,
					LastGC:        1000,
					PauseTotalNs:  10000,
					NumGC:         5,
					NumForcedGC:   1,
					GCCPUFraction: 0.25,
				},
			},
			wantLen: 28, // 27 metrics from MemStats + 1 random value
		},
		{
			name: "Zero metrics test",
			args: args{
				m: &runtime.MemStats{},
			},
			wantLen: 28,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CollectMetrics(tt.args.m)
			if len(got) != tt.wantLen {
				t.Errorf("CollectMetrics() returned %d metrics, want %d metrics", len(got), tt.wantLen)
			}
			for _, metric := range got {
				valueStr, err := metric.GetValueAsString()
				if err != nil {
					t.Errorf("Error getting value as string for metric %s: %v", metric.GetName(), err)
				}
				if valueStr == "" {
					t.Errorf("Metric %s returned an empty string", metric.GetName())
				}
			}
		})
	}
}
