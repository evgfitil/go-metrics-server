package metrics

import (
	"fmt"
	"reflect"
	"testing"
)

func TestMetrics_GetName(t *testing.T) {
	tests := []struct {
		name string
		m    Metrics
		want string
	}{
		{name: "Get name", m: Metrics{ID: "testMetric"}, want: "testMetric"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.GetName(); got != tt.want {
				t.Errorf("GetName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetrics_GetType(t *testing.T) {
	tests := []struct {
		name string
		m    Metrics
		want string
	}{
		{name: "Get type", m: Metrics{MType: "gauge"}, want: "gauge"},
		{name: "Get type counter", m: Metrics{MType: "counter"}, want: "counter"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.GetType(); got != tt.want {
				t.Errorf("GetType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetrics_GetValueAsString(t *testing.T) {
	gaugeValue := 42.42
	counterValue := int64(12345)

	tests := []struct {
		name    string
		m       Metrics
		want    string
		wantErr bool
	}{
		{
			name:    "Gauge value",
			m:       Metrics{MType: "gauge", Value: &gaugeValue},
			want:    fmt.Sprintf("%g", gaugeValue),
			wantErr: false,
		},
		{
			name:    "Counter value",
			m:       Metrics{MType: "counter", Delta: &counterValue},
			want:    fmt.Sprintf("%d", counterValue),
			wantErr: false,
		},
		{
			name:    "Unsupported type",
			m:       Metrics{MType: "unsupported"},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.GetValueAsString()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetValueAsString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetValueAsString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewGauge(t *testing.T) {
	tests := []struct {
		name  string
		id    string
		value float64
		want  Metrics
	}{
		{name: "New gauge", id: "testGauge", value: 42.42, want: Metrics{ID: "testGauge", MType: "gauge", Value: func() *float64 { v := 42.42; return &v }()}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewGauge(tt.id, tt.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGauge() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewCounter(t *testing.T) {
	tests := []struct {
		name  string
		id    string
		value int64
		want  Metrics
	}{
		{name: "New counter", id: "testCounter", value: 12345, want: Metrics{ID: "testCounter", MType: "counter", Delta: func() *int64 { v := int64(12345); return &v }()}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCounter(tt.id, tt.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCounter() = %v, want %v", got, tt.want)
			}
		})
	}
}
