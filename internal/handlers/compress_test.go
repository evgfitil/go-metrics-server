package handlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"github.com/evgfitil/go-metrics-server.git/internal/metrics"
	"github.com/evgfitil/go-metrics-server.git/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testCompressRouter(s *storage.MemStorage) chi.Router {
	r := chi.NewRouter()
	r.Use(GzipMiddleware)
	r.Route("/value", func(r chi.Router) {
		r.Post("/", GetMetricsJSON(s))
	})
	r.Route("/update", func(r chi.Router) {
		r.Post("/", UpdateMetricsJSON(s))
	})
	return r
}

func TestGzipMiddleware(t *testing.T) {
	mockStorage := storage.NewMemStorage()
	ts := httptest.NewServer(testCompressRouter(mockStorage))
	validCounterValue := int64(100)
	validCounterMetric := metrics.Metrics{ID: "testCounter", MType: "counter", Delta: &validCounterValue}
	defer ts.Close()

	type want struct {
		statusCode int
		body       string
	}
	tests := []struct {
		name          string
		requestMethod string
		requestPath   string
		requestBody   metrics.Metrics
		headers       http.Header
		want          want
	}{
		{
			name:          "valid update request",
			requestMethod: http.MethodPost,
			requestPath:   "/update/",
			requestBody:   validCounterMetric,
			headers: http.Header{
				"Content-Encoding": []string{"gzip"},
				"Content-Type":     []string{"application/json"},
			},
			want: want{
				statusCode: http.StatusOK,
				body:       `{"id":"testCounter","type":"counter","delta":100}`,
			},
		},
		{
			name:          "valid get request",
			requestMethod: http.MethodPost,
			requestPath:   "/value/",
			requestBody:   metrics.Metrics{ID: "testCounter", MType: "counter"},
			headers: http.Header{
				"Accept-Encoding": []string{"gzip"},
				"Content-Type":    []string{"application/json"},
			},
			want: want{
				statusCode: http.StatusOK,
				body:       `{"id":"testCounter","type":"counter","delta":100}`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.requestBody)
			if tt.headers.Get("Content-Encoding") == "gzip" {
				var b bytes.Buffer
				gz := gzip.NewWriter(&b)
				_, err := gz.Write(jsonBody)
				require.NoError(t, err)
				err = gz.Close()
				require.NoError(t, err)
				jsonBody = b.Bytes()
			}

			req, err := http.NewRequest(tt.requestMethod, ts.URL+tt.requestPath, bytes.NewBuffer(jsonBody))
			require.NoError(t, err)
			for key, values := range tt.headers {
				for _, value := range values {
					req.Header.Add(key, value)
				}
			}

			resp, err := ts.Client().Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			var body []byte
			if resp.Header.Get("Content-Encoding") == "gzip" {
				gz, err := gzip.NewReader(resp.Body)
				require.NoError(t, err)
				defer gz.Close()
				body, err = io.ReadAll(gz)
				require.NoError(t, err)
			} else {
				body, err = io.ReadAll(resp.Body)
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)
			if tt.want.statusCode == http.StatusOK {
				assert.JSONEq(t, tt.want.body, string(body))
			}
		})
	}
}
