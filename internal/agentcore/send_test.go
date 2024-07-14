package agentcore

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
)

func TestSendMetrics(t *testing.T) {
	type args struct {
		metrics   []MetricInterface
		serverURL string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Send single metric",
			args: args{
				metrics: []MetricInterface{
					metrics.NewGauge("Alloc", 123.45),
				},
				serverURL: "",
			},
		},
		{
			name: "Send multiple metrics",
			args: args{
				metrics: []MetricInterface{
					metrics.NewGauge("Alloc", 123.45),
					metrics.NewGauge("TotalAlloc", 678.90),
				},
				serverURL: "",
			},
		},
		{
			name: "Send empty metrics list",
			args: args{
				metrics:   []MetricInterface{},
				serverURL: "",
			},
		},
		{
			name: "Retry on failure",
			args: args{
				metrics: []MetricInterface{
					metrics.NewGauge("Alloc", 123.45),
				},
				serverURL: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Retry on failure" {
				var retries int
				mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if retries < retryCount {
						http.Error(w, "temporary error", http.StatusInternalServerError)
						retries++
					} else {
						w.WriteHeader(http.StatusOK)
					}
				}))
				defer mockServer.Close()
				tt.args.serverURL = mockServer.URL

				SendMetrics(tt.args.metrics, tt.args.serverURL)
				assert.Greater(t, retries, 0)
			} else {
				mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "/update/", r.URL.Path)
					assert.Equal(t, "application/json", r.Header.Get("Content-type"))

					var metric metrics.Metrics
					err := json.NewDecoder(r.Body).Decode(&metric)
					assert.NoError(t, err)

					expectedMetrics := []string{"Alloc", "TotalAlloc"}
					assert.Contains(t, expectedMetrics, metric.GetName())
				}))
				defer mockServer.Close()
				tt.args.serverURL = mockServer.URL
			}

			SendMetrics(tt.args.metrics, tt.args.serverURL)
		})
	}
}

func TestSendBatchMetrics(t *testing.T) {
	type args struct {
		metrics   []MetricInterface
		serverURL string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Send multiple metrics",
			args: args{
				metrics: []MetricInterface{
					metrics.NewGauge("Alloc", 123.45),
					metrics.NewGauge("TotalAlloc", 678.90),
				},
				serverURL: "",
			},
		},
		{
			name: "Retry on failure",
			args: args{
				metrics: []MetricInterface{
					metrics.NewGauge("Alloc", 123.45),
				},
				serverURL: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Retry on failure" {
				var retries int
				mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if retries < retryCount {
						http.Error(w, "temporary error", http.StatusInternalServerError)
						retries++
					} else {
						w.WriteHeader(http.StatusOK)
					}
				}))
				defer mockServer.Close()
				tt.args.serverURL = mockServer.URL

				SendBatchMetrics(tt.args.metrics, tt.args.serverURL)
				assert.Greater(t, retries, 0)
			} else {
				mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "/updates/", r.URL.Path)
					assert.Equal(t, "application/json", r.Header.Get("Content-type"))

					var metric []*metrics.Metrics
					err := json.NewDecoder(r.Body).Decode(&metric)
					assert.NoError(t, err)
				}))
				defer mockServer.Close()
				tt.args.serverURL = mockServer.URL
				SendBatchMetrics(tt.args.metrics, tt.args.serverURL)
			}
		})
	}
}
