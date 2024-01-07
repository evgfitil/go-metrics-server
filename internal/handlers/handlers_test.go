package handlers

import (
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
	"github.com/evgfitil/go-metrics-server.git/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testMetricsRouter(s *storage.MemStorage) chi.Router {
	r := chi.NewRouter()
	r.Get("/value/{type}/{name}", GetMetrics(s))
	r.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", UpdateMetrics(s))
	})
	return r
}

func TestGetMetricsHandler(t *testing.T) {
	mockStorage := storage.NewMemStorage()
	mockMetric := metrics.Counter{Name: "testCounter", Value: 100}
	mockStorage.Update(mockMetric)

	ts := httptest.NewServer(testMetricsRouter(mockStorage))
	defer ts.Close()

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
			requestPath:   "/value/counter/testCounter",
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name:          "wrong requestMethod",
			requestMethod: http.MethodPost,
			requestPath:   "/value/counter/testCounter",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
		},
		{
			name:          "wrong requestPath",
			requestMethod: http.MethodGet,
			requestPath:   "/wrong",
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
		{
			name:          "get non existing metric",
			requestMethod: http.MethodGet,
			requestPath:   "/value/counter/nonExistingMetric",
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.requestMethod, ts.URL+tt.requestPath, nil)
			require.NoError(t, err)

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
		})
	}
}

func TestUpdateMetricsHandler(t *testing.T) {
	mockStorage := storage.NewMemStorage()

	ts := httptest.NewServer(testMetricsRouter(mockStorage))
	defer ts.Close()

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
				statusCode: http.StatusMethodNotAllowed,
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
			req, err := http.NewRequest(tt.requestMethod, ts.URL+tt.requestPath, nil)
			require.NoError(t, err)

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
		})
	}
}
