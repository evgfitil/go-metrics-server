package handlers

import (
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
	"github.com/evgfitil/go-metrics-server.git/internal/storage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetMetricsHandler(t *testing.T) {
	mockStorage := storage.NewMemStorage()
	mockMetric := metrics.Counter{Name: "testCounter", Value: 100}
	mockStorage.Update(mockMetric)

	type want struct {
		statusCode int
	}
	tests := []struct {
		name          string
		requestMethod string
		requestPath   string
		want          want
	}{
		{
			name:          "valid request",
			requestMethod: http.MethodGet,
			requestPath:   "/get/testCounter",
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:          "wrong requestMethod",
			requestMethod: http.MethodPost,
			requestPath:   "/get/testCounter",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:          "wrong requestPath",
			requestMethod: http.MethodGet,
			requestPath:   "/wrong",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:          "get non existing metric",
			requestMethod: http.MethodGet,
			requestPath:   "/get/nonExistingMetric",
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.requestMethod, tt.requestPath, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(GetMetricsHandler(mockStorage))
			h(w, request)
			res := w.Result()
			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			defer res.Body.Close()
		})
	}
}

func TestUpdateMetricsHandler(t *testing.T) {
	mockStorage := storage.NewMemStorage()
	type want struct {
		statusCode int
	}
	tests := []struct {
		name          string
		requestMethod string
		requestPath   string
		want          want
	}{
		{
			name:          "valid Counter update",
			requestMethod: http.MethodPost,
			requestPath:   "/update/counter/testCounter/123",
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:          "valid Gauge update",
			requestMethod: http.MethodPost,
			requestPath:   "/update/gauge/testGauge/123.12",
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:          "invalid request method",
			requestMethod: http.MethodGet,
			requestPath:   "/update/gauge/testGauge/123.12",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:          "invalid path",
			requestMethod: http.MethodPost,
			requestPath:   "/update/gauge/test/Gauge/123.12",
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:          "invalid metric type",
			requestMethod: http.MethodPost,
			requestPath:   "/update/histogram/testHistogram/123.12",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:          "invalid Counter value",
			requestMethod: http.MethodPost,
			requestPath:   "/update/counter/testCounter/123.12",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.requestMethod, tt.requestPath, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(UpdateMetricsHandler(mockStorage))
			h(w, request)
			res := w.Result()
			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			defer res.Body.Close()
		})
	}
}
